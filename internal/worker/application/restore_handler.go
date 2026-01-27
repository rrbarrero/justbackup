package application

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/crypto"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

// HandleRestoreLocalTask orchestrates the restoration to a local CLI consumer.
func HandleRestoreLocalTask(ctx context.Context, task workerDto.WorkerTask, redisClient *redis.Client, resultQueue string) {
	log.Printf("Restoring local for task %s (encrypted: %v)", task.TaskID, task.Encrypted)

	conn, err := establishConnection(task.RestoreAddr)
	if err != nil {
		reportRestoreFailure(ctx, redisClient, resultQueue, task, "Failed to connect to CLI", err)
		return
	}
	defer conn.Close()

	if err := authenticateConnection(conn, task.RestoreToken); err != nil {
		log.Printf("Failed to send token: %v", err)
		return
	}

	if err := streamRestoreData(conn, task); err != nil {
		reportRestoreFailure(ctx, redisClient, resultQueue, task, "Data streaming failed", err)
		return
	}

	reportRestoreSuccess(ctx, redisClient, resultQueue, task, "Restore streaming completed successfully")
}

// HandleRestoreRemoteTask orchestrates the restoration to a remote host via rsync.
func HandleRestoreRemoteTask(ctx context.Context, task workerDto.WorkerTask, redisClient *redis.Client, resultQueue string) {
	log.Printf("Restoring remote for task %s to %s@%s:%s", task.TaskID, task.TargetUser, task.TargetHost, task.TargetPath)

	sourcePath, cleanup, err := prepareRestoreSource(task)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		reportRestoreFailure(ctx, redisClient, resultQueue, task, "Failed to prepare source", err)
		return
	}

	if err := executeRemoteRestore(task, sourcePath); err != nil {
		reportRestoreFailure(ctx, redisClient, resultQueue, task, "Remote restore execution failed", err)
		return
	}

	reportRestoreSuccess(ctx, redisClient, resultQueue, task, "Remote restore completed successfully")
}

// --- Local Restore Helpers ---

func establishConnection(addr string) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, 10*time.Second)
}

func authenticateConnection(conn net.Conn, token string) error {
	_, err := conn.Write([]byte(token))
	return err
}

func streamRestoreData(w io.Writer, task workerDto.WorkerTask) error {
	if task.Encrypted {
		return streamEncryptedData(w, task)
	}
	return streamPlainData(w, task)
}

func streamEncryptedData(w io.Writer, task workerDto.WorkerTask) error {
	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		return err
	}

	key, err := crypto.DeriveKey(cfg.EncryptionKey, task.BackupID)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	tempTarPath := task.Path + ".tmp.tar.gz"
	if err := crypto.DecryptFile(task.Path, tempTarPath, key); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}
	defer os.Remove(tempTarPath)

	file, err := os.Open(tempTarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	return err
}

func streamPlainData(w io.Writer, task workerDto.WorkerTask) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Walk the directory and write to tar structure
	return filepath.Walk(task.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Relativize path to the backup root
		relPath, err := filepath.Rel(filepath.Dir(task.Path), path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tw, file)
			return err
		}
		return nil
	})
}

// --- Remote Restore Helpers ---

func prepareRestoreSource(task workerDto.WorkerTask) (string, func(), error) {
	if !task.Encrypted {
		return task.Path, nil, nil
	}

	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		return "", nil, err
	}

	key, err := crypto.DeriveKey(cfg.EncryptionKey, task.BackupID)
	if err != nil {
		return "", nil, fmt.Errorf("key derivation failed: %w", err)
	}

	// 1. Decrypt to temp file
	tempTarPath := task.Path + ".tmp.tar.gz"
	if err := crypto.DecryptFile(task.Path, tempTarPath, key); err != nil {
		return "", nil, fmt.Errorf("decryption failed: %w", err)
	}

	// 2. Extract to temp dir
	tempDir, err := os.MkdirTemp("", "restore-*")
	if err != nil {
		os.Remove(tempTarPath)
		return "", nil, fmt.Errorf("temp dir creation failed: %w", err)
	}

	if err := crypto.DecompressTarGz(tempTarPath, tempDir); err != nil {
		os.Remove(tempTarPath)
		os.RemoveAll(tempDir)
		return "", nil, fmt.Errorf("decompression failed: %w", err)
	}

	// Cleanup closure
	cleanup := func() {
		os.Remove(tempTarPath)
		os.RemoveAll(tempDir)
	}

	// Return path with trailing slash for rsync content
	return tempDir + "/", cleanup, nil
}

func executeRemoteRestore(task workerDto.WorkerTask, sourcePath string) error {
	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		return err
	}

	destination := fmt.Sprintf("%s@%s:%s", task.TargetUser, task.TargetHost, task.TargetPath)
	sshOpts := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p %d", cfg.SSHKeyPath, task.TargetPort)

	// --no-t prevents permission errors on setting times for root-owned dirs like /tmp
	args := []string{"-avz", "--no-t", "-e", sshOpts, sourcePath, destination}
	cmd := exec.Command("rsync", args...)

	log.Printf("Executing Remote Restore: %s", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync failed: %w, output: %s", err, string(output))
	}

	return nil
}

// --- Reporting Helpers ---

func reportRestoreFailure(ctx context.Context, client *redis.Client, queue string, task workerDto.WorkerTask, msg string, err error) {
	fullMsg := fmt.Sprintf("%s: %v", msg, err)
	log.Printf("ERROR: %s", fullMsg)
	PublishResult(ctx, client, queue, workerDto.WorkerResult{
		Type:    taskTypeForRestore(task),
		TaskID:  task.TaskID,
		JobID:   task.JobID,
		Status:  "failed",
		Message: fullMsg,
	})
}

func reportRestoreSuccess(ctx context.Context, client *redis.Client, queue string, task workerDto.WorkerTask, msg string) {
	log.Printf("SUCCESS: %s", msg)
	PublishResult(ctx, client, queue, workerDto.WorkerResult{
		Type:    taskTypeForRestore(task),
		TaskID:  task.TaskID,
		JobID:   task.JobID,
		Status:  "completed",
		Message: msg,
	})
}

func taskTypeForRestore(task workerDto.WorkerTask) workerDto.TaskType {
	if task.TargetHost != "" {
		return workerDto.TaskTypeRestoreRemote
	}
	return workerDto.TaskTypeRestoreLocal
}

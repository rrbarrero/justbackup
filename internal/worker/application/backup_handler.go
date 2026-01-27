package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/crypto"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

// HandleBackupTask orchestrates the backup workflow.
// It follows a linear flow: Setup -> PreHooks -> Sync -> Encrypt -> PostHooks -> Report.
func HandleBackupTask(ctx context.Context, task workerDto.WorkerTask, redisClient *redis.Client, resultQueue string) {
	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		reportError(ctx, redisClient, resultQueue, task, "CRITICAL: Failed to load worker config", err)
		return
	}

	// 1. Prepare Destination
	finalDest, err := prepareBackupDestination(task, cfg)
	if err != nil {
		reportError(ctx, redisClient, resultQueue, task, "Failed to prepare destination", err)
		return
	}

	// 2. Setup Workspace (if needed)
	taskPath, sessionTempDir, cleanupWorkspace, err := setupEphemeralWorkspace(task)
	if cleanupWorkspace != nil {
		defer cleanupWorkspace()
	}
	if err != nil {
		reportError(ctx, redisClient, resultQueue, task, "Failed to setup ephemeral workspace", err)
		return
	}

	// 3. Pre-Backup Hooks
	if err := executeHooks(task.Hooks, "pre", finalDest, sessionTempDir); err != nil {
		reportError(ctx, redisClient, resultQueue, task, "Pre-backup hooks failed", err)
		return
	}

	// 4. Execute Backup (Rsync)
	if err := executeRsyncOperation(task, cfg, finalDest, taskPath); err != nil {
		reportError(ctx, redisClient, resultQueue, task, "Rsync execution failed", err)
		return
	}

	// 5. Post-Processing (Encryption & Compression)
	finalArtifactPath := finalDest
	if task.Encrypted {
		finalArtifactPath, err = performEncryptionWorkflow(task, finalDest, cfg)
		if err != nil {
			reportError(ctx, redisClient, resultQueue, task, "Encryption workflow failed", err)
			return
		}
	}

	// 6. Post-Backup Hooks
	if err := executeHooks(task.Hooks, "post", finalDest, sessionTempDir); err != nil {
		reportError(ctx, redisClient, resultQueue, task, "Post-backup hooks failed", err)
		return
	}

	// 7. Calculate Size & Report Success
	size := calculateArtifactSize(finalArtifactPath)

	destRelPath := NormalizePath(task.Destination, cfg.HostBackupRoot, task.HostPath)
	if task.Encrypted {
		destRelPath += ".tar.gz.enc"
	}

	PublishResult(ctx, redisClient, resultQueue, workerDto.WorkerResult{
		Type:    workerDto.TaskTypeBackup,
		TaskID:  task.TaskID,
		JobID:   task.JobID,
		Status:  "completed",
		Message: "Backup completed successfully",
		Data: map[string]string{
			"path": destRelPath,
			"size": size,
		},
	})
}

// prepareBackupDestination calculates the target directory and ensures it exists.
func prepareBackupDestination(task workerDto.WorkerTask, cfg *config.WorkerConfig) (string, error) {
	baseDest := NormalizePath(task.Destination, cfg.ContainerBackupRoot, task.HostPath)
	finalDest := baseDest

	if task.Incremental {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		finalDest = fmt.Sprintf("%s/%s", baseDest, timestamp)
	}

	if err := os.MkdirAll(finalDest, 0755); err != nil {
		return "", fmt.Errorf("mkdir %s failed: %w", finalDest, err)
	}

	return finalDest, nil
}

// setupEphemeralWorkspace handles {{SESSION_TEMP_DIR}} logic.
func setupEphemeralWorkspace(task workerDto.WorkerTask) (string, string, func(), error) {
	if task.Path != "{{SESSION_TEMP_DIR}}" {
		// Use original path, no cleanup, no session dir
		// We return empty string for sourcePath override so rsync uses remote syntax
		return "", "", nil, nil
	}

	tempDir, err := os.MkdirTemp("", "start-backup-session-*")
	if err != nil {
		return "", "", nil, err
	}

	log.Printf("Created ephemeral workspace: %s", tempDir)

	cleanup := func() {
		log.Printf("Cleaning up session temp dir: %s", tempDir)
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("WARNING: Failed to cleanup session temp dir %s: %v", tempDir, err)
		}
	}

	// Here we MUST return the tempDir as the sourcePath because we want rsync
	// to copy FROM this local temp directory, not from the remote host.
	return tempDir, tempDir, cleanup, nil
}

// executeRsyncOperation wraps the low-level rsync call logic.
func executeRsyncOperation(task workerDto.WorkerTask, cfg *config.WorkerConfig, finalDest string, sourcePath string) error {
	sshKeyPath := cfg.SSHKeyPath

	// Determine Link Dest for incremental
	var useLinkDest bool
	baseDest := NormalizePath(task.Destination, cfg.ContainerBackupRoot, task.HostPath)
	linkDest := fmt.Sprintf("%s/latest", baseDest)

	if task.Incremental {
		if _, err := os.Lstat(linkDest); err == nil {
			useLinkDest = true
		}
	}

	// Prepare Source string
	rsyncSource := fmt.Sprintf("%s@%s:%s", task.User, task.Host, task.Path)

	if sourcePath != "" && strings.HasPrefix(sourcePath, "/") {
		// Local source logic
		rsyncSource = sourcePath
		if !strings.HasSuffix(rsyncSource, "/") {
			rsyncSource += "/"
		}
	}

	// Prepare Excludes
	var excludeFlags []string
	for _, exception := range task.Excludes {
		excludeFlags = append(excludeFlags, fmt.Sprintf("--exclude=%s", exception))
	}

	args := BuildRsyncArgs(sshKeyPath, task, useLinkDest, excludeFlags, rsyncSource, finalDest)
	cmd := exec.Command("rsync", args...)

	log.Printf("Executing: %s", cmd.String())

	// Capture output for error reporting
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		if _, err := os.Stdout.Write(output); err != nil {
			log.Printf("Failed to write rsync output to stdout: %v", err)
		}
	}

	if err != nil {
		return handleRsyncError(err, output)
	}

	// Update Incremental Link
	if task.Incremental {
		updateLatestSymlink(linkDest, finalDest)
	}

	return nil
}

// performEncryptionWorkflow handles compression, encryption, and cleanup of raw files.
func performEncryptionWorkflow(task workerDto.WorkerTask, sourceDir string, cfg *config.WorkerConfig) (string, error) {
	log.Printf("Encryption requested for backup %s", task.TaskID)

	if cfg.EncryptionKey == "" {
		return "", fmt.Errorf("ENCRYPTION_KEY not set in worker configuration")
	}

	key, err := crypto.DeriveKey(cfg.EncryptionKey, task.TaskID)
	if err != nil {
		return "", fmt.Errorf("key derivation failed: %w", err)
	}

	tarPath := sourceDir + ".tar.gz"
	if err := crypto.CompressDirectory(sourceDir, tarPath); err != nil {
		return "", fmt.Errorf("compression failed: %w", err)
	}

	encPath := tarPath + ".enc"
	if err := crypto.EncryptFile(tarPath, encPath, key); err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	log.Printf("Backup encrypted successfully: %s", encPath)

	// Clean up intermediate files
	if err := os.RemoveAll(sourceDir); err != nil {
		log.Printf("WARNING: Failed to remove source dir %s: %v", sourceDir, err)
	}
	if err := os.Remove(tarPath); err != nil {
		log.Printf("WARNING: Failed to remove tarball %s: %v", tarPath, err)
	}

	// Update symlink if incremental
	if task.Incremental {
		baseDest := NormalizePath(task.Destination, cfg.ContainerBackupRoot, task.HostPath)
		linkDestEnc := path.Join(baseDest, "latest.tar.gz.enc")
		updateLatestSymlink(linkDestEnc, encPath)
	}

	return encPath, nil
}

// Helper: updates the 'latest' symlink for incremental backups
func updateLatestSymlink(linkPath, targetPath string) {
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		log.Printf("WARNING: Failed to remove old symlink %s: %v", linkPath, err)
	}
	if err := os.Symlink(filepath.Base(targetPath), linkPath); err != nil {
		log.Printf("WARNING: Failed to update symlink %s -> %s: %v", linkPath, targetPath, err)
	}
}

// calculateArtifactSize uses 'du' to measure the final output.
func calculateArtifactSize(path string) string {
	cmd := exec.Command("du", "-sk", path)
	// For file use -k, for dir -sk works for both usually but let's be safe
	// In linux du -k works for file too.
	// Original code had check. du -sk is safer for directory summary.

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to calculate size for %s: %v", path, err)
		return "0KB"
	}

	fields := strings.Fields(out.String())
	if len(fields) > 0 {
		return fields[0] + "KB"
	}
	return "0KB"
}

// handleRsyncError provides semantic error messages based on exit codes.
func handleRsyncError(err error, output []byte) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code := exitErr.ExitCode()
		// Code 24: Vanished files (warning only)
		if code == 24 {
			log.Printf("WARNING: Rsync finished with exit code 24 (Vanished source files). Marking as success.")
			return nil
		}

		// Parse output for better error message
		outStr := strings.TrimSpace(string(output))
		lines := strings.Split(outStr, "\n")
		lastLine := "Unknown rsync error"
		if len(lines) > 0 {
			lastLine = lines[len(lines)-1]
		}

		return fmt.Errorf("rsync failed (code %d): %s", code, lastLine)
	}
	return err
}

// reportError is a helper to log and publish failure results consistently.
func reportError(ctx context.Context, redisClient *redis.Client, queue string, task workerDto.WorkerTask, msg string, err error) {
	log.Printf("%s: %v", msg, err)
	PublishResult(ctx, redisClient, queue, workerDto.WorkerResult{
		Type:    workerDto.TaskTypeBackup,
		TaskID:  task.TaskID,
		JobID:   task.JobID,
		Status:  "failed",
		Message: fmt.Sprintf("%s: %v", msg, err),
	})
}

// BuildRsyncArgs (Pure function, extracted for testing)
func BuildRsyncArgs(sshKeyPath string, task workerDto.WorkerTask, useLinkDest bool, excludeFlags []string, source, finalDest string) []string {
	sshOpts := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -p %d", sshKeyPath, task.Port)

	args := []string{"-az", "--no-owner", "--no-group", "--numeric-ids", "-e", sshOpts}
	args = append(args, excludeFlags...)

	if useLinkDest {
		args = append(args, "--link-dest=../latest")
	}

	args = append(args, source, finalDest)
	return args
}

// Hook Execution Logic
func executeHooks(hooks []workerDto.HookTask, phase string, backupDest string, sessionTempDir string) error {
	for _, hook := range hooks {
		if !hook.Enabled || hook.Phase != phase {
			continue
		}
		if err := executeHook(hook, backupDest, sessionTempDir); err != nil {
			return err
		}
	}
	return nil
}

func executeHook(hook workerDto.HookTask, backupDest string, sessionTempDir string) error {
	pluginDir := "/app/plugins"
	scriptPath := filepath.Join(pluginDir, hook.Name+".sh")

	if err := ValidateHookPath(pluginDir, hook.Name); err != nil {
		return err
	}

	cleanScriptPath := filepath.Clean(scriptPath)
	if _, err := os.Stat(cleanScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin script not found: %s", cleanScriptPath)
	}

	log.Printf("Executing hook [%s]: %s", hook.Phase, hook.Name)
	cmd := exec.Command("bash", cleanScriptPath)

	env := os.Environ()
	env = append(env, fmt.Sprintf("BACKUP_DEST=%s", backupDest))
	env = append(env, fmt.Sprintf("HOOK_PHASE=%s", hook.Phase))
	if sessionTempDir != "" {
		env = append(env, fmt.Sprintf("SESSION_TEMP_DIR=%s", sessionTempDir))
	}
	for k, v := range hook.Params {
		key := fmt.Sprintf("HOOK_PARAM_%s", strings.ToUpper(strings.ReplaceAll(k, "-", "_")))
		env = append(env, fmt.Sprintf("%s=%s", key, v))
	}
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hook %s failed: %w (output: %s)", hook.Name, err, string(output))
	}

	log.Printf("Hook [%s] completed. Output: %s", hook.Name, string(output))
	return nil
}

func ValidateHookPath(pluginDir string, hookName string) error {
	scriptPath := filepath.Join(pluginDir, hookName+".sh")
	cleanScriptPath := filepath.Clean(scriptPath)
	if !strings.HasPrefix(cleanScriptPath, pluginDir) {
		return fmt.Errorf("security violation: invalid hook path %s", hookName)
	}
	return nil
}

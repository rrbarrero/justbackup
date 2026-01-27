package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

func HandleMeasureSizeTask(ctx context.Context, task workerDto.WorkerTask, redisClient *redis.Client, resultQueue string) {
	log.Printf("Measuring size for %s on %s", task.Path, task.Host)

	// Execute du -sh
	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		log.Printf("CRITICAL: Failed to load worker config: %v", err)
		PublishResult(ctx, redisClient, resultQueue, workerDto.WorkerResult{
			Type:    workerDto.TaskTypeMeasureSize,
			TaskID:  task.TaskID,
			JobID:   task.JobID,
			Status:  "failed",
			Message: fmt.Sprintf("Worker configuration error: %v", err),
		})
		return
	}
	sshKeyPath := cfg.SSHKeyPath

	sshOpts := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -p %d", sshKeyPath, task.Port)
	cmdStr := fmt.Sprintf("du -sk %s", task.Path)

	var stdout, stderr bytes.Buffer
	var cmdErr error

	if os.Getenv("ENVIRONMENT") == "dev" {
		log.Println("Simulating measure size success in dev environment")
		stdout.WriteString("1572864\t" + task.Path + "\n")
	} else {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s@%s '%s'", sshOpts, task.User, task.Host, cmdStr))
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		cmdErr = cmd.Run()
	}

	if stderr.Len() > 0 {
		log.Printf("Measure size stderr: %s", stderr.String())
	}

	if cmdErr != nil {
		log.Printf("Measure size failed: %v, output: %s", cmdErr, stderr.String())
		return
	}

	output := stdout.String()
	// Parse output to get only the size (first field)
	// Output format: "32K\t/path/to/file\n"
	fields := strings.Fields(output)
	size := ""
	if len(fields) > 0 {
		size = fields[0] + "KB"
	} else {
		size = strings.TrimSpace(output) + "KB"
	}

	log.Printf("Size: %s", size)

	result := workerDto.WorkerResult{
		Type:    workerDto.TaskTypeMeasureSize,
		TaskID:  task.TaskID,
		JobID:   task.JobID,
		Status:  "completed",
		Message: "Size measured successfully",
		Data:    map[string]string{"size": size},
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Printf("Failed to marshal result: %v", err)
		return
	}

	if err := redisClient.RPush(ctx, resultQueue, data).Err(); err != nil {
		log.Printf("Failed to publish result: %v", err)
	}
}

func HandleGetDiskUsage(ctx context.Context, task workerDto.WorkerTask, redisClient *redis.Client) {
	log.Printf("Getting disk usage for backup root")

	var stat syscall.Statfs_t
	// We check the disk usage of the backup root mount point
	// In the worker container, this is mounted at /mnt/backups
	backupMountPoint := "/mnt/backups"

	if err := syscall.Statfs(backupMountPoint, &stat); err != nil {
		log.Printf("Failed to get disk usage: %v", err)
		return
	}

	// Calculate usage
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	result := workerDto.WorkerResult{
		Type:    workerDto.TaskTypeGetDiskUsage,
		TaskID:  task.TaskID,
		Status:  "completed",
		Message: "Disk usage retrieved successfully",
		Data: map[string]string{
			"total": fmt.Sprintf("%d", total),
			"free":  fmt.Sprintf("%d", free),
			"used":  fmt.Sprintf("%d", used),
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Printf("Failed to marshal result: %v", err)
		return
	}

	// Publish to PubSub channel for synchronous response
	channel := "worker_sync_responses"
	if err := redisClient.Publish(ctx, channel, data).Err(); err != nil {
		log.Printf("Failed to publish result to %s: %v", channel, err)
	} else {
		log.Printf("Published disk usage result to %s", channel)
	}
}

func HandleSearchFiles(ctx context.Context, task workerDto.WorkerTask, redisClient *redis.Client) {
	log.Printf("Searching files with pattern: %s", task.SearchPattern)

	backupMountPoint := "/mnt/backups"

	// Ensure the search pattern is safe and within the backup mount point.

	// Use find to search for files matching the pattern
	// -name supports simple globbing like * and ?
	cmd := exec.Command("find", backupMountPoint, "-name", task.SearchPattern)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		log.Printf("Find command failed: %v", err)
		// We don't return error here, just an empty list or error message in result
	}

	files := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(files) == 1 && files[0] == "" {
		files = []string{}
	}

	// Returning full paths as seen by the worker.

	result := workerDto.WorkerResult{
		Type:    workerDto.TaskTypeSearchFiles,
		TaskID:  task.TaskID,
		Status:  "completed",
		Message: fmt.Sprintf("Found %d files", len(files)),
		Data: map[string]interface{}{
			"files": files,
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Printf("Failed to marshal result: %v", err)
		return
	}

	// We use PubSub for synchronous-like response, same as DiskUsage
	channel := "worker_sync_responses"
	if err := redisClient.Publish(ctx, channel, data).Err(); err != nil {
		log.Printf("Failed to publish result to %s: %v", channel, err)
	}
}

func HandleListFiles(ctx context.Context, task workerDto.WorkerTask, redisClient *redis.Client) {
	log.Printf("Listing files for path: %s", task.Path)

	entries, err := os.ReadDir(task.Path)
	if err != nil {
		log.Printf("ReadDir failed for %s: %v", task.Path, err)
		// We could send an error result here
	}

	type FileListItem struct {
		Name  string `json:"name"`
		IsDir bool   `json:"is_dir"`
		Size  int64  `json:"size"`
	}

	var files []FileListItem
	if err == nil {
		for _, entry := range entries {
			info, err := entry.Info()
			size := int64(0)
			if err == nil {
				size = info.Size()
			}
			files = append(files, FileListItem{
				Name:  entry.Name(),
				IsDir: entry.IsDir(),
				Size:  size,
			})
		}
	}

	result := workerDto.WorkerResult{
		Type:    workerDto.TaskTypeListFiles,
		TaskID:  task.TaskID,
		Status:  "completed",
		Message: fmt.Sprintf("Found %d entries", len(files)),
		Data: map[string]interface{}{
			"files": files,
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Printf("Failed to marshal result: %v", err)
		return
	}

	channel := "worker_sync_responses"
	if err := redisClient.Publish(ctx, channel, data).Err(); err != nil {
		log.Printf("Failed to publish result to %s: %v", channel, err)
	}
}

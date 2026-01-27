package application

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"strings"

	"github.com/redis/go-redis/v9"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

type TaskHandler interface {
	Handle(ctx context.Context, task workerDto.WorkerTask, redisClient *redis.Client, resultQueue string)
}

func PublishResult(ctx context.Context, redisClient *redis.Client, resultQueue string, result workerDto.WorkerResult) {
	data, err := json.Marshal(result)
	if err != nil {
		log.Printf("Failed to marshal result: %v", err)
		return
	}
	if err := redisClient.RPush(ctx, resultQueue, data).Err(); err != nil {
		log.Printf("Failed to publish result: %v", err)
	}
}

func NormalizePath(dest string, backupRoot string, hostPath string) string {
	dest = strings.TrimSuffix(dest, "/")
	dest = strings.TrimPrefix(dest, "/")
	root := strings.TrimSuffix(backupRoot, "/")

	if hostPath != "" {
		hostPath = strings.TrimSuffix(hostPath, "/")
		hostPath = strings.TrimPrefix(hostPath, "/")
		return root + "/" + hostPath + "/" + dest
	}

	return root + "/" + dest
}

// SortStrings sorts a slice of strings in ascending order using the standard library.
func SortStrings(s []string) {
	sort.Strings(s)
}

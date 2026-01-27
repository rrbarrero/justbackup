package http

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

type SystemHandler struct {
	redisClient *redis.Client
}

func NewSystemHandler(redisClient *redis.Client) *SystemHandler {
	return &SystemHandler{
		redisClient: redisClient,
	}
}

func (h *SystemHandler) RegisterRoutes(mux *http.ServeMux, middleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /system/disk-usage", middleware(h.GetDiskUsage))
}

// @Summary Get disk usage
// @Description Get disk usage statistics for the backup partition
// @Tags system
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]string
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /system/disk-usage [get]
func (h *SystemHandler) GetDiskUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := uuid.New().String()

	// Subscribe to response channel
	pubsub := h.redisClient.Subscribe(ctx, "worker_sync_responses")
	defer pubsub.Close()

	// Wait for subscription to be confirmed (optional but recommended)
	// ReceiveTimeout fits better to avoid blocking forever if redis is down
	_, err := pubsub.ReceiveTimeout(ctx, 1*time.Second)
	if err != nil {
		http.Error(w, "Failed to subscribe to redis", http.StatusInternalServerError)
		return
	}

	task := workerDto.WorkerTask{
		Type:   workerDto.TaskTypeGetDiskUsage,
		TaskID: taskID,
	}

	payload, err := json.Marshal(task)
	if err != nil {
		http.Error(w, "Failed to marshal task", http.StatusInternalServerError)
		return
	}

	if err := h.redisClient.RPush(ctx, "backup_tasks", payload).Err(); err != nil {
		http.Error(w, "Failed to push task", http.StatusInternalServerError)
		return
	}

	// Wait for response
	timeout := time.After(5 * time.Second)
	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			var result workerDto.WorkerResult
			if err := json.Unmarshal([]byte(msg.Payload), &result); err != nil {
				log.Printf("Failed to unmarshal result: %v", err)
				continue
			}

			if result.TaskID == taskID {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(result.Data)
				return
			}
		case <-timeout:
			http.Error(w, "Timeout waiting for worker response", http.StatusGatewayTimeout)
			return
		case <-ctx.Done():
			return
		}
	}
}

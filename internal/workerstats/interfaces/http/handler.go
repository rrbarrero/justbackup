package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rrbarrero/justbackup/internal/workerstats/application"
)

// WorkerStatsHandler handles stats-related HTTP requests
type WorkerStatsHandler struct {
	service *application.WorkerStatsService
}

// NewWorkerStatsHandler creates a new stats handler
func NewWorkerStatsHandler(service *application.WorkerStatsService) *WorkerStatsHandler {
	return &WorkerStatsHandler{
		service: service,
	}
}

// PostStats receives a new report from a worker
// @Summary Post worker statistics
// @Description Receive CPU, memory, and disk usage from a worker
// @Tags workerstats
// @Accept json
// @Produce json
// @Param stats body application.WorkerStatsReportDTO true "Worker Stats Report"
// @Success 202 "Accepted"
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string
// @Router /workers/stats [post]
func (h *WorkerStatsHandler) PostStats(w http.ResponseWriter, r *http.Request) {
	var dto application.WorkerStatsReportDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if dto.WorkerID == "" {
		http.Error(w, "worker_id is required", http.StatusBadRequest)
		return
	}

	if err := h.service.PostStats(r.Context(), dto); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetStats returns all worker stats
// @Summary Get all worker statistics
// @Description Get rolling window stats for all workers
// @Tags workerstats
// @Produce json
// @Success 200 {array} entities.WorkerStatsWindow
// @Failure 500 {string} string
// @Router /workers/stats [get]
func (h *WorkerStatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// RegisterRoutes registers the stats routes
func (h *WorkerStatsHandler) RegisterRoutes(mux *http.ServeMux, protected func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("POST /workers/stats", h.PostStats)
	mux.HandleFunc("GET /workers/stats", protected(h.GetStats))
}

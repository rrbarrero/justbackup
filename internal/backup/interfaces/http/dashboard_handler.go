package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rrbarrero/justbackup/internal/backup/application"
)

type DashboardHandler struct {
	service *application.DashboardService
}

func NewDashboardHandler(service *application.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		service: service,
	}
}

func (h *DashboardHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/dashboard/stats", h.GetStats)
}

// @Summary Get dashboard statistics
// @Description Get statistics for the dashboard
// @Tags dashboard
// @Accept  json
// @Produce  json
// @Success 200 {object} application.DashboardStats
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /dashboard/stats [get]
func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

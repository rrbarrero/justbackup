package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/memory"
)

type TestHandler struct {
	errorRepo *memory.BackupErrorRepositoryMemory
}

func NewTestHandler(errorRepo *memory.BackupErrorRepositoryMemory) *TestHandler {
	return &TestHandler{
		errorRepo: errorRepo,
	}
}

func (h *TestHandler) RegisterRoutes(mux *http.ServeMux, middleware func(http.HandlerFunc) http.HandlerFunc) {
	// Only register test routes in development/test mode
	mux.HandleFunc("POST /test/backup-errors/seed", middleware(h.SeedBackupErrors))
}

type SeedErrorRequest struct {
	BackupID     string `json:"backup_id"`
	JobID        string `json:"job_id"`
	ErrorMessage string `json:"error_message"`
	OccurredAt   string `json:"occurred_at"`
}

// SeedBackupErrors is a test-only endpoint to populate error data
// @Summary Seed test backup errors
// @Description Add predefined backup errors for E2E testing (test mode only)
// @Tags test
// @Accept  json
// @Produce  json
// @Param   errors body []SeedErrorRequest true "List of errors to seed"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/test/backup-errors/seed [post]
func (h *TestHandler) SeedBackupErrors(w http.ResponseWriter, r *http.Request) {
	var requests []SeedErrorRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errors := make([]*entities.BackupError, 0, len(requests))
	for _, req := range requests {
		backupID, err := valueobjects.NewBackupIDFromString(req.BackupID)
		if err != nil {
			http.Error(w, "Invalid backup ID: "+err.Error(), http.StatusBadRequest)
			return
		}

		occurredAt := time.Now()
		if req.OccurredAt != "" {
			parsed, err := time.Parse(time.RFC3339, req.OccurredAt)
			if err != nil {
				http.Error(w, "Invalid occurred_at format: "+err.Error(), http.StatusBadRequest)
				return
			}
			occurredAt = parsed
		}

		backupError := &entities.BackupError{
			ID:           "", // Will be generated if needed
			JobID:        req.JobID,
			BackupID:     backupID,
			OccurredAt:   occurredAt,
			ErrorMessage: req.ErrorMessage,
		}
		errors = append(errors, backupError)
	}

	h.errorRepo.SeedTestErrors(errors)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Seeded successfully",
		"count":   string(rune(len(errors))),
	})
}

package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/rrbarrero/justbackup/internal/backup/application"
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
)

type BackupHandler struct {
	lifecycleService *application.BackupLifecycleService
	queryService     *application.BackupQueryService
	searchService    *application.BackupSearchService
	restoreService   *application.BackupRestoreService
	taskService      *application.BackupTaskService
	hookService      *application.BackupHookService
}

func NewBackupHandler(
	lifecycleService *application.BackupLifecycleService,
	queryService *application.BackupQueryService,
	searchService *application.BackupSearchService,
	restoreService *application.BackupRestoreService,
	taskService *application.BackupTaskService,
	hookService *application.BackupHookService,
) *BackupHandler {
	return &BackupHandler{
		lifecycleService: lifecycleService,
		queryService:     queryService,
		searchService:    searchService,
		restoreService:   restoreService,
		taskService:      taskService,
		hookService:      hookService,
	}
}

func (h *BackupHandler) RegisterRoutes(mux *http.ServeMux, middleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /backups", middleware(h.List))
	mux.HandleFunc("GET /backups/{id}", middleware(h.GetByID))
	mux.HandleFunc("GET /backups/{id}/errors", middleware(h.GetErrors))
	mux.HandleFunc("DELETE /backups/{id}/errors", middleware(h.DeleteErrors))
	mux.HandleFunc("POST /backups", middleware(h.Create))
	mux.HandleFunc("PUT /backups/{id}", middleware(h.Update))
	mux.HandleFunc("DELETE /backups/{id}", middleware(h.Delete))
	mux.HandleFunc("POST /backups/{id}/run", middleware(h.Run))
	mux.HandleFunc("POST /hosts/{id}/run", middleware(h.RunHostBackups))
	mux.HandleFunc("GET /files/search", middleware(h.SearchFiles))
	mux.HandleFunc("POST /backups/{id}/restore", middleware(h.Restore))
	mux.HandleFunc("GET /backups/{id}/files", middleware(h.ListFiles))

	// Backup Hook Routes
	mux.HandleFunc("POST /backups/{id}/hooks", middleware(h.CreateHook))
	mux.HandleFunc("PUT /backups/{id}/hooks/{hookID}", middleware(h.UpdateHook))
	mux.HandleFunc("DELETE /backups/{id}/hooks/{hookID}", middleware(h.DeleteHook))
}

// @Summary Run a backup
// @Description Trigger a backup execution immediately
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Backup ID"
// @Success 200 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /backups/{id}/run [post]
func (h *BackupHandler) Run(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	taskID, err := h.lifecycleService.RunBackup(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"task_id": taskID}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// @Summary Run all backups for a host
// @Description Trigger all backup executions for a specific host immediately
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Host ID"
// @Success 200 {object} map[string][]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /hosts/{id}/run [post]
func (h *BackupHandler) RunHostBackups(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	taskIDs, err := h.lifecycleService.RunHostBackups(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string][]string{"task_ids": taskIDs}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// @Summary Delete a backup
// @Description Delete a backup configuration
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Backup ID"
// @Success 204 "No Content"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /backups/{id} [delete]
func (h *BackupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.lifecycleService.DeleteBackup(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Create a backup
// @Description Create a new backup configuration
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   backup     body    dto.CreateBackupRequest     true  "Backup Configuration"
// @Success 201 {object} dto.BackupResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /backups [post]
func (h *BackupHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.lifecycleService.CreateBackup(r.Context(), req)
	if err != nil {
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "above maximum (6)") {
			errorMessage = fmt.Sprintf("Invalid Cron Expression: Day of Week (field 5 or 6) must be 0-6. You used '7', please use '0' for Sunday. Original error: %v", err)
			http.Error(w, errorMessage, http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}

}

// @Summary Update a backup
// @Description Update an existing backup configuration
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Backup ID"
// @Param   backup     body    dto.UpdateBackupRequest     true  "Backup Configuration"
// @Success 200 {object} dto.BackupResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /backups/{id} [put]
func (h *BackupHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req dto.UpdateBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.ID = id

	resp, err := h.lifecycleService.UpdateBackup(r.Context(), req)
	if err != nil {
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "above maximum (6)") {
			errorMessage = fmt.Sprintf("Invalid Cron Expression: Day of Week (field 5 or 6) must be 0-6. You used '7', please use '0' for Sunday. Original error: %v", err)
			http.Error(w, errorMessage, http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// @Summary List backups
// @Description Get a list of backups, optionally filtered by host_id and status
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   host_id     query    string     false  "Host ID"
// @Param   status      query    string     false  "Backup Status (pending, running, completed, failed)"
// @Success 200 {array} dto.BackupResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /backups [get]
func (h *BackupHandler) List(w http.ResponseWriter, r *http.Request) {
	hostID := r.URL.Query().Get("host_id")
	status := r.URL.Query().Get("status")
	backups, err := h.queryService.ListBackups(r.Context(), hostID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(backups); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// @Summary Measure backup size
// @Description Measure the size of a directory on a host
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Host ID"
// @Param   path     body    map[string]string     true  "Path to measure (key: path)"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /hosts/{id}/measure [post]
func (h *BackupHandler) MeasureSize(w http.ResponseWriter, r *http.Request) {
	hostID := r.PathValue("id")
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	taskID, err := h.taskService.MeasureSize(r.Context(), hostID, req.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"task_id": taskID}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// @Summary Get task result
// @Description Get the result of a background task
// @Tags tasks
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Task ID"
// @Success 200 {string} string "Task Result"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Result not found"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /tasks/{id} [get]
func (h *BackupHandler) GetTaskResult(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	result, err := h.taskService.GetTaskResult(r.Context(), taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result == "" {
		http.Error(w, "Result not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(result))
}

// @Summary Get backup by ID
// @Description Get detailed information about a specific backup
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Backup ID"
// @Success 200 {object} dto.BackupResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Backup not found"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /backups/{id} [get]
func (h *BackupHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	backup, err := h.queryService.GetBackupByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(backup); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// @Summary Get backup errors
// @Description Get all error logs for a specific backup
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Backup ID"
// @Success 200 {array} dto.BackupErrorResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /backups/{id}/errors [get]
func (h *BackupHandler) GetErrors(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	errors, err := h.queryService.GetBackupErrors(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(errors); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// @Summary Delete backup errors
// @Description Delete all error logs for a specific backup
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Backup ID"
// @Success 204 "No Content"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /backups/{id}/errors [delete]
func (h *BackupHandler) DeleteErrors(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.queryService.DeleteBackupErrors(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Search files in backups
// @Description Search for files in all backups using a POSIX pattern
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   pattern     query    string     true  "POSIX pattern (e.g. *.txt)"
// @Success 200 {array} dto.FileSearchResult
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /files/search [get]
func (h *BackupHandler) SearchFiles(w http.ResponseWriter, r *http.Request) {
	pattern := r.URL.Query().Get("pattern")
	if pattern == "" {
		http.Error(w, "Pattern is required", http.StatusBadRequest)
		return
	}

	results, err := h.searchService.SearchFiles(r.Context(), pattern)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// @Summary Restore files
// @Description Trigger a restoration process
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Backup ID"
// @Param   restore     body    dto.RestoreRequest     true  "Restore Configuration"
// @Success 202 {object} map[string]string
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /backups/{id}/restore [post]
func (h *BackupHandler) Restore(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Restore] Received restore request for backup ID: %s", r.PathValue("id"))
	id := r.PathValue("id")
	var req dto.RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.BackupID = id

	taskID, err := h.restoreService.Restore(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"task_id": taskID})
}

// @Summary List files in a backup
// @Description Get a list of files and directories for a specific backup
// @Tags backups
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Backup ID"
// @Param   path   query   string     false "Subpath to list"
// @Success 200 {array} dto.BackupFileResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Backup not found"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /backups/{id}/files [get]
func (h *BackupHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	path := r.URL.Query().Get("path")

	files, err := h.searchService.ListFiles(r.Context(), id, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *BackupHandler) CreateHook(w http.ResponseWriter, r *http.Request) {
	backupID := r.PathValue("id")
	var req dto.CreateHookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.hookService.CreateHook(r.Context(), backupID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *BackupHandler) UpdateHook(w http.ResponseWriter, r *http.Request) {
	backupID := r.PathValue("id")
	hookID := r.PathValue("hookID")
	var req dto.UpdateHookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.hookService.UpdateHook(r.Context(), backupID, hookID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *BackupHandler) DeleteHook(w http.ResponseWriter, r *http.Request) {
	backupID := r.PathValue("id")
	hookID := r.PathValue("hookID")

	if err := h.hookService.DeleteHook(r.Context(), backupID, hookID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package http

import (
	"encoding/json"
	"net/http"

	"github.com/rrbarrero/justbackup/internal/notification/application"
	"github.com/rrbarrero/justbackup/internal/notification/application/dto"
	"github.com/rrbarrero/justbackup/internal/notification/domain/entities"
	"github.com/rrbarrero/justbackup/internal/shared/domain"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/auth"
)

type NotificationHandler struct {
	service *application.NotificationService
}

func NewNotificationHandler(service *application.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

func (h *NotificationHandler) RegisterRoutes(mux *http.ServeMux, protected func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /settings/notifications", protected(h.GetSettings))
	mux.HandleFunc("PUT /settings/notifications", protected(h.UpdateSettings))
}

// @Summary Get notification settings
// @Description Get notification settings for the current user
// @Tags settings
// @Accept json
// @Produce json
// @Param provider query string true "Provider Type (e.g. gotify)"
// @Success 200 {object} dto.NotificationSettingsResponse
// @Failure 404 {string} string "Settings not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /settings/notifications [get]
func (h *NotificationHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	providerType := r.URL.Query().Get("provider")
	if providerType == "" {
		http.Error(w, "provider query param is required", http.StatusBadRequest)
		return
	}

	settings, err := h.service.GetSettings(r.Context(), userID, providerType)
	if err == domain.ErrNotFound {
		// Return empty default settings instead of 404 for better UX
		json.NewEncoder(w).Encode(dto.NotificationSettingsResponse{
			ProviderType: providerType,
			Config:       make(map[string]interface{}),
			Enabled:      false,
		})
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var configMap map[string]interface{}
	if len(settings.Config) > 0 {
		if err := json.Unmarshal(settings.Config, &configMap); err != nil {
			http.Error(w, "failed to parse config", http.StatusInternalServerError)
			return
		}
	}

	resp := dto.NotificationSettingsResponse{
		ProviderType: settings.ProviderType,
		Config:       configMap,
		Enabled:      settings.Enabled,
	}

	json.NewEncoder(w).Encode(resp)
}

// @Summary Update notification settings
// @Description Update notification settings for the current user
// @Tags settings
// @Accept json
// @Produce json
// @Param settings body dto.UpdateNotificationSettingsRequest true "Notification Settings"
// @Success 200 {object} dto.NotificationSettingsResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /settings/notifications [put]
func (h *NotificationHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.UpdateNotificationSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	configBytes, err := json.Marshal(req.Config)
	if err != nil {
		http.Error(w, "Invalid config format", http.StatusBadRequest)
		return
	}

	settings := entities.NewNotificationSettings(userID, req.ProviderType, json.RawMessage(configBytes), req.Enabled)

	if err := h.service.SaveSettings(r.Context(), settings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated settings
	resp := dto.NotificationSettingsResponse{
		ProviderType: settings.ProviderType,
		Config:       req.Config,
		Enabled:      settings.Enabled,
	}

	json.NewEncoder(w).Encode(resp)
}

package http

import (
	"encoding/json"
	"net/http"

	"github.com/rrbarrero/justbackup/internal/auth/application"
)

type AuthHandler struct {
	service *application.AuthService
}

func NewAuthHandler(service *application.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux, middleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("POST /settings/token", middleware(h.GenerateToken))
	mux.HandleFunc("GET /settings/token", middleware(h.GetTokenStatus))
	mux.HandleFunc("DELETE /settings/token", middleware(h.RevokeToken))
}

// @Summary Generate API Token
// @Description Generate a new API token for CLI access
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.TokenResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /settings/token [post]
func (h *AuthHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GenerateToken(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// @Summary Get API Token Status
// @Description Get the status of the current API token
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.TokenResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /settings/token [get]
func (h *AuthHandler) GetTokenStatus(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetTokenStatus(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// @Summary Revoke API Token
// @Description Revoke the current API token
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 204 "No Content"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /settings/token [delete]
func (h *AuthHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	if err := h.service.RevokeToken(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

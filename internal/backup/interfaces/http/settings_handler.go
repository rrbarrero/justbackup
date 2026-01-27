package http

import (
	"encoding/json"
	"net/http"
	"os"
)

type SettingsHandler struct {
	sshKeyPath string
}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{
		sshKeyPath: os.Getenv("SSH_PUBLIC_KEY_PATH"),
	}
}

func (h *SettingsHandler) RegisterRoutes(mux *http.ServeMux, middleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /settings/ssh-key", middleware(h.GetSSHKey))
}

// @Summary Get SSH Public Key
// @Description Get the SSH public key for the server
// @Tags settings
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /settings/ssh-key [get]
func (h *SettingsHandler) GetSSHKey(w http.ResponseWriter, r *http.Request) {
	if h.sshKeyPath == "" {
		http.Error(w, "SSH public key path not configured", http.StatusInternalServerError)
		return
	}

	key, err := os.ReadFile(h.sshKeyPath)
	if err != nil {
		http.Error(w, "Failed to read SSH public key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"publicKey": string(key),
	})
}

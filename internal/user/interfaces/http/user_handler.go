package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/auth"
	"github.com/rrbarrero/justbackup/internal/user/application"
)

type UserHandler struct {
	userService *application.UserService
	jwtService  *auth.JWTService
}

func NewUserHandler(userService *application.UserService, jwtService *auth.JWTService) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

type SetupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type SetupStatusResponse struct {
	SetupRequired bool `json:"setupRequired"`
}

// @Summary Get setup status
// @Description Check if the initial setup is required
// @Tags user
// @Accept  json
// @Produce  json
// @Success 200 {object} SetupStatusResponse
// @Failure 500 {string} string "Internal Server Error"
// @Router /setup-status [get]
func (h *UserHandler) GetSetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	required, err := h.userService.IsSetupRequired(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Explicitly ignore the error as the connection might be closed
	if err := json.NewEncoder(w).Encode(SetupStatusResponse{SetupRequired: required}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// @Summary Initial setup
// @Description Register the initial user
// @Tags user
// @Accept  json
// @Produce  json
// @Param   setup     body    SetupRequest     true  "Setup Credentials"
// @Success 201 "Created"
// @Failure 400 {string} string "Invalid request body"
// @Router /setup [post]
func (h *UserHandler) Setup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	err := h.userService.RegisterInitialUser(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // Could be "setup already completed"
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// @Summary Login
// @Description Authenticate user and get JWT token
// @Tags user
// @Accept  json
// @Produce  json
// @Param   login     body    LoginRequest     true  "Login Credentials"
// @Success 200 {object} LoginResponse
// @Failure 401 {string} string "Invalid credentials"
// @Failure 500 {string} string "Internal Server Error"
// @Router /login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(LoginResponse{Token: token}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

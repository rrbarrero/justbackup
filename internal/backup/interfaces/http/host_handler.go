package http

import (
	"encoding/json"
	"net/http"

	"github.com/rrbarrero/justbackup/internal/backup/application"
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
)

type HostHandler struct {
	service *application.HostService
}

func NewHostHandler(service *application.HostService) *HostHandler {
	return &HostHandler{
		service: service,
	}
}

func (h *HostHandler) RegisterRoutes(mux *http.ServeMux, middleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /hosts", middleware(h.List))
	mux.HandleFunc("POST /hosts", middleware(h.Create))
	mux.HandleFunc("GET /hosts/{id}", middleware(h.Get))
	mux.HandleFunc("PUT /hosts/{id}", middleware(h.Update))
	mux.HandleFunc("DELETE /hosts/{id}", middleware(h.Delete))
}

// @Summary Create a host
// @Description Create a new host configuration
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param   host     body    dto.CreateHostRequest     true  "Host Configuration"
// @Success 201 {object} dto.HostResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /hosts [post]
func (h *HostHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreateHost(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// @Summary List hosts
// @Description Get a list of all hosts
// @Tags hosts
// @Accept  json
// @Produce  json
// @Success 200 {array} dto.HostResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /hosts [get]
func (h *HostHandler) List(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.service.ListHosts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(hosts)
}

// @Summary Get a host
// @Description Get details of a specific host
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Host ID"
// @Success 200 {object} dto.HostResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Host not found"
// @Security BasicAuth
// @Router /hosts/{id} [get]
func (h *HostHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	host, err := h.service.GetHost(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(host)
}

// @Summary Update a host
// @Description Update an existing host configuration
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Host ID"
// @Param   host     body    dto.UpdateHostRequest     true  "Host Configuration"
// @Success 200 {object} dto.HostResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /hosts/{id} [put]
func (h *HostHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req dto.UpdateHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.ID = id
	resp, err := h.service.UpdateHost(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// @Summary Delete a host
// @Description Delete a host configuration
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param   id     path    string     true  "Host ID"
// @Success 204 "No Content"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security BasicAuth
// @Router /hosts/{id} [delete]
func (h *HostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.DeleteHost(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

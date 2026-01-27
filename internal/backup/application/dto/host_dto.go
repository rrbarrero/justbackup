package dto

import (
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
)

type CreateHostRequest struct {
	Name          string `json:"name"`
	Hostname      string `json:"hostname"`
	User          string `json:"user"`
	Port          int    `json:"port"`
	Path          string `json:"path"`
	IsWorkstation bool   `json:"is_workstation"`
}

type HostResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Hostname           string `json:"hostname"`
	User               string `json:"user"`
	Port               int    `json:"port"`
	Path               string `json:"path"`
	IsWorkstation      bool   `json:"is_workstation"`
	FailedBackupsCount int    `json:"failed_backups_count"`
}

func ToHostResponse(h *entities.Host) *HostResponse {
	return &HostResponse{
		ID:            h.ID().String(),
		Name:          h.Name(),
		Hostname:      h.Hostname(),
		User:          h.User(),
		Port:          h.Port(),
		Path:          h.Path(),
		IsWorkstation: h.IsWorkstation(),
	}
}

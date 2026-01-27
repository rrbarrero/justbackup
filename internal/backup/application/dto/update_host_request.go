package dto

type UpdateHostRequest struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Hostname      string `json:"hostname"`
	User          string `json:"user"`
	Port          int    `json:"port"`
	Path          string `json:"path"`
	IsWorkstation bool   `json:"is_workstation"`
}

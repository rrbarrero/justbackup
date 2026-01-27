package dto

type HookDTO struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Phase   string            `json:"phase"`
	Enabled bool              `json:"enabled"`
	Params  map[string]string `json:"params"`
}

type CreateHookRequest struct {
	Name    string            `json:"name" binding:"required"`
	Phase   string            `json:"phase" binding:"required"` // "pre" or "post"
	Params  map[string]string `json:"params"`
	Enabled bool              `json:"enabled"`
}

type UpdateHookRequest struct {
	Name    string            `json:"name"`
	Phase   string            `json:"phase"`
	Params  map[string]string `json:"params"`
	Enabled *bool             `json:"enabled"`
}

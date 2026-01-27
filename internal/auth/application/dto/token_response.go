package dto

import "time"

type TokenResponse struct {
	Token     string    `json:"token,omitempty"` // Only present when generated
	CreatedAt time.Time `json:"created_at"`
	Exists    bool      `json:"exists"`
}

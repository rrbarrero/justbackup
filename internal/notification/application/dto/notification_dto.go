package dto

type NotificationSettingsResponse struct {
	ProviderType string                 `json:"provider_type"`
	Config       map[string]interface{} `json:"config"`
	Enabled      bool                   `json:"enabled"`
}

type UpdateNotificationSettingsRequest struct {
	ProviderType string                 `json:"provider_type"`
	Config       map[string]interface{} `json:"config"`
	Enabled      bool                   `json:"enabled"`
}

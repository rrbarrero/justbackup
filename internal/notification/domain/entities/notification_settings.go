package entities

import "encoding/json"

type NotificationSettings struct {
	UserID       int             `json:"user_id"`
	ProviderType string          `json:"provider_type"`
	Config       json.RawMessage `json:"config"`
	Enabled      bool            `json:"enabled"`
}

func NewNotificationSettings(userID int, providerType string, config json.RawMessage, enabled bool) *NotificationSettings {
	return &NotificationSettings{
		UserID:       userID,
		ProviderType: providerType,
		Config:       config,
		Enabled:      enabled,
	}
}

// GotifyConfig helper struct for unmarshalling
type GotifyConfig struct {
	URL             string `json:"url"`
	Token           string `json:"token"`
	NotifyOnSuccess bool   `json:"notify_on_success"`
}

// SMTPConfig helper struct for unmarshalling
type SMTPConfig struct {
	Host            string   `json:"host"`
	Port            int      `json:"port"`
	User            string   `json:"user"`
	Password        string   `json:"password"`
	From            string   `json:"from"`
	To              []string `json:"to"`
	NotifyOnSuccess bool     `json:"notify_on_success"`
}

// PushbulletConfig helper struct for unmarshalling
type PushbulletConfig struct {
	AccessToken     string `json:"access_token"`
	NotifyOnSuccess bool   `json:"notify_on_success"`
}

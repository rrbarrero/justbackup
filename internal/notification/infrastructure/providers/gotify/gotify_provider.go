package gotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rrbarrero/justbackup/internal/notification/domain/entities"
	"github.com/rrbarrero/justbackup/internal/notification/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/notification/domain/valueobjects"
)

func init() {
	interfaces.RegisterProvider("gotify", func(config json.RawMessage) (interfaces.NotificationProvider, bool, error) {
		var c entities.GotifyConfig
		if err := json.Unmarshal(config, &c); err != nil {
			return nil, false, err
		}
		return NewGotifyProvider(c.URL, c.Token), c.NotifyOnSuccess, nil
	})
}

type GotifyProvider struct {
	url   string
	token string
}

func NewGotifyProvider(url, token string) *GotifyProvider {
	return &GotifyProvider{
		url:   url,
		token: token,
	}
}

func (p *GotifyProvider) Type() string {
	return "gotify"
}

type gotifyMessage struct {
	Title    string                 `json:"title"`
	Message  string                 `json:"message"`
	Priority int                    `json:"priority"`
	Extras   map[string]interface{} `json:"extras,omitempty"`
}

func (p *GotifyProvider) Send(ctx context.Context, title, message string, level valueobjects.NotificationLevel) error {
	priority := 5
	switch level {
	case valueobjects.Info:
		priority = 5
	case valueobjects.Warning:
		priority = 8
	case valueobjects.Error:
		priority = 10
	}

	msg := gotifyMessage{
		Title:    title,
		Message:  message,
		Priority: priority,
		Extras: map[string]interface{}{
			"client::display": map[string]string{
				"contentType": "text/markdown",
			},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal gotify message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/message", p.url), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", p.token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to gotify: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("gotify returned error status: %d", resp.StatusCode)
	}

	return nil
}

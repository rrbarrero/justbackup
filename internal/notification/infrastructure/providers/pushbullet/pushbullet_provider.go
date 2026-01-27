package pushbullet

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
	interfaces.RegisterProvider("pushbullet", func(config json.RawMessage) (interfaces.NotificationProvider, bool, error) {
		var c entities.PushbulletConfig
		if err := json.Unmarshal(config, &c); err != nil {
			return nil, false, err
		}
		return NewPushbulletProvider(c.AccessToken), c.NotifyOnSuccess, nil
	})
}

type PushbulletProvider struct {
	accessToken string
}

func NewPushbulletProvider(accessToken string) *PushbulletProvider {
	return &PushbulletProvider{
		accessToken: accessToken,
	}
}

func (p *PushbulletProvider) Type() string {
	return "pushbullet"
}

type pushbulletMessage struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (p *PushbulletProvider) Send(ctx context.Context, title, message string, level valueobjects.NotificationLevel) error {
	// Add prefix based on level for better visibility in Pushbullet
	prefix := ""
	switch level {
	case valueobjects.Warning:
		prefix = "⚠️ "
	case valueobjects.Error:
		prefix = "❌ "
	case valueobjects.Info:
		prefix = "✅ "
	}

	msg := pushbulletMessage{
		Type:  "note",
		Title: fmt.Sprintf("%s%s", prefix, title),
		Body:  message,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal pushbullet message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.pushbullet.com/v2/pushes", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Token", p.accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to pushbullet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("pushbullet returned error status: %d", resp.StatusCode)
	}

	return nil
}

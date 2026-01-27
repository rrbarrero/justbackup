package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/rrbarrero/justbackup/internal/notification/domain/entities"
)

type NotificationRepositoryMemory struct {
	settings map[string]*entities.NotificationSettings // Key: userID:providerType
	mu       sync.RWMutex
}

func NewNotificationRepositoryMemory() *NotificationRepositoryMemory {
	return &NotificationRepositoryMemory{
		settings: make(map[string]*entities.NotificationSettings),
	}
}

func (r *NotificationRepositoryMemory) GetSettings(ctx context.Context, userID int, providerType string) (*entities.NotificationSettings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%d:%s", userID, providerType)
	if setting, exists := r.settings[key]; exists {
		return setting, nil
	}
	return nil, nil // Or error not found, but usually nil means no settings yet
}

func (r *NotificationRepositoryMemory) SaveSettings(ctx context.Context, settings *entities.NotificationSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%d:%s", settings.UserID, settings.ProviderType)
	r.settings[key] = settings
	return nil
}

func (r *NotificationRepositoryMemory) GetAllEnabledSettings(ctx context.Context) ([]*entities.NotificationSettings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var enabled []*entities.NotificationSettings
	for _, s := range r.settings {
		if s.Enabled {
			enabled = append(enabled, s)
		}
	}
	return enabled, nil
}

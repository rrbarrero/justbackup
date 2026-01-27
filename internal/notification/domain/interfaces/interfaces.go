package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rrbarrero/justbackup/internal/notification/domain/entities"
	"github.com/rrbarrero/justbackup/internal/notification/domain/valueobjects"
)

// NotificationProvider defines the contract for any notification service
type NotificationProvider interface {
	Send(ctx context.Context, title, message string, level valueobjects.NotificationLevel) error
	Type() string
}

// ProviderFactory is a function that creates a provider from its configuration
type ProviderFactory func(config json.RawMessage) (NotificationProvider, bool, error)

var (
	providersMu sync.RWMutex
	providers   = make(map[string]ProviderFactory)
)

// RegisterProvider registers a notification provider factory
func RegisterProvider(providerType string, factory ProviderFactory) {
	providersMu.Lock()
	defer providersMu.Unlock()
	providers[providerType] = factory
}

// GetProviderFactory returns a registered provider factory
func GetProviderFactory(providerType string) (ProviderFactory, error) {
	providersMu.RLock()
	defer providersMu.RUnlock()
	factory, ok := providers[providerType]
	if !ok {
		return nil, fmt.Errorf("provider type %s not registered", providerType)
	}
	return factory, nil
}

// NotificationRepository defines how to persist notification settings
type NotificationRepository interface {
	GetSettings(ctx context.Context, userID int, providerType string) (*entities.NotificationSettings, error)
	SaveSettings(ctx context.Context, settings *entities.NotificationSettings) error
	GetAllEnabledSettings(ctx context.Context) ([]*entities.NotificationSettings, error)
}

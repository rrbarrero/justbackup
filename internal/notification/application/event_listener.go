package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rrbarrero/justbackup/internal/backup/domain/events"
	"github.com/rrbarrero/justbackup/internal/notification/domain/valueobjects"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/event"
)

type NotificationEventListener struct {
	service  *NotificationService
	eventBus *event.RedisEventBus
}

func NewNotificationEventListener(service *NotificationService, eventBus *event.RedisEventBus) *NotificationEventListener {
	return &NotificationEventListener{
		service:  service,
		eventBus: eventBus,
	}
}

func (l *NotificationEventListener) Start(ctx context.Context) {
	// Subscribe to BackupFailed
	l.eventBus.Subscribe(ctx, events.BackupFailedEvent, func(data []byte) error {
		var event events.BackupFailed
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal BackupFailed event: %w", err)
		}
		return l.handleBackupFailed(ctx, event)
	})

	// Subscribe to BackupCompleted event.
	l.eventBus.Subscribe(ctx, events.BackupCompletedEvent, func(data []byte) error {
		var event events.BackupCompleted
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal BackupCompleted event: %w", err)
		}
		return l.handleBackupCompleted(ctx, event)
	})
}

func (l *NotificationEventListener) handleBackupFailed(ctx context.Context, event events.BackupFailed) error {
	title := "Backup Failed"
	message := fmt.Sprintf("Backup %s for host '%s' (Source: %s) failed: %s", event.BackupID, event.HostName, event.SourcePath, event.ErrorMessage)

	// We notify all users for now (since we don't have user-specific backup ownership yet)
	// The service iterates over all enabled settings.
	return l.service.Notify(ctx, title, message, valueobjects.Error)
}

func (l *NotificationEventListener) handleBackupCompleted(ctx context.Context, event events.BackupCompleted) error {
	// Handle successful backup notifications.
	title := "Backup Completed"
	message := fmt.Sprintf("Backup %s for host '%s' (Source: %s) completed successfully. Size: %s", event.BackupID, event.HostName, event.SourcePath, event.Size)

	return l.service.Notify(ctx, title, message, valueobjects.Info)
}

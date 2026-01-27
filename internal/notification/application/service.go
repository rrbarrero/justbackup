package application

import (
	"context"
	"fmt"
	"log"

	"github.com/rrbarrero/justbackup/internal/notification/domain/entities"
	"github.com/rrbarrero/justbackup/internal/notification/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/notification/domain/valueobjects"
)

type NotificationService struct {
	repo interfaces.NotificationRepository
}

func NewNotificationService(repo interfaces.NotificationRepository) *NotificationService {
	return &NotificationService{
		repo: repo,
	}
}

func (s *NotificationService) Notify(ctx context.Context, title, message string, level valueobjects.NotificationLevel) error {
	// 1. Get all enabled settings
	settingsList, err := s.repo.GetAllEnabledSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch notification settings: %w", err)
	}

	if len(settingsList) == 0 {
		return nil
	}

	// 2. For each setting, instantiate provider and send
	for _, settings := range settingsList {
		provider, notifyOnSuccess, err := s.createProvider(settings)
		if err != nil {
			log.Printf("Failed to create provider for user %d: %v", settings.UserID, err)
			continue
		}

		// Filter out success notifications if not enabled
		if level == valueobjects.Info && !notifyOnSuccess {
			continue
		}

		if err := provider.Send(ctx, title, message, level); err != nil {
			log.Printf("Failed to send notification to user %d via %s: %v", settings.UserID, settings.ProviderType, err)
			// We continue to next user/provider even if one fails
		}
	}

	return nil
}

func (s *NotificationService) createProvider(settings *entities.NotificationSettings) (interfaces.NotificationProvider, bool, error) {
	factory, err := interfaces.GetProviderFactory(settings.ProviderType)
	if err != nil {
		return nil, false, err
	}
	return factory(settings.Config)
}

func (s *NotificationService) GetSettings(ctx context.Context, userID int, providerType string) (*entities.NotificationSettings, error) {
	return s.repo.GetSettings(ctx, userID, providerType)
}

func (s *NotificationService) SaveSettings(ctx context.Context, settings *entities.NotificationSettings) error {
	return s.repo.SaveSettings(ctx, settings)
}

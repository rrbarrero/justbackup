package application

import (
	"context"
	"log"
	"time"

	backupEntities "github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	backupInterfaces "github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/maintenance/domain/entities"
	"github.com/rrbarrero/justbackup/internal/maintenance/domain/interfaces"
)

type MaintenanceTaskPublisher interface {
	PublishPurgeTask(ctx context.Context, backup *backupEntities.Backup) error
}

type MaintenanceService struct {
	repo       interfaces.MaintenanceTaskRepository
	backupRepo backupInterfaces.BackupRepository
	publisher  MaintenanceTaskPublisher
}

func NewMaintenanceService(
	repo interfaces.MaintenanceTaskRepository,
	backupRepo backupInterfaces.BackupRepository,
	publisher MaintenanceTaskPublisher,
) *MaintenanceService {
	return &MaintenanceService{
		repo:       repo,
		backupRepo: backupRepo,
		publisher:  publisher,
	}
}

func (s *MaintenanceService) ProcessDueTasks(ctx context.Context) error {
	tasks, err := s.repo.FindDueTasks(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		log.Printf("Processing maintenance task: %s (%s)", task.Name(), task.Type())

		if err := s.executeTask(ctx, task); err != nil {
			log.Printf("Error executing maintenance task %s: %v", task.ID(), err)
			continue
		}

		task.SetLastRun(time.Now())
		if err := task.CalculateNextRun(); err != nil {
			log.Printf("Error calculating next run for task %s: %v", task.ID(), err)
		}

		if err := s.repo.Save(ctx, task); err != nil {
			log.Printf("Error saving maintenance task %s: %v", task.ID(), err)
		}
	}

	return nil
}

func (s *MaintenanceService) executeTask(ctx context.Context, task *entities.MaintenanceTask) error {
	switch task.Type() {
	case entities.MaintenanceTaskTypePurge:
		return s.purgeIncrementalBackups(ctx)
	default:
		log.Printf("Unknown maintenance task type: %s", task.Type())
		return nil
	}
}

func (s *MaintenanceService) purgeIncrementalBackups(ctx context.Context) error {
	backups, err := s.backupRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	for _, backup := range backups {
		if !backup.Incremental() || backup.Retention() <= 0 {
			continue
		}

		log.Printf("Queueing purge task for backup: %s (retention: %d)", backup.ID(), backup.Retention())
		if err := s.publisher.PublishPurgeTask(ctx, backup); err != nil {
			log.Printf("Failed to publish purge task for backup %s: %v", backup.ID(), err)
		}
	}

	return nil
}

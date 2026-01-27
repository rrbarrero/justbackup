package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	maintApp "github.com/rrbarrero/justbackup/internal/maintenance/application"
)

type Scheduler struct {
	repo         interfaces.BackupRepository
	maintService *maintApp.MaintenanceService
	publisher    *RedisPublisher
	interval     time.Duration
}

func NewScheduler(repo interfaces.BackupRepository, maintService *maintApp.MaintenanceService, publisher *RedisPublisher, interval time.Duration) *Scheduler {
	return &Scheduler{
		repo:         repo,
		maintService: maintService,
		publisher:    publisher,
		interval:     interval,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Println("Scheduler started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler stopped")
			return
		case <-ticker.C:
			if err := s.processDueBackups(ctx); err != nil {
				log.Printf("Error processing due backups: %v", err)
			}
			if s.maintService != nil {
				if err := s.maintService.ProcessDueTasks(ctx); err != nil {
					log.Printf("Error processing due maintenance tasks: %v", err)
				}
			}
		}
	}
}

func (s *Scheduler) processDueBackups(ctx context.Context) error {
	backups, err := s.repo.FindDueBackups(ctx)
	if err != nil {
		return err
	}

	for _, backup := range backups {
		log.Printf("Processing due backup: %s", backup.ID())

		// Publish to Redis
		if err := s.publisher.Publish(ctx, backup); err != nil {
			log.Printf("Failed to publish backup %s: %v", backup.ID(), err)
			continue
		}

		// Update NextRunAt
		if err := backup.CalculateNextRun(); err != nil {
			log.Printf("Failed to calculate next run for backup %s: %v", backup.ID(), err)
			continue
		}

		// Persist changes
		if err := s.repo.Save(ctx, backup); err != nil {
			log.Printf("Failed to save backup %s: %v", backup.ID(), err)
		}
	}

	return nil
}

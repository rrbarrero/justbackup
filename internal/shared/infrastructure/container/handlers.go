package container

import (
	"github.com/redis/go-redis/v9"
	authHttp "github.com/rrbarrero/justbackup/internal/auth/interfaces/http"
	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/memory"
	backupHttp "github.com/rrbarrero/justbackup/internal/backup/interfaces/http"
	notifHttp "github.com/rrbarrero/justbackup/internal/notification/interfaces/http"
	userHttp "github.com/rrbarrero/justbackup/internal/user/interfaces/http"
	workerStatsHttp "github.com/rrbarrero/justbackup/internal/workerstats/interfaces/http"
)

// initializeHandlers initializes all HTTP handlers
func (c *Container) initializeHandlers(services *Services, repos *Repositories, env string, redisClient *redis.Client) *Handlers {
	handlers := &Handlers{
		Backup: backupHttp.NewBackupHandler(
			services.BackupLifecycle,
			services.BackupQuery,
			services.BackupSearch,
			services.BackupRestore,
			services.BackupTask,
			services.BackupHook,
		),
		Host:         backupHttp.NewHostHandler(services.Host),
		Settings:     backupHttp.NewSettingsHandler(),
		User:         userHttp.NewUserHandler(services.User, services.JWT),
		Auth:         authHttp.NewAuthHandler(services.Auth),
		Notification: notifHttp.NewNotificationHandler(services.Notification),
		Dashboard:    backupHttp.NewDashboardHandler(services.Dashboard),
		System:       backupHttp.NewSystemHandler(redisClient),
		WorkerStats:  workerStatsHttp.NewWorkerStatsHandler(services.WorkerStats),
	}

	if env == "dev" || env == "test" || env == "e2e" {
		if memErrorRepo, ok := repos.BackupError.(*memory.BackupErrorRepositoryMemory); ok {
			handlers.Test = backupHttp.NewTestHandler(memErrorRepo)
		}
	}

	return handlers
}

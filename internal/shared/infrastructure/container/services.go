package container

import (
	authApp "github.com/rrbarrero/justbackup/internal/auth/application"
	"github.com/rrbarrero/justbackup/internal/backup/application"
	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	maintApp "github.com/rrbarrero/justbackup/internal/maintenance/application"
	notifApp "github.com/rrbarrero/justbackup/internal/notification/application"
	"github.com/rrbarrero/justbackup/internal/scheduler"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/auth"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	userApp "github.com/rrbarrero/justbackup/internal/user/application"
	workerStatsApp "github.com/rrbarrero/justbackup/internal/workerstats/application"
)

// initializeServices initializes all application services
func (c *Container) initializeServices(repos *Repositories, redisPublisher *scheduler.RedisPublisher, resultStore *scheduler.RedisResultStore, workerQueryBus interfaces.WorkerQueryBus, cfg *config.ServerConfig) *Services {
	hostService := application.NewHostService(repos.Host, repos.Backup)
	backupAssembler := assembler.NewBackupAssembler()

	return &Services{
		Host:            hostService,
		BackupLifecycle: application.NewBackupLifecycleService(repos.Backup, hostService, redisPublisher, backupAssembler),
		BackupQuery:     application.NewBackupQueryService(repos.Backup, hostService, repos.BackupError, backupAssembler),
		BackupSearch:    application.NewBackupSearchService(repos.Backup, hostService, workerQueryBus, backupAssembler),
		BackupRestore:   application.NewBackupRestoreService(repos.Backup, hostService, redisPublisher),
		BackupTask:      application.NewBackupTaskService(redisPublisher, resultStore),
		BackupHook:      application.NewBackupHookService(repos.Backup, backupAssembler),
		BackupAssembler: backupAssembler,
		User:            userApp.NewUserService(repos.User),
		Auth:            authApp.NewAuthService(repos.AuthToken),
		Notification:    notifApp.NewNotificationService(repos.Notification),
		Dashboard:       application.NewDashboardService(repos.Backup, repos.Host, workerStatsApp.NewWorkerStatsService(repos.WorkerStats, c.webSocketHub)),
		Maintenance:     maintApp.NewMaintenanceService(repos.Maintenance, repos.Backup, redisPublisher),
		JWT:             auth.NewJWTService(cfg.JWTSecret, "justbackup"),
		WorkerStats:     workerStatsApp.NewWorkerStatsService(repos.WorkerStats, c.webSocketHub),
	}
}

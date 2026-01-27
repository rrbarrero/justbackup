package container

import (
	"net/http"

	"github.com/redis/go-redis/v9"
	authApp "github.com/rrbarrero/justbackup/internal/auth/application"
	authInterfaces "github.com/rrbarrero/justbackup/internal/auth/domain/interfaces"
	authHttp "github.com/rrbarrero/justbackup/internal/auth/interfaces/http"
	"github.com/rrbarrero/justbackup/internal/backup/application"
	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	backupHttp "github.com/rrbarrero/justbackup/internal/backup/interfaces/http"
	maintApp "github.com/rrbarrero/justbackup/internal/maintenance/application"
	maintInterfaces "github.com/rrbarrero/justbackup/internal/maintenance/domain/interfaces"
	notifApp "github.com/rrbarrero/justbackup/internal/notification/application"
	notifInterfaces "github.com/rrbarrero/justbackup/internal/notification/domain/interfaces"
	notifHttp "github.com/rrbarrero/justbackup/internal/notification/interfaces/http"
	"github.com/rrbarrero/justbackup/internal/scheduler"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/auth"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/event"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/module"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/websocket"
	userApp "github.com/rrbarrero/justbackup/internal/user/application"
	userInterfaces "github.com/rrbarrero/justbackup/internal/user/domain/interfaces"
	userHttp "github.com/rrbarrero/justbackup/internal/user/interfaces/http"
	workerStatsApp "github.com/rrbarrero/justbackup/internal/workerstats/application"
	workerStatsInterfaces "github.com/rrbarrero/justbackup/internal/workerstats/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/workerstats/infrastructure/monitoring"
	workerStatsHttp "github.com/rrbarrero/justbackup/internal/workerstats/interfaces/http"
)

// Repositories holds all repository implementations
type Repositories struct {
	Backup       interfaces.BackupRepository
	Host         interfaces.HostRepository
	User         userInterfaces.UserRepository
	AuthToken    authInterfaces.AuthTokenRepository
	BackupError  interfaces.BackupErrorRepository
	Notification notifInterfaces.NotificationRepository
	Maintenance  maintInterfaces.MaintenanceTaskRepository
	WorkerStats  workerStatsInterfaces.WorkerStatsRepository
}

// Services holds all application services
type Services struct {
	Host            *application.HostService
	BackupLifecycle *application.BackupLifecycleService
	BackupQuery     *application.BackupQueryService
	BackupSearch    *application.BackupSearchService
	BackupRestore   *application.BackupRestoreService
	BackupTask      *application.BackupTaskService
	BackupHook      *application.BackupHookService
	BackupAssembler *assembler.BackupAssembler
	User            *userApp.UserService
	Auth            *authApp.AuthService
	Notification    *notifApp.NotificationService
	Dashboard       *application.DashboardService
	Maintenance     *maintApp.MaintenanceService
	JWT             *auth.JWTService
	WorkerStats     *workerStatsApp.WorkerStatsService
}

// Handlers holds all HTTP handlers
type Handlers struct {
	Backup       *backupHttp.BackupHandler
	Host         *backupHttp.HostHandler
	Settings     *backupHttp.SettingsHandler
	User         *userHttp.UserHandler
	Auth         *authHttp.AuthHandler
	Notification *notifHttp.NotificationHandler
	Dashboard    *backupHttp.DashboardHandler
	Test         *backupHttp.TestHandler
	System       *backupHttp.SystemHandler
	WorkerStats  *workerStatsHttp.WorkerStatsHandler
}

// ServiceContainer defines the interface for a service container
type ServiceContainer interface {
	GetConfig() *config.ServerConfig
	GetRedisClient() *redis.Client
	GetEventBus() *event.RedisEventBus
	GetBackupScheduler() *scheduler.Scheduler
	GetResultConsumer() *scheduler.ResultConsumer
	GetNotificationListener() *notifApp.NotificationEventListener
	GetWebSocketHub() *websocket.Hub
	GetHandler() http.Handler
	GetRepositories() *Repositories
	GetServices() *Services
	GetHandlers() *Handlers
}

// Container holds all application dependencies
type Container struct {
	config               *config.ServerConfig
	redisClient          *redis.Client
	eventBus             *event.RedisEventBus
	backupScheduler      *scheduler.Scheduler
	resultConsumer       *scheduler.ResultConsumer
	notificationListener *notifApp.NotificationEventListener
	webSocketHub         *websocket.Hub
	handler              http.Handler
	repositories         *Repositories
	services             *Services
	handlers             *Handlers
	moduleManager        *module.ModuleManager
	httpServerModule     *module.HTTPServerModule
	redisModule          *module.RedisModule
	eventBusModule       *module.EventBusModule
	webSocketModule      *module.WebSocketModule
	serverStatsCollector *monitoring.ServerStatsCollector
}

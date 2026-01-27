package container

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	notifApp "github.com/rrbarrero/justbackup/internal/notification/application"
	"github.com/rrbarrero/justbackup/internal/scheduler"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/event"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/module"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/websocket"
	"github.com/rrbarrero/justbackup/internal/workerstats/infrastructure/monitoring"

	// Notification providers registration
	_ "github.com/rrbarrero/justbackup/internal/notification/infrastructure/providers/gotify"
	_ "github.com/rrbarrero/justbackup/internal/notification/infrastructure/providers/pushbullet"
	_ "github.com/rrbarrero/justbackup/internal/notification/infrastructure/providers/smtp"
)

// NewContainer creates a new service container
func NewContainer() (*Container, error) {
	return &Container{}, nil
}

// Initialize sets up all application dependencies using modular approach
func (c *Container) Initialize() error {
	// Create config service and load configuration
	configService := config.NewConfigService()
	cfg, err := configService.LoadServerConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	env := cfg.Environment

	// Initialize repositories (needed for services that depend on them)
	repos, err := c.initializeRepositories(cfg, env)
	if err != nil {
		return fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Seed Auto-Backup configuration
	if err := SeedSelfBackup(repos.Host, repos.Backup, cfg); err != nil {
		log.Printf("WARNING: Failed to seed self-backup: %v", err)
		// We don't fail startup for this, just log warning
	}

	// Initialize Redis client first (needed for other components)
	redisAddr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	c.redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Initialize event bus
	c.eventBus = event.NewRedisEventBus(c.redisClient)

	// Initialize WebSocket hub first
	webSocketHub := websocket.NewHub()
	c.webSocketHub = webSocketHub

	// Initialize services with proper Redis components
	redisPublisher := scheduler.NewRedisPublisher(c.redisClient, "backup_tasks", repos.Host)
	resultStore := scheduler.NewRedisResultStore(c.redisClient)
	workerQueryBus := scheduler.NewRedisWorkerQueryBus(c.redisClient, redisPublisher)
	services := c.initializeServices(repos, redisPublisher, resultStore, workerQueryBus, cfg)

	// Initialize scheduler components
	c.backupScheduler = scheduler.NewScheduler(repos.Backup, services.Maintenance, redisPublisher, 1*time.Minute) // Use 1 minute as default

	c.resultConsumer = scheduler.NewResultConsumer(c.redisClient, "backup_results", repos.Backup, services.Host, repos.BackupError, webSocketHub, c.eventBus)

	// Initialize notification listener
	c.notificationListener = notifApp.NewNotificationEventListener(services.Notification, c.eventBus)

	// Initialize handlers using the fully initialized components
	handlers := c.initializeHandlers(services, repos, env, c.redisClient)

	// Setup router
	handler := c.setupRouter(handlers, services, webSocketHub, cfg)

	// Create and register modules
	c.moduleManager = module.NewModuleManager()
	c.redisModule = module.NewRedisModule(cfg, c.redisClient)
	c.eventBusModule = module.NewEventBusModule(cfg, c.redisModule, c.notificationListener)
	c.webSocketModule = module.NewWebSocketModule(cfg, c.webSocketHub)
	c.httpServerModule = module.NewHTTPServerModule(cfg, handler)
	schedulerModule := module.NewSchedulerModule(cfg, c.backupScheduler, c.resultConsumer)

	c.moduleManager.RegisterModule(c.redisModule)
	c.moduleManager.RegisterModule(c.eventBusModule)
	c.moduleManager.RegisterModule(c.webSocketModule)
	c.moduleManager.RegisterModule(schedulerModule)
	c.moduleManager.RegisterModule(c.httpServerModule)

	// Initialize configuration and other fields
	c.config = cfg
	c.repositories = repos
	c.services = services
	c.handlers = handlers
	c.handler = handler

	// Initialize Server Stats Collector
	c.serverStatsCollector = monitoring.NewServerStatsCollector(cfg, services.WorkerStats)

	return nil
}

// Run starts all modules
func (c *Container) Run(ctx context.Context) error {
	// Initialize all modules
	if err := c.moduleManager.InitializeAll(c.config); err != nil {
		return fmt.Errorf("failed to initialize modules: %w", err)
	}

	// Start all modules
	if err := c.moduleManager.StartAll(ctx); err != nil {
		return fmt.Errorf("failed to start modules: %w", err)
	}

	// Start backend statistics collection
	if c.serverStatsCollector != nil {
		go c.serverStatsCollector.Start(ctx)
	}

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// Close properly closes all resources and stops all modules
func (c *Container) Close() error {
	// Stop all modules
	if c.moduleManager != nil {
		if err := c.moduleManager.StopAll(); err != nil {
			return fmt.Errorf("error stopping modules: %w", err)
		}
	}

	// Close Redis client if not managed by modules
	if c.redisClient != nil {
		return c.redisClient.Close()
	}
	return nil
}

// GetConfig returns the server configuration
func (c *Container) GetConfig() *config.ServerConfig {
	return c.config
}

// GetRedisClient returns the Redis client
func (c *Container) GetRedisClient() *redis.Client {
	return c.redisClient
}

// GetEventBus returns the event bus
func (c *Container) GetEventBus() *event.RedisEventBus {
	return c.eventBus
}

// GetBackupScheduler returns the backup scheduler
func (c *Container) GetBackupScheduler() *scheduler.Scheduler {
	return c.backupScheduler
}

// GetResultConsumer returns the result consumer
func (c *Container) GetResultConsumer() *scheduler.ResultConsumer {
	return c.resultConsumer
}

// GetNotificationListener returns the notification listener
func (c *Container) GetNotificationListener() *notifApp.NotificationEventListener {
	return c.notificationListener
}

// GetWebSocketHub returns the WebSocket hub
func (c *Container) GetWebSocketHub() *websocket.Hub {
	return c.webSocketHub
}

// GetHandler returns the HTTP handler
func (c *Container) GetHandler() http.Handler {
	return c.handler
}

// GetRepositories returns the repositories
func (c *Container) GetRepositories() *Repositories {
	return c.repositories
}

// GetServices returns the services
func (c *Container) GetServices() *Services {
	return c.services
}

// GetHandlers returns the handlers
func (c *Container) GetHandlers() *Handlers {
	return c.handlers
}

// InitializeContainer creates and initializes all application dependencies
func InitializeContainer() (*Container, error) {
	container := &Container{}
	if err := container.Initialize(); err != nil {
		return nil, err
	}
	return container, nil
}

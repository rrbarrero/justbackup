package module

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
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/websocket"
)

// HTTPServerModule manages the HTTP server
type HTTPServerModule struct {
	config  *config.ServerConfig
	handler http.Handler
	server  *http.Server
}

// NewHTTPServerModule creates a new HTTP server module
func NewHTTPServerModule(cfg *config.ServerConfig, handler http.Handler) *HTTPServerModule {
	return &HTTPServerModule{
		config:  cfg,
		handler: handler,
	}
}

// Name returns the module name
func (m *HTTPServerModule) Name() string {
	return "HTTPServer"
}

// Initialize sets up the HTTP server
func (m *HTTPServerModule) Initialize(cfg *config.ServerConfig) error {
	port := "8080"
	if cfg.ServerPort != "" {
		port = cfg.ServerPort
	}

	m.server = &http.Server{
		Addr:    ":" + port,
		Handler: m.handler,
	}

	log.Printf("HTTP Server module initialized on port %s", port)
	return nil
}

// Start starts the HTTP server
func (m *HTTPServerModule) Start(ctx context.Context) error {
	if m.server == nil {
		return fmt.Errorf("HTTP server not initialized")
	}

	go func() {
		log.Printf("Starting HTTP server on %s", m.server.Addr)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the HTTP server
func (m *HTTPServerModule) Stop() error {
	if m.server != nil {
		log.Println("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return m.server.Shutdown(ctx)
	}
	return nil
}

// RedisModule manages the Redis connection
type RedisModule struct {
	config *config.ServerConfig
	client *redis.Client
}

// NewRedisModule creates a new Redis module
func NewRedisModule(cfg *config.ServerConfig, client *redis.Client) *RedisModule {
	return &RedisModule{
		config: cfg,
		client: client,
	}
}

// Name returns the module name
func (m *RedisModule) Name() string {
	return "Redis"
}

// Initialize sets up the Redis connection
func (m *RedisModule) Initialize(cfg *config.ServerConfig) error {
	redisAddr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	m.client = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis module initialized and connected")
	return nil
}

// Start starts the Redis module (nothing needed here as connection is established in Initialize)
func (m *RedisModule) Start(ctx context.Context) error {
	log.Println("Redis module started")
	return nil
}

// Stop stops the Redis module
func (m *RedisModule) Stop() error {
	if m.client != nil {
		log.Println("Closing Redis connection...")
		return m.client.Close()
	}
	return nil
}

// EventBusModule manages the event bus
type EventBusModule struct {
	config   *config.ServerConfig
	redisMod *RedisModule
	eventBus *event.RedisEventBus
	listener *notifApp.NotificationEventListener
}

// NewEventBusModule creates a new event bus module
func NewEventBusModule(cfg *config.ServerConfig, redisMod *RedisModule, listener *notifApp.NotificationEventListener) *EventBusModule {
	return &EventBusModule{
		config:   cfg,
		redisMod: redisMod,
		listener: listener,
	}
}

// Name returns the module name
func (m *EventBusModule) Name() string {
	return "EventBus"
}

// Initialize sets up the event bus
func (m *EventBusModule) Initialize(cfg *config.ServerConfig) error {
	redisClient := m.redisMod.client
	if redisClient == nil {
		return fmt.Errorf("Redis client not available")
	}

	m.eventBus = event.NewRedisEventBus(redisClient)
	log.Println("EventBus module initialized")
	return nil
}

// Start starts the event bus listener
func (m *EventBusModule) Start(ctx context.Context) error {
	go m.listener.Start(ctx)
	log.Println("EventBus module started with notification listener")
	return nil
}

// Stop stops the event bus module
func (m *EventBusModule) Stop() error {
	log.Println("EventBus module stopped")
	return nil
}

// SchedulerModule manages the backup scheduler
type SchedulerModule struct {
	config          *config.ServerConfig
	backupScheduler *scheduler.Scheduler
	resultConsumer  *scheduler.ResultConsumer
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewSchedulerModule creates a new scheduler module
func NewSchedulerModule(cfg *config.ServerConfig, scheduler *scheduler.Scheduler, consumer *scheduler.ResultConsumer) *SchedulerModule {
	return &SchedulerModule{
		config:          cfg,
		backupScheduler: scheduler,
		resultConsumer:  consumer,
	}
}

// Name returns the module name
func (m *SchedulerModule) Name() string {
	return "Scheduler"
}

// Initialize sets up the scheduler
func (m *SchedulerModule) Initialize(cfg *config.ServerConfig) error {
	log.Println("Scheduler module initialized")
	return nil
}

// Start starts the scheduler and result consumer
func (m *SchedulerModule) Start(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)

	go m.backupScheduler.Start(m.ctx)
	go m.resultConsumer.Start(m.ctx)

	log.Println("Scheduler module started")
	return nil
}

// Stop stops the scheduler
func (m *SchedulerModule) Stop() error {
	if m.cancel != nil {
		m.cancel()
		log.Println("Scheduler module stopped")
	}
	return nil
}

// WebSocketModule manages the WebSocket hub
type WebSocketModule struct {
	config *config.ServerConfig
	hub    *websocket.Hub
}

// NewWebSocketModule creates a new WebSocket module
func NewWebSocketModule(cfg *config.ServerConfig, hub *websocket.Hub) *WebSocketModule {
	return &WebSocketModule{
		config: cfg,
		hub:    hub,
	}
}

// Name returns the module name
func (m *WebSocketModule) Name() string {
	return "WebSocket"
}

// Initialize sets up the WebSocket hub
func (m *WebSocketModule) Initialize(cfg *config.ServerConfig) error {
	log.Println("WebSocket module initialized")
	return nil
}

// Start starts the WebSocket hub
func (m *WebSocketModule) Start(ctx context.Context) error {
	go m.hub.Run()
	log.Println("WebSocket module started")
	return nil
}

// Stop stops the WebSocket module
func (m *WebSocketModule) Stop() error {
	log.Println("WebSocket module stopped")
	return nil
}

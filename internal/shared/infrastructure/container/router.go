package container

import (
	"log"
	"net/http"
	"strings"

	"github.com/rrbarrero/justbackup/internal/backup/interfaces/http/middleware"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/websocket"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	xwebsocket "golang.org/x/net/websocket"
)

// setupRouter configures all HTTP routes
func (c *Container) setupRouter(handlers *Handlers, services *Services, hub *websocket.Hub, cfg *config.ServerConfig) http.Handler {
	mainMux := http.NewServeMux()
	apiMux := http.NewServeMux()

	// Basic health check
	apiMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Public Routes
	apiMux.HandleFunc("/login", handlers.User.Login)
	apiMux.HandleFunc("/setup", handlers.User.Setup)
	apiMux.HandleFunc("/setup-status", handlers.User.GetSetupStatus)

	// Swagger UI route.
	// The Swagger documentation is integrated under the versioned API path.
	apiMux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// WebSocket Route
	apiMux.Handle("/ws", xwebsocket.Handler(hub.Handle))

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(services.JWT, services.Auth, services.User)
	protected := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authMiddleware.Handle(http.HandlerFunc(h)).ServeHTTP(w, r)
		}
	}

	// Register Routes
	handlers.Backup.RegisterRoutes(apiMux, protected)
	handlers.Host.RegisterRoutes(apiMux, protected)
	handlers.Settings.RegisterRoutes(apiMux, protected)
	handlers.Auth.RegisterRoutes(apiMux, protected)
	handlers.Notification.RegisterRoutes(apiMux, protected)
	handlers.System.RegisterRoutes(apiMux, protected)
	handlers.WorkerStats.RegisterRoutes(apiMux, protected)

	if handlers.Test != nil {
		log.Println("Registering test endpoints")
		handlers.Test.RegisterRoutes(apiMux, protected)
	}

	// Additional Routes
	apiMux.HandleFunc("POST /hosts/{id}/measure", protected(handlers.Backup.MeasureSize))
	apiMux.HandleFunc("GET /tasks/{id}", protected(handlers.Backup.GetTaskResult))
	apiMux.HandleFunc("/dashboard/stats", protected(handlers.Dashboard.GetStats))

	// Mount API Mux with versioning.
	// We use http.StripPrefix to ensure that versioned requests (e.g., /api/v1/backups)
	// are matched correctly by handlers registered without the version prefix.
	// This approach maintains compatibility with Go 1.22+ pattern matching standards.
	mainMux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiMux))

	// CORS
	allowedOriginsStr := cfg.CORSAllowedOrigin
	allowedOrigins := []string{}
	for _, origin := range strings.Split(allowedOriginsStr, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowedOrigins = append(allowedOrigins, trimmed)
		}
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	return corsMiddleware.Handler(mainMux)
}

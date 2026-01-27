package container

import (
	"fmt"
	"log"
	"os"

	authMem "github.com/rrbarrero/justbackup/internal/auth/infrastructure/persistence/memory"
	authPostgres "github.com/rrbarrero/justbackup/internal/auth/infrastructure/persistence/postgres"
	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/memory"
	"github.com/rrbarrero/justbackup/internal/backup/infrastructure/persistence/postgres"
	maintMem "github.com/rrbarrero/justbackup/internal/maintenance/infrastructure/persistence/memory"
	maintPostgres "github.com/rrbarrero/justbackup/internal/maintenance/infrastructure/persistence/postgres"
	notifMem "github.com/rrbarrero/justbackup/internal/notification/infrastructure/persistence/memory"
	notifPostgres "github.com/rrbarrero/justbackup/internal/notification/infrastructure/persistence/postgres"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/crypto"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/db"
	userMem "github.com/rrbarrero/justbackup/internal/user/infrastructure/persistence/memory"
	userPostgres "github.com/rrbarrero/justbackup/internal/user/infrastructure/persistence/postgres"
	workerStatsMem "github.com/rrbarrero/justbackup/internal/workerstats/infrastructure/persistence/memory"
)

// initializeRepositories initializes all repository implementations
func (c *Container) initializeRepositories(cfg *config.ServerConfig, env string) (*Repositories, error) {
	repos := &Repositories{}

	if env == "dev" || env == "development" {
		log.Println("Using In-Memory Repositories for development purposes")
		repos.Backup = memory.NewBackupRepositoryMemory()
		repos.Host = memory.NewHostRepositoryMemory()
		repos.Maintenance = maintMem.NewMaintenanceRepositoryMemory()
		repos.User = userMem.NewUserRepositoryMemory()
		repos.AuthToken = authMem.NewAuthTokenRepositoryMemory()
		repos.BackupError = memory.NewBackupErrorRepositoryMemory()
		repos.Notification = notifMem.NewNotificationRepositoryMemory()
		repos.WorkerStats = workerStatsMem.NewWorkerStatsRepositoryMemory()

		// Try to connect to DB for notifications if configured, otherwise use memory
		if os.Getenv("DB_HOST") != "" {
			dbConfig := db.Config{
				Host:     os.Getenv("DB_HOST"),
				Port:     os.Getenv("DB_PORT"),
				User:     os.Getenv("DB_USER"),
				Password: os.Getenv("DB_PASSWORD"),
				DBName:   os.Getenv("DB_NAME"),
			}
			conn, err := db.NewPostgresConnection(dbConfig)
			if err == nil {
				encryptionService, err := crypto.NewAESGCMEncryptionService(cfg.EncryptionKey)
				if err != nil {
					log.Printf("Failed to create encryption service in dev: %v", err)
				} else {
					log.Println("Using Postgres for Notifications in dev mode")
					repos.Notification = notifPostgres.NewNotificationRepositoryPostgres(conn, encryptionService)
				}
			} else {
				log.Printf("Could not connect to DB in dev mode for notifications: %v", err)
			}
		}
	} else {
		log.Println("Using Postgres Repositories")
		dbConfig := db.Config{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   os.Getenv("DB_NAME"),
		}
		conn, err := db.NewPostgresConnection(dbConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}

		// Run Migrations
		log.Println("Running database migrations...")
		if err := db.RunMigrations(dbConfig, "migrations"); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}

		encryptionService, err := crypto.NewAESGCMEncryptionService(cfg.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create encryption service: %w", err)
		}

		repos.Backup = postgres.NewBackupRepositoryPostgres(conn, encryptionService)
		repos.Host = postgres.NewHostRepositoryPostgres(conn)
		repos.User = userPostgres.NewUserRepositoryPostgres(conn)
		repos.AuthToken = authPostgres.NewAuthTokenRepositoryPostgres(conn)
		repos.BackupError = postgres.NewBackupErrorRepositoryPostgres(conn)
		repos.Maintenance = maintPostgres.NewMaintenanceRepositoryPostgres(conn)
		repos.Notification = notifPostgres.NewNotificationRepositoryPostgres(conn, encryptionService)
		repos.WorkerStats = workerStatsMem.NewWorkerStatsRepositoryMemory()
	}

	return repos, nil
}

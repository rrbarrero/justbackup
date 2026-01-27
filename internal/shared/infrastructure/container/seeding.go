package container

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	domain "github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/config"
)

// SeedSelfBackup ensures the existence of the internal "Self Backup" task
// It connects to the database via the repositories to check/create the host and task.
func SeedSelfBackup(hostRepo interfaces.HostRepository, backupRepo interfaces.BackupRepository, cfg *config.ServerConfig) error {
	ctx := context.Background()

	// 1. Ensure "Internal Host" exists
	hosts, err := hostRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list hosts: %w", err)
	}

	var internalHost *domain.Host
	for _, h := range hosts {
		if h.Name() == "Internal System" {
			internalHost = h
			break
		}
	}

	if internalHost == nil {
		log.Println("Seeding: Creating 'Internal System' host...")
		newHost := domain.NewHost(
			"Internal System",
			"127.0.0.1", // Loopback
			"backup",    // Dummy user
			22,          // Dummy port
			"",          // Path
			false,       // isWorkstation
		)
		if err := hostRepo.Save(ctx, newHost); err != nil {
			return fmt.Errorf("failed to save internal host: %w", err)
		}
		internalHost = newHost
	}

	// 2. Ensure "Self Backup" task exists for this host
	backups, err := backupRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	for _, b := range backups {
		if b.HostID() == internalHost.ID() && b.Destination() == "internal_db_backups" {
			// Task already exists
			log.Println("Seeding: 'Self Backup' task already exists.")
			return nil
		}
	}

	log.Println("Seeding: Creating 'Self Backup' task...")

	// Gather connection details from Environment (as seen by the Server)
	// We pass these as Hook Params.
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	hookParams := map[string]string{
		"host":     dbHost,
		"port":     dbPort,
		"user":     dbUser,
		"password": dbPassword,
		"db_name":  dbName,
	}

	schedule := domain.NewBackupSchedule("@daily")

	newBackup, err := domain.NewBackup(
		internalHost.ID(),
		"{{SESSION_TEMP_DIR}}", // Ephemeral source
		"internal_db_backups",  // Destination
		schedule,
		[]string{}, // Excludes
		false,      // incremental (full dump)
		30,         // Retention
		true,       // Encrypted
	)
	if err != nil {
		return fmt.Errorf("failed to create initial self-backup: %w", err)
	}

	// Create Hook
	// Convert BackupID (string wrapper) to UUID for Hook
	backupIDUuid, err := uuid.Parse(newBackup.ID().String())
	if err != nil {
		return fmt.Errorf("failed to parse new backup ID as UUID: %w", err)
	}

	hook := domain.NewBackupHook(
		backupIDUuid,
		"postgres_dump",
		domain.HookPhasePre,
		hookParams,
	)

	newBackup.AddHook(hook)

	if err := backupRepo.Save(ctx, newBackup); err != nil {
		return fmt.Errorf("failed to save self backup: %w", err)
	}

	log.Println("Seeding: 'Self Backup' task created successfully.")
	return nil
}

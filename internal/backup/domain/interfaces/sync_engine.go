package interfaces

import "context"

// SyncEngine defines the interface for performing file synchronization.
type SyncEngine interface {
	// Sync performs the synchronization from source to destination.
	// It returns an error if the synchronization fails.
	Sync(ctx context.Context, source, destination string, excludes []string, dryRun bool) error
}

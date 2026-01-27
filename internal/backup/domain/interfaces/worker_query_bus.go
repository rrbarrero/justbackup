package interfaces

import (
	"context"

	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
)

// WorkerQueryBus defines the interface for querying workers synchronously (request-response over messaging)
type WorkerQueryBus interface {
	SearchFiles(ctx context.Context, pattern string) (workerDto.SearchFilesResult, error)
	ListFiles(ctx context.Context, path string) (workerDto.ListFilesResult, error)
}

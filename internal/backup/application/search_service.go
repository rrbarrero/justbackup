package application

import (
	"context"
	"strings"

	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
)

type BackupSearchService struct {
	repo        interfaces.BackupRepository
	hostService *HostService
	queryBus    interfaces.WorkerQueryBus
	assembler   *assembler.BackupAssembler
}

func NewBackupSearchService(
	repo interfaces.BackupRepository,
	hostService *HostService,
	queryBus interfaces.WorkerQueryBus,
	assembler *assembler.BackupAssembler,
) *BackupSearchService {
	return &BackupSearchService{
		repo:        repo,
		hostService: hostService,
		queryBus:    queryBus,
		assembler:   assembler,
	}
}

const baseMountPoint = "/mnt/backups"

func (s *BackupSearchService) SearchFiles(ctx context.Context, pattern string) ([]*dto.FileSearchResult, error) {
	searchResult, err := s.queryBus.SearchFiles(ctx, pattern)
	if err != nil {
		return nil, err
	}

	if len(searchResult.Files) == 0 {
		return []*dto.FileSearchResult{}, nil
	}

	metadata, err := s.loadBackupMetadata(ctx)
	if err != nil {
		return nil, err
	}

	return s.matchFilesToBackups(searchResult.Files, metadata), nil
}

func (s *BackupSearchService) ListFiles(ctx context.Context, backupID string, path string) ([]*dto.BackupFileResponse, error) {
	bid, err := valueobjects.NewBackupIDFromString(backupID)
	if err != nil {
		return nil, err
	}

	backup, err := s.repo.FindByID(ctx, bid)
	if err != nil {
		return nil, err
	}

	hostResp, err := s.hostService.GetHost(ctx, backup.HostID().String())
	if err != nil {
		return nil, err
	}

	fullPath := s.computeFullBackupPath(hostResp.Path, backup.Destination())
	if path != "" {
		fullPath = fullPath + "/" + strings.TrimPrefix(path, "/")
	}

	listResult, err := s.queryBus.ListFiles(ctx, fullPath)
	if err != nil {
		return nil, err
	}

	var files []*dto.BackupFileResponse
	for _, f := range listResult.Files {
		files = append(files, &dto.BackupFileResponse{
			Name:  f.Name,
			IsDir: f.IsDir,
			Size:  f.Size,
		})
	}

	return files, nil
}

type backupMetadata struct {
	Backup *entities.Backup
	Host   *dto.HostResponse
	Prefix string
}

func (s *BackupSearchService) loadBackupMetadata(ctx context.Context) ([]backupMetadata, error) {
	backups, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	hostIDs := make([]entities.HostID, 0)
	seenHosts := make(map[string]bool)
	for _, b := range backups {
		if !seenHosts[b.HostID().String()] {
			hostIDs = append(hostIDs, b.HostID())
			seenHosts[b.HostID().String()] = true
		}
	}

	hostMap, err := s.hostService.GetHostsByIDs(ctx, hostIDs)
	if err != nil {
		return nil, err
	}

	var meta []backupMetadata
	for _, b := range backups {
		if h, ok := hostMap[b.HostID().String()]; ok {
			prefix := s.computeFullBackupPath(h.Path, b.Destination())
			meta = append(meta, backupMetadata{
				Backup: b,
				Host:   h,
				Prefix: prefix,
			})
		}
	}
	return meta, nil
}

func (s *BackupSearchService) matchFilesToBackups(files []string, metadata []backupMetadata) []*dto.FileSearchResult {
	var results []*dto.FileSearchResult
	for _, filePath := range files {
		var matched backupMetadata

		// Find the matching backup using prefix matching
		for _, m := range metadata {
			if filePath == m.Prefix || strings.HasPrefix(filePath, m.Prefix+"/") {
				matched = m
				break
			}
		}

		res := &dto.FileSearchResult{
			Path: filePath,
		}

		if matched.Backup != nil {
			res.Backup = s.assembler.ToBackupResponse(matched.Backup, matched.Host.Name, matched.Host.Hostname)
		}

		results = append(results, res)
	}
	return results
}

func (s *BackupSearchService) computeFullBackupPath(hostPath, destination string) string {
	return baseMountPoint + "/" + strings.Trim(hostPath, "/") + "/" + strings.Trim(destination, "/")
}

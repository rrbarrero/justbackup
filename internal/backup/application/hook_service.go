package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rrbarrero/justbackup/internal/backup/application/assembler"
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
	"github.com/rrbarrero/justbackup/internal/backup/domain/interfaces"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
)

type BackupHookService struct {
	repo      interfaces.BackupRepository
	assembler *assembler.BackupAssembler
}

func NewBackupHookService(repo interfaces.BackupRepository, assembler *assembler.BackupAssembler) *BackupHookService {
	return &BackupHookService{
		repo:      repo,
		assembler: assembler,
	}
}

func (s *BackupHookService) CreateHook(ctx context.Context, backupID string, req dto.CreateHookRequest) (*dto.HookDTO, error) {
	bid, err := valueobjects.NewBackupIDFromString(backupID)
	if err != nil {
		return nil, err
	}

	backup, err := s.repo.FindByID(ctx, bid)
	if err != nil {
		return nil, err
	}

	hookUUID, _ := uuid.Parse(bid.String())
	hook := entities.NewBackupHook(hookUUID, req.Name, entities.HookPhase(req.Phase), req.Params)
	hook.Enabled = req.Enabled
	backup.AddHook(hook)

	if err := s.repo.Save(ctx, backup); err != nil {
		return nil, err
	}

	return s.assembler.ToHookDTO(hook), nil
}

func (s *BackupHookService) UpdateHook(ctx context.Context, backupID, hookID string, req dto.UpdateHookRequest) (*dto.HookDTO, error) {
	bid, err := valueobjects.NewBackupIDFromString(backupID)
	if err != nil {
		return nil, err
	}

	backup, err := s.repo.FindByID(ctx, bid)
	if err != nil {
		return nil, err
	}

	var targetHook *entities.BackupHook
	for _, h := range backup.Hooks() {
		if h.ID.String() == hookID {
			targetHook = h
			break
		}
	}

	if targetHook == nil {
		return nil, fmt.Errorf("hook not found")
	}

	if req.Name != "" {
		targetHook.Name = req.Name
	}
	if req.Phase != "" {
		targetHook.Phase = entities.HookPhase(req.Phase)
	}
	if req.Params != nil {
		targetHook.Params = req.Params
	}
	if req.Enabled != nil {
		targetHook.Enabled = *req.Enabled
	}
	targetHook.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, backup); err != nil {
		return nil, err
	}

	return s.assembler.ToHookDTO(targetHook), nil
}

func (s *BackupHookService) DeleteHook(ctx context.Context, backupID, hookID string) error {
	bid, err := valueobjects.NewBackupIDFromString(backupID)
	if err != nil {
		return err
	}

	backup, err := s.repo.FindByID(ctx, bid)
	if err != nil {
		return err
	}

	hooks := backup.Hooks()
	found := false
	newHooks := make([]*entities.BackupHook, 0, len(hooks))
	for _, h := range hooks {
		if h.ID.String() == hookID {
			found = true
			continue
		}
		newHooks = append(newHooks, h)
	}

	if !found {
		return fmt.Errorf("hook not found")
	}

	backup.SetHooks(newHooks)
	return s.repo.Save(ctx, backup)
}

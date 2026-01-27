package entities

import (
	"time"

	"github.com/google/uuid"
)

type HookPhase string

const (
	HookPhasePre  HookPhase = "pre"
	HookPhasePost HookPhase = "post"
)

type BackupHook struct {
	ID        uuid.UUID
	BackupID  uuid.UUID
	Name      string
	Phase     HookPhase
	Enabled   bool
	Params    map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewBackupHook(backupID uuid.UUID, name string, phase HookPhase, params map[string]string) *BackupHook {
	now := time.Now()
	return &BackupHook{
		ID:        uuid.New(),
		BackupID:  backupID,
		Name:      name,
		Phase:     phase,
		Enabled:   true,
		Params:    params,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

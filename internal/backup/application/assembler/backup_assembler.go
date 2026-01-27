package assembler

import (
	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/backup/domain/entities"
)

type BackupAssembler struct{}

func NewBackupAssembler() *BackupAssembler {
	return &BackupAssembler{}
}

func (a *BackupAssembler) ToBackupResponse(backup *entities.Backup, hostName, hostAddress string) *dto.BackupResponse {
	return &dto.BackupResponse{
		ID:          backup.ID().String(),
		HostID:      backup.HostID().String(),
		HostName:    hostName,
		HostAddress: hostAddress,
		Path:        backup.Path(),
		Destination: backup.Destination(),
		Status:      backup.Status().String(),
		Schedule:    backup.Schedule().CronExpression,
		LastRun:     backup.Schedule().LastRun,
		Excludes:    backup.Excludes(),
		Incremental: backup.Incremental(),
		Size:        backup.Size(),
		Retention:   backup.Retention(),
		Encrypted:   backup.Encrypted(),
		Hooks:       a.ToHookDTOs(backup.Hooks()),
	}
}

func (a *BackupAssembler) ToHookDTOs(hooks []*entities.BackupHook) []dto.HookDTO {
	res := make([]dto.HookDTO, 0, len(hooks))
	for _, h := range hooks {
		res = append(res, dto.HookDTO{
			ID:      h.ID.String(),
			Name:    h.Name,
			Phase:   string(h.Phase),
			Params:  h.Params,
			Enabled: h.Enabled,
		})
	}
	return res
}

func (a *BackupAssembler) ToHookDTO(h *entities.BackupHook) *dto.HookDTO {
	return &dto.HookDTO{
		ID:      h.ID.String(),
		Name:    h.Name,
		Phase:   string(h.Phase),
		Params:  h.Params,
		Enabled: h.Enabled,
	}
}

func (a *BackupAssembler) ToBackupErrorResponse(e *entities.BackupError) *dto.BackupErrorResponse {
	return &dto.BackupErrorResponse{
		ID:           e.ID,
		JobID:        e.JobID,
		BackupID:     e.BackupID.String(),
		OccurredAt:   e.OccurredAt,
		ErrorMessage: e.ErrorMessage,
	}
}

func (a *BackupAssembler) ToBackupErrorResponses(errors []*entities.BackupError) []*dto.BackupErrorResponse {
	responses := make([]*dto.BackupErrorResponse, 0, len(errors))
	for _, e := range errors {
		responses = append(responses, a.ToBackupErrorResponse(e))
	}
	return responses
}

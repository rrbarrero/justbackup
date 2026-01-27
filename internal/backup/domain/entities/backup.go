package entities

import (
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rrbarrero/justbackup/internal/backup/domain/valueobjects"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

// NowFunc is a variable that holds the current time function.
// It can be overridden in tests for deterministic time.
var NowFunc = time.Now

type Backup struct {
	shared.AggregateRoot
	id          valueobjects.BackupID
	hostID      HostID
	path        string
	destination string
	status      valueobjects.BackupStatus
	schedule    BackupSchedule
	createdAt   time.Time
	updatedAt   time.Time
	nextRunAt   *time.Time
	excludes    []string
	enabled     bool
	incremental bool
	size        string
	retention   int
	encrypted   bool
	hooks       []*BackupHook
}

func NewBackup(hostID HostID, path, destination string, schedule BackupSchedule, excludes []string, incremental bool, retention int, encrypted bool) (*Backup, error) {
	currentTime := NowFunc()
	b := &Backup{
		id:          valueobjects.NewBackupID(),
		hostID:      hostID,
		path:        path,
		destination: destination,
		status:      valueobjects.BackupStatusPending,
		schedule:    schedule,
		createdAt:   currentTime,
		updatedAt:   currentTime,
		excludes:    sanitizeExcludes(excludes),
		enabled:     true,
		incremental: incremental,
		size:        "",
		retention:   retention,
		encrypted:   encrypted,
		hooks:       []*BackupHook{},
	}
	if err := b.CalculateNextRun(); err != nil {
		return nil, err
	}
	return b, nil
}

func NewBackupWithID(id valueobjects.BackupID, hostID HostID, path, destination string, schedule BackupSchedule, excludes []string, incremental bool, retention int, encrypted bool) (*Backup, error) {
	currentTime := NowFunc()
	b := &Backup{
		id:          id,
		hostID:      hostID,
		path:        path,
		destination: destination,
		status:      valueobjects.BackupStatusPending,
		schedule:    schedule,
		createdAt:   currentTime,
		updatedAt:   currentTime,
		excludes:    sanitizeExcludes(excludes),
		enabled:     true,
		incremental: incremental,
		size:        "",
		retention:   retention,
		encrypted:   encrypted,
		hooks:       []*BackupHook{},
	}
	if err := b.CalculateNextRun(); err != nil {
		return nil, err
	}
	return b, nil
}

func RestoreBackup(id valueobjects.BackupID, hostID HostID, path, destination string, status valueobjects.BackupStatus, schedule BackupSchedule, createdAt, updatedAt time.Time, nextRunAt *time.Time, excludes []string, enabled bool, incremental bool, size string, retention int, encrypted bool) *Backup {
	return &Backup{
		id:          id,
		hostID:      hostID,
		path:        path,
		destination: destination,
		status:      status,
		schedule:    schedule,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		nextRunAt:   nextRunAt,
		excludes:    sanitizeExcludes(excludes),
		enabled:     enabled,
		incremental: incremental,
		size:        size,
		retention:   retention,
		encrypted:   encrypted,
		hooks:       []*BackupHook{},
	}
}

func (b *Backup) ID() valueobjects.BackupID {
	return b.id
}

func (b *Backup) HostID() HostID {
	return b.hostID
}

func (b *Backup) Path() string {
	return b.path
}

func (b *Backup) Destination() string {
	return b.destination
}

func (b *Backup) Status() valueobjects.BackupStatus {
	return b.status
}

func (b *Backup) Schedule() BackupSchedule {
	return b.schedule
}

func (b *Backup) NextRunAt() *time.Time {
	return b.nextRunAt
}

func (b *Backup) Enabled() bool {
	return b.enabled
}

func (b *Backup) Excludes() []string {
	if len(b.excludes) == 0 {
		return []string{}
	}
	result := make([]string, len(b.excludes))
	copy(result, b.excludes)
	return result
}

func (b *Backup) Enable() {
	b.enabled = true
	b.updatedAt = NowFunc()
	b.CalculateNextRun()
}

func (b *Backup) Disable() {
	b.enabled = false
	b.updatedAt = NowFunc()
	b.CalculateNextRun()
}

func (b *Backup) CreatedAt() time.Time {
	return b.createdAt
}

func (b *Backup) UpdatedAt() time.Time {
	return b.updatedAt
}

func (b *Backup) CalculateNextRun() error {
	if !b.enabled {
		b.nextRunAt = nil
		return nil
	}

	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(b.schedule.CronExpression)
	if err != nil {
		return err
	}

	next := schedule.Next(NowFunc())
	b.nextRunAt = &next
	return nil
}

func (b *Backup) Start() error {
	if b.status == valueobjects.BackupStatusRunning {
		return nil // Already running
	}
	b.status = valueobjects.BackupStatusRunning
	b.updatedAt = NowFunc() // Use NowFunc
	// Record event: BackupStarted
	return nil
}

func (b *Backup) Complete() {
	b.status = valueobjects.BackupStatusCompleted
	b.updatedAt = NowFunc()        // Use NowFunc
	b.schedule.LastRun = NowFunc() // Use NowFunc
	b.CalculateNextRun()           // Recalculate next run
	// Record event: BackupCompleted
}

func (b *Backup) Fail() {
	b.status = valueobjects.BackupStatusFailed
	b.updatedAt = NowFunc() // Use NowFunc
	// Record event: BackupFailed
}

func (b *Backup) Incremental() bool {
	return b.incremental
}

func (b *Backup) Size() string {
	return b.size
}

func (b *Backup) SetSize(size string) {
	b.size = size
}

func (b *Backup) Retention() int {
	return b.retention
}

func (b *Backup) Encrypted() bool {
	return b.encrypted
}

func (b *Backup) Update(path, destination string, schedule BackupSchedule, excludes []string, incremental bool, retention int, encrypted bool) error {
	b.path = path
	b.destination = destination
	b.schedule = schedule
	b.excludes = sanitizeExcludes(excludes)
	b.incremental = incremental
	b.retention = retention
	b.encrypted = encrypted
	b.updatedAt = NowFunc()
	return b.CalculateNextRun()
}

func (b *Backup) Hooks() []*BackupHook {
	if b.hooks == nil {
		return []*BackupHook{}
	}
	return b.hooks
}

func (b *Backup) SetHooks(hooks []*BackupHook) {
	b.hooks = hooks
}

func (b *Backup) AddHook(hook *BackupHook) {
	b.hooks = append(b.hooks, hook)
}

func sanitizeExcludes(excludes []string) []string {
	clean := make([]string, 0, len(excludes))
	for _, ex := range excludes {
		trimmed := strings.TrimSpace(ex)
		if trimmed == "" {
			continue
		}
		clean = append(clean, trimmed)
	}
	return clean
}

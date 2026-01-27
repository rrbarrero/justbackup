package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type MaintenanceTaskType string

const (
	MaintenanceTaskTypePurge MaintenanceTaskType = "purge"
)

type MaintenanceTask struct {
	id        uuid.UUID
	name      string
	taskType  MaintenanceTaskType
	schedule  string
	nextRunAt *time.Time
	lastRunAt *time.Time
	enabled   bool
	createdAt time.Time
	updatedAt time.Time
}

func NewMaintenanceTask(name string, taskType MaintenanceTaskType, schedule string) (*MaintenanceTask, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if schedule == "" {
		return nil, fmt.Errorf("schedule is required")
	}

	task := &MaintenanceTask{
		id:        uuid.New(),
		name:      name,
		taskType:  taskType,
		schedule:  schedule,
		enabled:   true,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}

	if err := task.CalculateNextRun(); err != nil {
		return nil, err
	}

	return task, nil
}

func RestoreMaintenanceTask(id uuid.UUID, name string, taskType MaintenanceTaskType, schedule string, nextRunAt, lastRunAt *time.Time, enabled bool, createdAt, updatedAt time.Time) *MaintenanceTask {
	return &MaintenanceTask{
		id:        id,
		name:      name,
		taskType:  taskType,
		schedule:  schedule,
		nextRunAt: nextRunAt,
		lastRunAt: lastRunAt,
		enabled:   enabled,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (t *MaintenanceTask) ID() uuid.UUID             { return t.id }
func (t *MaintenanceTask) Name() string              { return t.name }
func (t *MaintenanceTask) Type() MaintenanceTaskType { return t.taskType }
func (t *MaintenanceTask) Schedule() string          { return t.schedule }
func (t *MaintenanceTask) NextRunAt() *time.Time     { return t.nextRunAt }
func (t *MaintenanceTask) LastRunAt() *time.Time     { return t.lastRunAt }
func (t *MaintenanceTask) Enabled() bool             { return t.enabled }
func (t *MaintenanceTask) CreatedAt() time.Time      { return t.createdAt }
func (t *MaintenanceTask) UpdatedAt() time.Time      { return t.updatedAt }

func (t *MaintenanceTask) CalculateNextRun() error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	s, err := parser.Parse(t.schedule)
	if err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}

	next := s.Next(time.Now())
	t.nextRunAt = &next
	return nil
}

func (t *MaintenanceTask) SetLastRun(at time.Time) {
	t.lastRunAt = &at
	t.updatedAt = time.Now()
}

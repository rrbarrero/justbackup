package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBackupSchedule(t *testing.T) {
	t.Run("creates a schedule with the given cron expression", func(t *testing.T) {
		testCases := []string{
			"0 0 * * *",
			"0 0 * * 1-5",
			"0 0 1,15 * *",
			"*/15 * * * *",
			"@daily",
			"@weekly",
		}

		for _, cronExpr := range testCases {
			t.Run(cronExpr, func(t *testing.T) {
				schedule := NewBackupSchedule(cronExpr)
				assert.Equal(t, cronExpr, schedule.CronExpression)
			})
		}
	})

	t.Run("initial LastRun and NextRun should be zero", func(t *testing.T) {
		schedule := NewBackupSchedule("any cron")
		assert.True(t, schedule.LastRun.IsZero())
		assert.True(t, schedule.NextRun.IsZero())
	})
}

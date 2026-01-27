package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectBackupsToPurge(t *testing.T) {
	tests := []struct {
		name      string
		backups   []string // Assumed sorted oldest first
		retention int
		expected  []string
	}{
		{
			name:      "No backups, retention 5",
			backups:   []string{},
			retention: 5,
			expected:  []string{},
		},
		{
			name:      "3 backups, retention 5",
			backups:   []string{"2023-01-01", "2023-01-02", "2023-01-03"},
			retention: 5,
			expected:  []string{},
		},
		{
			name:      "5 backups, retention 5",
			backups:   []string{"2023-01-01", "2023-01-02", "2023-01-03", "2023-01-04", "2023-01-05"},
			retention: 5,
			expected:  []string{},
		},
		{
			name:      "6 backups, retention 5 (Purge oldest)",
			backups:   []string{"2023-01-01", "2023-01-02", "2023-01-03", "2023-01-04", "2023-01-05", "2023-01-06"},
			retention: 5,
			expected:  []string{"2023-01-01"},
		},
		{
			name:      "10 backups, retention 3",
			backups:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			retention: 3,
			expected:  []string{"1", "2", "3", "4", "5", "6", "7"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectBackupsToPurge(tt.backups, tt.retention)
			assert.Equal(t, tt.expected, result)
		})
	}
}

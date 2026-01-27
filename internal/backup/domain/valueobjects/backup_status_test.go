package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBackupStatus(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expected      BackupStatus
		expectedError error
	}{
		{
			name:          "valid pending status",
			input:         "pending",
			expected:      BackupStatusPending,
			expectedError: nil,
		},
		{
			name:          "valid running status",
			input:         "running",
			expected:      BackupStatusRunning,
			expectedError: nil,
		},
		{
			name:          "valid completed status",
			input:         "completed",
			expected:      BackupStatusCompleted,
			expectedError: nil,
		},
		{
			name:          "valid failed status",
			input:         "failed",
			expected:      BackupStatusFailed,
			expectedError: nil,
		},
		{
			name:          "invalid status",
			input:         "unknown",
			expected:      "",
			expectedError: ErrInvalidStatus,
		},
		{
			name:          "empty status",
			input:         "",
			expected:      "",
			expectedError: ErrInvalidStatus,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, err := NewBackupStatus(tc.input)
			assert.Equal(t, tc.expected, status)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestBackupStatus_String(t *testing.T) {
	testCases := []struct {
		name     string
		status   BackupStatus
		expected string
	}{
		{
			name:     "pending status string",
			status:   BackupStatusPending,
			expected: "pending",
		},
		{
			name:     "running status string",
			status:   BackupStatusRunning,
			expected: "running",
		},
		{
			name:     "completed status string",
			status:   BackupStatusCompleted,
			expected: "completed",
		},
		{
			name:     "failed status string",
			status:   BackupStatusFailed,
			expected: "failed",
		},
		{
			name:     "empty status string",
			status:   "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.status.String())
		})
	}
}

package application

import (
	"fmt"
	"testing"

	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
	"github.com/stretchr/testify/assert"
)

func TestBuildRsyncArgs(t *testing.T) {
	sshKeyPath := "/path/to/key"
	task := workerDto.WorkerTask{
		Host: "example.com",
		User: "user",
		Port: 2222,
	}
	source := "/source"
	finalDest := "/dest"

	t.Run("Basic arguments", func(t *testing.T) {
		args := BuildRsyncArgs(sshKeyPath, task, false, []string{}, source, finalDest)

		expectedSSHOpts := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -p %d", sshKeyPath, task.Port)
		assert.Contains(t, args, "-az")
		assert.Contains(t, args, "--no-owner")
		assert.Contains(t, args, "--no-group")
		assert.Contains(t, args, "--numeric-ids")
		assert.Contains(t, args, "-e")
		assert.Contains(t, args, expectedSSHOpts)
		assert.Equal(t, source, args[len(args)-2])
		assert.Equal(t, finalDest, args[len(args)-1])
	})

	t.Run("With excludes", func(t *testing.T) {
		excludes := []string{"--exclude=*.log", "--exclude=tmp/"}
		args := BuildRsyncArgs(sshKeyPath, task, false, excludes, source, finalDest)

		assert.Contains(t, args, "--exclude=*.log")
		assert.Contains(t, args, "--exclude=tmp/")
	})

	t.Run("Incremental with link dest", func(t *testing.T) {
		args := BuildRsyncArgs(sshKeyPath, task, true, []string{}, source, finalDest)

		assert.Contains(t, args, "--link-dest=../latest")
	})

	t.Run("Incremental without link dest", func(t *testing.T) {
		args := BuildRsyncArgs(sshKeyPath, task, false, []string{}, source, finalDest)

		assert.NotContains(t, args, "--link-dest=../latest")
	})
}

func TestValidateHookPath(t *testing.T) {
	pluginDir := "/app/plugins"

	tests := []struct {
		name      string
		hookName  string
		expectErr bool
	}{
		{
			name:      "Valid hook name",
			hookName:  "my-hook",
			expectErr: false,
		},
		{
			name:      "Valid nested hook",
			hookName:  "subdir/hook",
			expectErr: false,
		},
		{
			name:      "Path traversal attempt",
			hookName:  "../../etc/passwd",
			expectErr: true,
		},
		{
			name:      "Path traversal with parent",
			hookName:  "../outside",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHookPath(pluginDir, tt.hookName)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

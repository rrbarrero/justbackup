package application

import (
	"fmt"
	"testing"

	workerDto "github.com/rrbarrero/justbackup/internal/worker/dto"
	"github.com/stretchr/testify/assert"
)

func TestSetupEphemeralWorkspace(t *testing.T) {
	t.Run("Standard Backup Path", func(t *testing.T) {
		task := workerDto.WorkerTask{
			Path: "/var/www/html",
		}

		sourceOverride, sessionDir, cleanup, err := setupEphemeralWorkspace(task)

		assert.NoError(t, err)
		assert.Equal(t, "", sourceOverride, "Should return empty override for standard paths to trigger remote rsync")
		assert.Equal(t, "", sessionDir)
		assert.Nil(t, cleanup)
	})

	t.Run("Ephemeral Session Path", func(t *testing.T) {
		task := workerDto.WorkerTask{
			Path: "{{SESSION_TEMP_DIR}}",
		}

		sourceOverride, sessionDir, cleanup, err := setupEphemeralWorkspace(task)

		assert.NoError(t, err)
		assert.NotEmpty(t, sourceOverride, "Should return path override for ephemeral session")
		assert.NotEmpty(t, sessionDir)
		assert.NotNil(t, cleanup)

		// Cleanup logic test
		cleanup()
	})
}

func TestExecuteRsyncOperation_Logic(t *testing.T) {
	// This test validates the command construction logic without executing rsync
	// We are testing internal logic via BuildRsyncArgs since executeRsyncOperation does side effects

	sshKey := "/keys/id_rsa"
	finalDest := "/mnt/backups/target"

	t.Run("Remote Source Construction", func(t *testing.T) {
		task := workerDto.WorkerTask{
			User: "root",
			Host: "example.com",
			Path: "/remote/path",
			Port: 22,
		}

		// Emulate behavior when sourceOverride is empty (standard case)
		rsyncSource := fmt.Sprintf("%s@%s:%s", task.User, task.Host, task.Path)
		args := BuildRsyncArgs(sshKey, task, false, nil, rsyncSource, finalDest)

		assert.Contains(t, args, "root@example.com:/remote/path")
		assert.NotContains(t, args, "/remote/path/")
	})

	t.Run("Local Ephemeral Source Construction", func(t *testing.T) {
		task := workerDto.WorkerTask{
			Path: "{{SESSION_TEMP_DIR}}", // Original task path
		}

		// Emulate behavior when sourceOverride is set (ephemeral case)
		localSource := "/tmp/session-123"
		rsyncSource := localSource + "/"

		args := BuildRsyncArgs(sshKey, task, false, nil, rsyncSource, finalDest)

		assert.Contains(t, args, "/tmp/session-123/")
	})
}

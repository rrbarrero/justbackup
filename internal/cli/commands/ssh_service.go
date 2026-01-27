package commands

import (
	"bytes"
	"fmt"

	"golang.org/x/crypto/ssh"
)

type sshServiceImpl struct{}

func NewSSHService() SSHService {
	return &sshServiceImpl{}
}

func (s *sshServiceImpl) InstallKey(host string, port int, user string, password string, publicKey string) error {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Bypassing strict host key verification for the initial bootstrap process.
		Timeout:         0,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer func() { _ = session.Close() }()

	// Command to append key to authorized_keys, ensuring permissions are correct and restricting capabilities
	// valid commands are: "rsync --server" (for backups) and "du -sk" (for size measurement)
	wrapper := `case "$SSH_ORIGINAL_COMMAND" in rsync\ --server*|du\ -sk*) $SSH_ORIGINAL_COMMAND ;; *) echo "Access Denied by JustBackup"; exit 1 ;; esac`
	options := fmt.Sprintf(`command="%s",no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty`, wrapper)

	installCmd := fmt.Sprintf(`mkdir -p ~/.ssh && echo "%s %s" >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && chmod 700 ~/.ssh`, options, publicKey)

	var stderr bytes.Buffer
	session.Stderr = &stderr
	if err := session.Run(installCmd); err != nil {
		return fmt.Errorf("failed to install key: %v (stderr: %s)", err, stderr.String())
	}

	return nil
}

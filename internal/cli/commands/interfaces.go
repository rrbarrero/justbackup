package commands

import (
	"io"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
)

// SSHService defines the interface for SSH operations.
type SSHService interface {
	InstallKey(host string, port int, user string, password string, publicKey string) error
}

// APIService defines the interface for API operations required by CLI commands.
type APIService interface {
	GetSSHKey() (string, error)
	RegisterHost(req dto.CreateHostRequest) error
	RequestRestore(backupID string, req dto.RestoreRequest) (string, error)
}

// NetService defines the interface for network and data streaming operations.
type NetService interface {
	GetLocalIP() string
	ListenTCP() (string, int, io.Closer, error)
	AcceptAndValidate(listener io.Closer, token string) (io.ReadCloser, error)
	ExtractTarGz(r io.Reader, dest string) error
}

// OSUserRetriever defines the interface for retrieving the current OS user.
type OSUserRetriever interface {
	GetCurrentUsername() (string, error)
}

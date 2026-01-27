package commands

import (
	"os/user"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/shared/utils"
)

type BootstrapService struct {
	sshService SSHService
	apiService APIService
}

func NewBootstrapService(ssh SSHService, api APIService) *BootstrapService {
	return &BootstrapService{
		sshService: ssh,
		apiService: api,
	}
}

type BootstrapParams struct {
	Name     string
	Host     string
	User     string
	Port     int
	Password string
}

func (s *BootstrapService) Execute(params BootstrapParams) error {
	// 1. Fetch SSH Public Key from API
	publicKey, err := s.apiService.GetSSHKey()
	if err != nil {
		return err
	}

	// 2. Connect via SSH and Install Key
	err = s.sshService.InstallKey(params.Host, params.Port, params.User, params.Password, publicKey)
	if err != nil {
		return err
	}

	// 3. Register Host via API
	path := utils.Slugify(params.Name)
	createHostReq := dto.CreateHostRequest{
		Name:          params.Name,
		Hostname:      params.Host,
		User:          params.User,
		Port:          params.Port,
		Path:          path,
		IsWorkstation: false,
	}

	return s.apiService.RegisterHost(createHostReq)
}

type osUserRetrieverImpl struct{}

func NewOSUserRetriever() OSUserRetriever {
	return &osUserRetrieverImpl{}
}

func (r *osUserRetrieverImpl) GetCurrentUsername() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Username, nil
}

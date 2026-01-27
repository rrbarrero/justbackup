package commands

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
)

type RestoreService struct {
	apiService APIService
	netService NetService
}

func NewRestoreService(api APIService, net NetService) *RestoreService {
	return &RestoreService{
		apiService: api,
		netService: net,
	}
}

type RemoteRestoreParams struct {
	BackupID     string
	Path         string
	TargetHostID string
	TargetPath   string
}

func (s *RestoreService) ExecuteRemote(params RemoteRestoreParams) (string, error) {
	req := dto.RestoreRequest{
		BackupID:     params.BackupID,
		Path:         params.Path,
		RestoreType:  "remote",
		TargetHostID: params.TargetHostID,
		TargetPath:   params.TargetPath,
	}

	return s.apiService.RequestRestore(params.BackupID, req)
}

type LocalRestoreParams struct {
	BackupID   string
	Path       string
	LocalDest  string
	CustomAddr string
}

func (s *RestoreService) ExecuteLocal(params LocalRestoreParams) error {
	// 1. Generate Token
	token := s.generateRandomToken(16)

	// 2. Start Listener
	_, port, listener, err := s.netService.ListenTCP()
	if err != nil {
		return fmt.Errorf("error starting listener: %w", err)
	}
	defer listener.Close()

	// 3. Determine Address
	restoreAddr := params.CustomAddr
	if restoreAddr == "" {
		restoreAddr = fmt.Sprintf("%s:%d", s.netService.GetLocalIP(), port)
	} else if !s.containsPort(restoreAddr) {
		restoreAddr = fmt.Sprintf("%s:%d", restoreAddr, port)
	}

	fmt.Printf("Listening for worker connection on %s\n", restoreAddr)
	fmt.Printf("Authentication token: %s\n", token)

	// 4. Send request to Backend
	req := dto.RestoreRequest{
		BackupID:     params.BackupID,
		Path:         params.Path,
		RestoreType:  "local",
		RestoreAddr:  restoreAddr,
		RestoreToken: token,
	}

	fmt.Println("Requesting restore from server...")
	_, err = s.apiService.RequestRestore(params.BackupID, req)
	if err != nil {
		return err
	}

	fmt.Println("Waiting for worker to connect and send data...")

	// 5. Accept Connection and stream data
	stream, err := s.netService.AcceptAndValidate(listener, token)
	if err != nil {
		return err
	}
	defer stream.Close()

	fmt.Println("Token validated. Receiving data...")

	// 6. Extract data
	return s.netService.ExtractTarGz(stream, params.LocalDest)
}

func (s *RestoreService) generateRandomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "fallback-token-123"
	}
	return hex.EncodeToString(b)
}

func (s *RestoreService) containsPort(sAddr string) bool {
	_, _, err := net.SplitHostPort(sAddr)
	return err == nil
}

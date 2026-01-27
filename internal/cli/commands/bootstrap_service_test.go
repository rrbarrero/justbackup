package commands

import (
	"errors"
	"testing"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
)

func TestBootstrapService_Execute(t *testing.T) {
	tests := []struct {
		name        string
		apiMock     *MockAPIService
		sshMock     *MockSSHService
		expectError bool
	}{
		{
			name: "successful bootstrap",
			apiMock: &MockAPIService{
				GetSSHKeyFunc: func() (string, error) {
					return "valid-key", nil
				},
				RegisterHostFunc: func(req dto.CreateHostRequest) error {
					if req.Name != "Test Host" {
						return errors.New("unexpected host name")
					}
					return nil
				},
			},
			sshMock: &MockSSHService{
				InstallKeyFunc: func(host string, port int, user string, password string, publicKey string) error {
					if publicKey != "valid-key" {
						return errors.New("unexpected public key")
					}
					return nil
				},
			},
			expectError: false,
		},
		{
			name: "failed to get SSH key",
			apiMock: &MockAPIService{
				GetSSHKeyFunc: func() (string, error) {
					return "", errors.New("api error")
				},
			},
			sshMock:     &MockSSHService{},
			expectError: true,
		},
		{
			name:    "failed to install SSH key",
			apiMock: &MockAPIService{},
			sshMock: &MockSSHService{
				InstallKeyFunc: func(host string, port int, user string, password string, publicKey string) error {
					return errors.New("ssh error")
				},
			},
			expectError: true,
		},
		{
			name: "failed to register host",
			apiMock: &MockAPIService{
				RegisterHostFunc: func(req dto.CreateHostRequest) error {
					return errors.New("registration error")
				},
			},
			sshMock:     &MockSSHService{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewBootstrapService(tt.sshMock, tt.apiMock)
			params := BootstrapParams{
				Name: "Test Host",
				Host: "1.2.3.4",
				User: "root",
				Port: 22,
			}
			err := svc.Execute(params)

			if (err != nil) != tt.expectError {
				t.Fatalf("expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}

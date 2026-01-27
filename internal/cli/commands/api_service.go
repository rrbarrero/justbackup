package commands

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/rrbarrero/justbackup/internal/backup/application/dto"
	"github.com/rrbarrero/justbackup/internal/cli/client"
)

type apiServiceImpl struct {
	client *client.Client
}

func NewAPIService(client *client.Client) APIService {
	return &apiServiceImpl{client: client}
}

func (s *apiServiceImpl) GetSSHKey() (string, error) {
	keyData, err := s.client.Get("/settings/ssh-key")
	if err != nil {
		return "", fmt.Errorf("error fetching SSH key: %w", err)
	}

	var keyResp map[string]string
	if err := json.Unmarshal(keyData, &keyResp); err != nil {
		return "", fmt.Errorf("error parsing SSH key response: %w", err)
	}
	publicKey, ok := keyResp["publicKey"]
	if !ok || publicKey == "" {
		return "", fmt.Errorf("server returned empty SSH public key")
	}

	return publicKey, nil
}

func (s *apiServiceImpl) RegisterHost(req dto.CreateHostRequest) error {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	_, err = s.client.Post("/hosts", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error registering host: %w", err)
	}

	return nil
}

func (s *apiServiceImpl) RequestRestore(backupID string, req dto.RestoreRequest) (string, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	endpoint := fmt.Sprintf("/backups/%s/restore", backupID)
	data, err := s.client.Post(endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("error requesting restore: %w", err)
	}

	var resp struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("restore triggered, but failed to parse response: %w", err)
	}

	return resp.TaskID, nil
}

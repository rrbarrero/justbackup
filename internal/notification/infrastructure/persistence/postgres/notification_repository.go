package postgres

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/rrbarrero/justbackup/internal/notification/domain/entities"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
	"github.com/rrbarrero/justbackup/internal/shared/infrastructure/crypto"
)

type NotificationRepositoryPostgres struct {
	db        *sql.DB
	cryptoSvc crypto.EncryptionService
}

func NewNotificationRepositoryPostgres(db *sql.DB, cryptoSvc crypto.EncryptionService) *NotificationRepositoryPostgres {
	return &NotificationRepositoryPostgres{
		db:        db,
		cryptoSvc: cryptoSvc,
	}
}

func (r *NotificationRepositoryPostgres) GetSettings(ctx context.Context, userID int, providerType string) (*entities.NotificationSettings, error) {
	query := `
		SELECT user_id, provider_type, config, enabled
		FROM notification_settings
		WHERE user_id = $1 AND provider_type = $2
	`
	var uid int
	var pType string
	var config []byte
	var enabled bool

	err := r.db.QueryRowContext(ctx, query, userID, providerType).Scan(&uid, &pType, &config, &enabled)
	if err == sql.ErrNoRows {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Unwrap JSON object
	var wrapper map[string]string
	if err := json.Unmarshal(config, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification config wrapper: %w", err)
	}

	encryptedBase64, ok := wrapper["encrypted"]
	if !ok {
		return nil, fmt.Errorf("config is not encrypted (missing 'encrypted' key)")
	}

	// Decode Base64
	decodedConfig, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode notification config from base64: %w", err)
	}

	// Decrypt config
	decryptedConfig, err := r.cryptoSvc.Decrypt(decodedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt notification config: %w", err)
	}

	return entities.NewNotificationSettings(uid, pType, json.RawMessage(decryptedConfig), enabled), nil
}

func (r *NotificationRepositoryPostgres) SaveSettings(ctx context.Context, settings *entities.NotificationSettings) error {
	// Encrypt config
	encryptedConfig, err := r.cryptoSvc.Encrypt([]byte(settings.Config))
	if err != nil {
		return fmt.Errorf("failed to encrypt notification config: %w", err)
	}

	// Encode to Base64
	encodedConfig := base64.StdEncoding.EncodeToString(encryptedConfig)

	// Wrap in JSON object to satisfy JSONB column constraint
	wrapper := map[string]string{"encrypted": encodedConfig}
	jsonWrapper, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted config wrapper: %w", err)
	}

	query := `
		INSERT INTO notification_settings (user_id, provider_type, config, enabled)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, provider_type) DO UPDATE SET
			config = EXCLUDED.config,
			enabled = EXCLUDED.enabled
	`
	_, err = r.db.ExecContext(ctx, query, settings.UserID, settings.ProviderType, jsonWrapper, settings.Enabled)
	return err
}

func (r *NotificationRepositoryPostgres) GetAllEnabledSettings(ctx context.Context) ([]*entities.NotificationSettings, error) {
	query := `
		SELECT user_id, provider_type, config, enabled
		FROM notification_settings
		WHERE enabled = TRUE
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var settingsList []*entities.NotificationSettings
	for rows.Next() {
		var uid int
		var pType string
		var config []byte
		var enabled bool

		if err := rows.Scan(&uid, &pType, &config, &enabled); err != nil {
			return nil, err
		}

		// Unwrap JSON object
		var wrapper map[string]string
		if err := json.Unmarshal(config, &wrapper); err != nil {
			return nil, fmt.Errorf("failed to unmarshal notification config wrapper for user %d: %w", uid, err)
		}

		encryptedBase64, ok := wrapper["encrypted"]
		if !ok {
			return nil, fmt.Errorf("config for user %d is not encrypted", uid)
		}

		// Decode Base64
		decodedConfig, err := base64.StdEncoding.DecodeString(encryptedBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode notification config from base64 for user %d: %w", uid, err)
		}

		// Decrypt config
		decryptedConfig, err := r.cryptoSvc.Decrypt(decodedConfig)
		if err != nil {
			// Return error if decryption fails to ensure data integrity.
			return nil, fmt.Errorf("failed to decrypt notification config for user %d: %w", uid, err)
		}

		settingsList = append(settingsList, entities.NewNotificationSettings(uid, pType, json.RawMessage(decryptedConfig), enabled))
	}
	return settingsList, nil
}

package postgres

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

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
		log.Printf("ERROR: Database scan error for userID=%d, provider=%s: %v", userID, providerType, err)
		return nil, err
	}

	// Unwrap JSON object
	var wrapper map[string]string
	if err := json.Unmarshal(config, &wrapper); err != nil {
		log.Printf("ERROR: failed to unmarshal notification config wrapper for user %d, provider %s: %v", userID, providerType, err)
		return nil, fmt.Errorf("failed to unmarshal notification config wrapper: %w", err)
	}

	encryptedBase64, ok := wrapper["encrypted"]
	if !ok {
		log.Printf("WARNING: notification config for user %d, provider %s is not encrypted", userID, providerType)
		return entities.NewNotificationSettings(uid, pType, json.RawMessage("{}"), enabled), nil
	}

	// Decode Base64
	decodedConfig, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		log.Printf("ERROR: failed to decode notification config from base64 for user %d, provider %s: %v", userID, providerType, err)
		return entities.NewNotificationSettings(uid, pType, json.RawMessage("{}"), enabled), nil
	}

	// Decrypt config
	decryptedConfig, err := r.cryptoSvc.Decrypt(decodedConfig)
	if err != nil {
		log.Printf("SECURITY_ERROR: failed to decrypt notification config for user %d, provider %s: %v. Ciphertext length: %d", userID, providerType, err, len(decodedConfig))
		// We return a clean object even if decryption fails to avoid breaking the UI
		return entities.NewNotificationSettings(uid, pType, json.RawMessage("{}"), enabled), nil
	}

	return entities.NewNotificationSettings(uid, pType, json.RawMessage(decryptedConfig), enabled), nil
}

func (r *NotificationRepositoryPostgres) SaveSettings(ctx context.Context, settings *entities.NotificationSettings) error {
	// Encrypt config
	encryptedConfig, err := r.cryptoSvc.Encrypt([]byte(settings.Config))
	if err != nil {
		log.Printf("ERROR: failed to encrypt notification config for user %d: %v", settings.UserID, err)
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
			log.Printf("ERROR: failed to unmarshal notification config wrapper for user %d: %v", uid, err)
			continue // Skip corrupted entry
		}

		encryptedBase64, ok := wrapper["encrypted"]
		if !ok {
			log.Printf("WARNING: config for user %d is not encrypted", uid)
			continue
		}

		// Decode Base64
		decodedConfig, err := base64.StdEncoding.DecodeString(encryptedBase64)
		if err != nil {
			log.Printf("ERROR: failed to decode notification config from base64 for user %d: %v", uid, err)
			continue
		}

		// Decrypt config
		decryptedConfig, err := r.cryptoSvc.Decrypt(decodedConfig)
		if err != nil {
			log.Printf("SECURITY_ERROR: failed to decrypt notification config for user %d: %v", uid, err)
			continue
		}

		settingsList = append(settingsList, entities.NewNotificationSettings(uid, pType, json.RawMessage(decryptedConfig), enabled))
	}
	return settingsList, nil
}

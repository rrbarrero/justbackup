package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/rrbarrero/justbackup/internal/auth/domain/entities"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

type AuthTokenRepositoryPostgres struct {
	db *sql.DB
}

func NewAuthTokenRepositoryPostgres(db *sql.DB) *AuthTokenRepositoryPostgres {
	return &AuthTokenRepositoryPostgres{
		db: db,
	}
}

func (r *AuthTokenRepositoryPostgres) Save(ctx context.Context, token *entities.AuthToken) error {
	query := `
		INSERT INTO auth_tokens (id, token_hash, created_at, last_used_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			token_hash = EXCLUDED.token_hash,
			created_at = EXCLUDED.created_at,
			last_used_at = EXCLUDED.last_used_at
	`

	_, err := r.db.ExecContext(ctx, query,
		token.ID(),
		token.TokenHash(),
		token.CreatedAt(),
		token.LastUsedAt(),
	)
	return err
}

func (r *AuthTokenRepositoryPostgres) Get(ctx context.Context) (*entities.AuthToken, error) {
	query := `
		SELECT id, token_hash, created_at, last_used_at
		FROM auth_tokens
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query)

	var id uuid.UUID
	var tokenHash string
	var createdAt time.Time
	var lastUsedAt *time.Time

	err := row.Scan(&id, &tokenHash, &createdAt, &lastUsedAt)
	if err == sql.ErrNoRows {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return entities.RestoreAuthToken(id, tokenHash, createdAt, lastUsedAt), nil
}

func (r *AuthTokenRepositoryPostgres) Delete(ctx context.Context) error {
	query := `DELETE FROM auth_tokens`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

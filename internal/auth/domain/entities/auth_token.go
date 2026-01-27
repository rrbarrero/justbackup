package entities

import (
	"time"

	"github.com/google/uuid"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
)

var NowFunc = time.Now

type AuthToken struct {
	shared.AggregateRoot
	id         uuid.UUID
	tokenHash  string
	createdAt  time.Time
	lastUsedAt *time.Time
}

func NewAuthToken(hash string) *AuthToken {
	return &AuthToken{
		id:        uuid.New(),
		tokenHash: hash,
		createdAt: NowFunc(),
	}
}

// RestoreAuthToken restores an AuthToken from persistence
func RestoreAuthToken(id uuid.UUID, hash string, createdAt time.Time, lastUsedAt *time.Time) *AuthToken {
	return &AuthToken{
		id:         id,
		tokenHash:  hash,
		createdAt:  createdAt,
		lastUsedAt: lastUsedAt,
	}
}

func (t *AuthToken) ID() uuid.UUID {
	return t.id
}

func (t *AuthToken) TokenHash() string {
	return t.tokenHash
}

func (t *AuthToken) CreatedAt() time.Time {
	return t.createdAt
}

func (t *AuthToken) LastUsedAt() *time.Time {
	return t.lastUsedAt
}

func (t *AuthToken) MarkUsed() {
	now := NowFunc()
	t.lastUsedAt = &now
}

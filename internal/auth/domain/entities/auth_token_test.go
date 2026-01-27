package entities_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rrbarrero/justbackup/internal/auth/domain/entities"
	"github.com/stretchr/testify/assert"
)

func TestAuthToken(t *testing.T) {
	originalNowFunc := entities.NowFunc
	defer func() { entities.NowFunc = originalNowFunc }()

	t.Run("NewAuthToken should create a token with generated ID and current time", func(t *testing.T) {
		hash := "hashed-token-value"

		// Mock time.Now for deterministic test
		fixedTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return fixedTime }

		token := entities.NewAuthToken(hash)

		assert.NotNil(t, token)
		assert.NotEmpty(t, token.ID())
		_, err := uuid.Parse(token.ID().String())
		assert.NoError(t, err)
		assert.Equal(t, hash, token.TokenHash())
		assert.Equal(t, fixedTime, token.CreatedAt())
		assert.Nil(t, token.LastUsedAt())
	})

	t.Run("RestoreAuthToken should create a token with provided values", func(t *testing.T) {
		id := uuid.New()
		hash := "restored-hash"
		createdAt := time.Date(2022, time.December, 31, 23, 59, 59, 0, time.UTC)
		lastUsedAt := time.Date(2023, time.January, 1, 10, 0, 0, 0, time.UTC)

		token := entities.RestoreAuthToken(id, hash, createdAt, &lastUsedAt)

		assert.NotNil(t, token)
		assert.Equal(t, id, token.ID())
		assert.Equal(t, hash, token.TokenHash())
		assert.Equal(t, createdAt, token.CreatedAt())
		assert.NotNil(t, token.LastUsedAt())
		assert.Equal(t, lastUsedAt, *token.LastUsedAt())
	})

	t.Run("RestoreAuthToken should handle nil lastUsedAt", func(t *testing.T) {
		id := uuid.New()
		hash := "restored-hash"
		createdAt := time.Date(2022, time.December, 31, 23, 59, 59, 0, time.UTC)

		token := entities.RestoreAuthToken(id, hash, createdAt, nil)

		assert.NotNil(t, token)
		assert.Equal(t, id, token.ID())
		assert.Equal(t, hash, token.TokenHash())
		assert.Equal(t, createdAt, token.CreatedAt())
		assert.Nil(t, token.LastUsedAt())
	})

	t.Run("MarkUsed should set lastUsedAt to current time", func(t *testing.T) {
		creationTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return creationTime }

		token := entities.NewAuthToken("test-hash")

		assert.Nil(t, token.LastUsedAt())

		// Simulate time passing
		usedTime := time.Date(2023, time.January, 2, 14, 30, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return usedTime }

		token.MarkUsed()

		assert.NotNil(t, token.LastUsedAt())
		assert.Equal(t, usedTime, *token.LastUsedAt())
	})

	t.Run("MarkUsed should update lastUsedAt when called multiple times", func(t *testing.T) {
		creationTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return creationTime }

		token := entities.NewAuthToken("test-hash")

		// First use
		firstUseTime := time.Date(2023, time.January, 2, 10, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return firstUseTime }
		token.MarkUsed()
		assert.Equal(t, firstUseTime, *token.LastUsedAt())

		// Second use
		secondUseTime := time.Date(2023, time.January, 3, 15, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return secondUseTime }
		token.MarkUsed()
		assert.Equal(t, secondUseTime, *token.LastUsedAt())
	})

	t.Run("AuthToken getter methods should return correct values", func(t *testing.T) {
		id := uuid.New()
		hash := "getter-test-hash"
		createdAt := time.Date(2023, time.February, 15, 9, 30, 0, 0, time.UTC)
		lastUsedAt := time.Date(2023, time.March, 1, 14, 45, 0, 0, time.UTC)

		token := entities.RestoreAuthToken(id, hash, createdAt, &lastUsedAt)

		assert.Equal(t, id, token.ID())
		assert.Equal(t, hash, token.TokenHash())
		assert.Equal(t, createdAt, token.CreatedAt())
		assert.NotNil(t, token.LastUsedAt())
		assert.Equal(t, lastUsedAt, *token.LastUsedAt())
	})

	t.Run("AuthToken should preserve hash value", func(t *testing.T) {
		hash := "$2a$10$randomHashedPasswordValue1234567890"

		fixedTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
		entities.NowFunc = func() time.Time { return fixedTime }

		token := entities.NewAuthToken(hash)

		assert.Equal(t, hash, token.TokenHash())
	})
}

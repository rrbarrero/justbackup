package application

import (
	"context"
	"errors"
	"testing"

	"github.com/rrbarrero/justbackup/internal/auth/domain/entities"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockAuthTokenRepository is a mock of interfaces.AuthTokenRepository
type MockAuthTokenRepository struct {
	mock.Mock
}

func (m *MockAuthTokenRepository) Save(ctx context.Context, token *entities.AuthToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockAuthTokenRepository) Get(ctx context.Context) (*entities.AuthToken, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AuthToken), args.Error(1)
}

func (m *MockAuthTokenRepository) Delete(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestAuthService_GenerateToken(t *testing.T) {
	mockRepo := new(MockAuthTokenRepository)
	service := NewAuthService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Delete", ctx).Return(nil).Once()
		mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.AuthToken")).Return(nil).Once()

		res, err := service.GenerateToken(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.Token)
		assert.True(t, res.Exists)
		assert.NotZero(t, res.CreatedAt)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository delete error", func(t *testing.T) {
		expectedErr := errors.New("failed to delete existing token")
		mockRepo.On("Delete", ctx).Return(expectedErr).Once()

		res, err := service.GenerateToken(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, res)

		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Save")
	})

	t.Run("repository save error", func(t *testing.T) {
		expectedErr := errors.New("failed to save token")
		mockRepo.On("Delete", ctx).Return(nil).Once()
		mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.AuthToken")).Return(expectedErr).Once()

		res, err := service.GenerateToken(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, res)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_GetTokenStatus(t *testing.T) {
	mockRepo := new(MockAuthTokenRepository)
	service := NewAuthService(mockRepo)
	ctx := context.Background()

	t.Run("token exists", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("test-token"), bcrypt.DefaultCost)
		existingToken := entities.NewAuthToken(string(hash))

		mockRepo.On("Get", ctx).Return(existingToken, nil).Once()

		res, err := service.GetTokenStatus(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, res.Exists)
		assert.Equal(t, existingToken.CreatedAt(), res.CreatedAt)
		assert.Empty(t, res.Token) // Token should not be returned in status

		mockRepo.AssertExpectations(t)
	})

	t.Run("token not found", func(t *testing.T) {
		mockRepo.On("Get", ctx).Return(nil, shared.ErrNotFound).Once()

		res, err := service.GetTokenStatus(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.False(t, res.Exists)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("db connection error")
		mockRepo.On("Get", ctx).Return(nil, expectedErr).Once()

		res, err := service.GetTokenStatus(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, res)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_RevokeToken(t *testing.T) {
	mockRepo := new(MockAuthTokenRepository)
	service := NewAuthService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Delete", ctx).Return(nil).Once()

		err := service.RevokeToken(ctx)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("failed to delete token")
		mockRepo.On("Delete", ctx).Return(expectedErr).Once()

		err := service.RevokeToken(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockRepo := new(MockAuthTokenRepository)
	service := NewAuthService(mockRepo)
	ctx := context.Background()

	t.Run("valid token", func(t *testing.T) {
		token := "my-secret-token"
		hash, _ := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
		authToken := entities.NewAuthToken(string(hash))

		mockRepo.On("Get", ctx).Return(authToken, nil).Once()
		mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.AuthToken")).Return(nil).Once()

		valid, err := service.ValidateToken(ctx, token)

		assert.NoError(t, err)
		assert.True(t, valid)
		assert.NotNil(t, authToken.LastUsedAt()) // Verify token was marked as used

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid token - wrong password", func(t *testing.T) {
		correctToken := "correct-token"
		wrongToken := "wrong-token"
		hash, _ := bcrypt.GenerateFromPassword([]byte(correctToken), bcrypt.DefaultCost)
		authToken := entities.NewAuthToken(string(hash))

		mockRepo.On("Get", ctx).Return(authToken, nil).Once()

		valid, err := service.ValidateToken(ctx, wrongToken)

		assert.NoError(t, err)
		assert.False(t, valid)

		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Save") // Should not update lastUsedAt for invalid token
	})

	t.Run("token not found", func(t *testing.T) {
		mockRepo.On("Get", ctx).Return(nil, shared.ErrNotFound).Once()

		valid, err := service.ValidateToken(ctx, "any-token")

		assert.NoError(t, err)
		assert.False(t, valid)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("db connection error")
		mockRepo.On("Get", ctx).Return(nil, expectedErr).Once()

		valid, err := service.ValidateToken(ctx, "any-token")

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.False(t, valid)

		mockRepo.AssertExpectations(t)
	})
}

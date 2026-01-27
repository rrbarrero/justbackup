package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/rrbarrero/justbackup/internal/auth/application/dto"
	"github.com/rrbarrero/justbackup/internal/auth/domain/entities"
	"github.com/rrbarrero/justbackup/internal/auth/domain/interfaces"
	shared "github.com/rrbarrero/justbackup/internal/shared/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo interfaces.AuthTokenRepository
}

func NewAuthService(repo interfaces.AuthTokenRepository) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

func (s *AuthService) GenerateToken(ctx context.Context) (*dto.TokenResponse, error) {
	// 0. Revoke existing tokens to ensure only one valid token exists
	if err := s.repo.Delete(ctx); err != nil {
		return nil, err
	}

	// 1. Generate random token
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	token := base64.URLEncoding.EncodeToString(b)

	// 2. Hash token
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. Create entity
	authToken := entities.NewAuthToken(string(hash))

	// 4. Save (upsert handled by repo)
	if err := s.repo.Save(ctx, authToken); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		Token:     token,
		CreatedAt: authToken.CreatedAt(),
		Exists:    true,
	}, nil
}

func (s *AuthService) GetTokenStatus(ctx context.Context) (*dto.TokenResponse, error) {
	token, err := s.repo.Get(ctx)
	if errors.Is(err, shared.ErrNotFound) {
		return &dto.TokenResponse{Exists: false}, nil
	}
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		CreatedAt: token.CreatedAt(),
		Exists:    true,
	}, nil
}

func (s *AuthService) RevokeToken(ctx context.Context) error {
	return s.repo.Delete(ctx)
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (bool, error) {
	authToken, err := s.repo.Get(ctx)
	if errors.Is(err, shared.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(authToken.TokenHash()), []byte(token))
	if err != nil {
		return false, nil
	}

	// Update last used
	authToken.MarkUsed()
	_ = s.repo.Save(ctx, authToken) // Ignore error, not critical

	return true, nil
}

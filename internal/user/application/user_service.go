package application

import (
	"context"
	"errors"

	"github.com/rrbarrero/justbackup/internal/user/domain"
	"github.com/rrbarrero/justbackup/internal/user/domain/entities"
	"github.com/rrbarrero/justbackup/internal/user/domain/interfaces"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo interfaces.UserRepository
}

func NewUserService(repo interfaces.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) IsSetupRequired(ctx context.Context) (bool, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *UserService) RegisterInitialUser(ctx context.Context, username, password string) error {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("setup already completed")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := entities.NewUser(username, string(hashedPassword))
	return s.repo.Save(ctx, user)
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (*entities.User, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		// We should probably check for sql.ErrNoRows here if the repo doesn't wrap it
		// But assuming the repo returns domain error or sql error
		// For now let's just return nil, nil if user not found logic is handled by caller checking for nil user
		// Actually, let's keep the existing logic but return user
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, nil
		}
		// If it's sql.ErrNoRows, we should treat it as not found
		// But we don't import sql here.
		// Let's assume repo returns error.
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

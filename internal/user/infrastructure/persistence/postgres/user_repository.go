package postgres

import (
	"context"
	"database/sql"

	"github.com/rrbarrero/justbackup/internal/user/domain/entities"
)

type UserRepositoryPostgres struct {
	db *sql.DB
}

func NewUserRepositoryPostgres(db *sql.DB) *UserRepositoryPostgres {
	return &UserRepositoryPostgres{db: db}
}

func (r *UserRepositoryPostgres) Save(ctx context.Context, user *entities.User) error {
	query := "INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, user.Username, user.PasswordHash).Scan(&user.ID)
	return err
}

func (r *UserRepositoryPostgres) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	query := "SELECT id, username, password_hash FROM users WHERE username = $1"
	row := r.db.QueryRowContext(ctx, query, username)
	var user entities.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryPostgres) Count(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM users"
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

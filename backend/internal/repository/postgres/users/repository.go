package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

func NewRepository(pool *pgxpool.Pool, timeout time.Duration) *Repository {
	return &Repository{pool: pool, timeout: timeout}
}

func (repository *Repository) Create(ctx context.Context, email, passwordHash string) (domain.User, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.timeout)
	defer cancel()

	const query = `
		INSERT INTO users (role, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, role, email, password_hash, created_at
	`
	var user domain.User
	err := repository.pool.QueryRow(
		queryContext,
		query,
		domain.RoleUser,
		email,
		passwordHash,
	).Scan(&user.ID, &user.Role, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return domain.User{}, coreerrors.ErrEmailExists
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (repository *Repository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.timeout)
	defer cancel()

	const query = `
		SELECT id, role, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`
	var user domain.User
	err := repository.pool.QueryRow(queryContext, query, email).Scan(
		&user.ID,
		&user.Role,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, coreerrors.ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("select user by email: %w", err)
	}
	return user, nil
}

func (repository *Repository) GetByID(_ context.Context, _ string) (domain.User, error) {
	return domain.User{}, coreerrors.ErrNotImplemented
}

func (repository *Repository) SaveRefreshSession(_ context.Context, _, _ string) error {
	return coreerrors.ErrNotImplemented
}

func (repository *Repository) GetRefreshSession(_ context.Context, _ string) (string, error) {
	return "", coreerrors.ErrNotImplemented
}

func (repository *Repository) DeleteRefreshSession(_ context.Context, _ string) error {
	return coreerrors.ErrNotImplemented
}

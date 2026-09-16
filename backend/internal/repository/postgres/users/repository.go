package repository_users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"SPOproject/internal/core/domain"
	core_errors "SPOproject/internal/core/errors"
	"SPOproject/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	txManager repository.TxManager
}

func NewRepository(txManager repository.TxManager) *Repository {
	return &Repository{txManager: txManager}
}
func (repository *Repository) CreateUser(ctx context.Context, email, passwordHash string) (domain.User, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.txManager.GetTimeout())
	defer cancel()

	const query = `
		INSERT INTO users (role, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, role, email, password_hash, created_at
	`
	var user domain.User
	err := repository.txManager.GetExecutor(queryContext).QueryRow(
		queryContext,
		query,
		domain.RoleUser,
		email,
		passwordHash,
	).Scan(&user.ID, &user.Role, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return domain.User{}, core_errors.ErrEmailExists
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (repository *Repository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.txManager.GetTimeout())
	defer cancel()

	const query = `
		SELECT id, role, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`
	var user domain.User
	err := repository.txManager.GetExecutor(queryContext).QueryRow(queryContext, query, email).Scan(
		&user.ID,
		&user.Role,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, core_errors.ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("select user by email: %w", err)
	}
	return user, nil
}

func (repository *Repository) GetByID(ctx context.Context, userID string) (domain.User, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.txManager.GetTimeout())
	defer cancel()

	const query = `
		SELECT id, role, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`
	var user domain.User
	if err := repository.txManager.GetExecutor(queryContext).QueryRow(
		queryContext,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Role,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, core_errors.ErrNotAuthorized
		}
		return domain.User{}, fmt.Errorf("select user by id: %w", err)
	}

	return user, nil
}

func (repository *Repository) CreateRefreshSession(
	ctx context.Context,
	userID string,
	refreshTokenHash string,
	expiresAt time.Time,
) error {
	queryContext, cancel := context.WithTimeout(ctx, repository.txManager.GetTimeout())
	defer cancel()

	const query = `
		INSERT INTO auth_sessions (user_id, refresh_token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	if _, err := repository.txManager.GetExecutor(queryContext).Exec(
		queryContext,
		query,
		userID,
		refreshTokenHash,
		expiresAt,
	); err != nil {
		return fmt.Errorf("create refresh session: %w", err)
	}

	return nil
}

func (repository *Repository) RotateRefreshSession(
	ctx context.Context,
	userID string,
	currentTokenHash string,
	newTokenHash string,
	expiresAt time.Time,
) error {
	queryContext, cancel := context.WithTimeout(ctx, repository.txManager.GetTimeout())
	defer cancel()

	const query = `
		UPDATE auth_sessions
		SET refresh_token_hash = $1,
			expires_at = $2,
			revoked_at = NULL,
			created_at = now()
		WHERE user_id = $3
			AND refresh_token_hash = $4
			AND revoked_at IS NULL
			AND expires_at > now()
	`
	commandTag, err := repository.txManager.GetExecutor(queryContext).Exec(
		queryContext,
		query,
		newTokenHash,
		expiresAt,
		userID,
		currentTokenHash,
	)
	if err != nil {
		return fmt.Errorf("rotate refresh session: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return core_errors.ErrNotAuthorized
	}

	return nil
}

func (repository *Repository) GetRefreshSession(ctx context.Context, refreshTokenHash string) (string, error) {
	queryContext, cancel := context.WithTimeout(ctx, repository.txManager.GetTimeout())
	defer cancel()

	const query = `
		SELECT user_id
		FROM auth_sessions
		WHERE refresh_token_hash = $1
			AND revoked_at IS NULL
			AND expires_at > now()
	`
	var userID string
	if err := repository.txManager.GetExecutor(queryContext).QueryRow(
		queryContext,
		query,
		refreshTokenHash,
	).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", core_errors.ErrNotAuthorized
		}
		return "", fmt.Errorf("get refresh session: %w", err)
	}

	return userID, nil
}

func (repository *Repository) DeleteRefreshSession(ctx context.Context, userID, refreshTokenHash string) error {
	queryContext, cancel := context.WithTimeout(ctx, repository.txManager.GetTimeout())
	defer cancel()

	const query = `
		UPDATE auth_sessions
		SET revoked_at = now()
		WHERE user_id = $1
			AND refresh_token_hash = $2
			AND revoked_at IS NULL
	`
	if _, err := repository.txManager.GetExecutor(queryContext).Exec(
		queryContext,
		query,
		userID,
		refreshTokenHash,
	); err != nil {
		return fmt.Errorf("delete refresh session: %w", err)
	}

	return nil
}

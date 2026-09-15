package auth

import (
	"context"

	"SPOproject/internal/core/domain"
)

type userRepository interface {
	Create(ctx context.Context, email, passwordHash string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, userID string) (domain.User, error)
	SaveRefreshSession(ctx context.Context, userID, refreshToken string) error
	GetRefreshSession(ctx context.Context, refreshToken string) (string, error)
	DeleteRefreshSession(ctx context.Context, refreshToken string) error
}

type passwordHasher interface {
	HashPassword(password []byte) ([]byte, error)
	Compare(password, passwordHash []byte) error
}

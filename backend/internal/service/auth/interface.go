package auth

import (
	"context"
	"time"

	"SPOproject/internal/core/auth"
	"SPOproject/internal/core/domain"
)

type userRepository interface {
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	CreateUser(ctx context.Context, email, passwordHash string) (domain.User, error)
	GetByID(ctx context.Context, userID string) (domain.User, error)
	CreateRefreshSession(ctx context.Context, userID, refreshTokenHash string, expiresAt time.Time) error
	RotateRefreshSession(ctx context.Context, userID, currentTokenHash, newTokenHash string, expiresAt time.Time) error
	GetRefreshSession(ctx context.Context, refreshTokenHash string) (string, error)
	DeleteRefreshSession(ctx context.Context, userID, refreshTokenHash string) error
}

type hasher interface {
	HashPassword(password []byte) ([]byte, error)
	Compare(password, passwordHash []byte) error
	HashToken(token string) string
}

type tokensProvider interface {
	NewTokenWithClaims(user domain.User, tokenType auth.TokenType) (string, auth.Claims, error)
	NewToken(user domain.User, tokenType auth.TokenType) (string, error)
	ParseToken(tokenString string, tokenType auth.TokenType) (auth.Claims, error)
}

type txManager interface {
	WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

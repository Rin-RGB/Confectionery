package auth

import (
	"context"

	"SPOproject/internal/core/domain"
)

type authService interface {
	Register(ctx context.Context, credentials domain.Credentials) (domain.TokenPair, error)
	Login(ctx context.Context, credentials domain.Credentials) (domain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

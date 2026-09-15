package auth

import (
	"SPOproject/internal/core/auth"
	"context"

	"SPOproject/internal/core/domain"
)

type authService interface {
	Register(ctx context.Context, credentials domain.Credentials) (domain.User, error)
	Login(ctx context.Context, credentials domain.Credentials) (domain.User, error)
	Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}
type tokenProvider interface {
	ParseToken(tokenString string, tokenType auth.TokenType) (auth.Claims, error)
	NewToken(user domain.User, tokenType auth.TokenType) (string, error)
}

package auth

import (
	"context"
	"fmt"
	"strings"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
)

type Service struct {
	users  userRepository
	hasher passwordHasher
}

func NewService(users userRepository, hasher passwordHasher) *Service {
	return &Service{
		users:  users,
		hasher: hasher,
	}
}

func (service *Service) Register(ctx context.Context, credentials domain.Credentials) (domain.User, error) {
	if err := credentials.Validate(); err != nil {
		return domain.User{}, err
	}

	passwordHash, err := service.hasher.HashPassword([]byte(credentials.Password))
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	user, err := service.users.Create(
		ctx,
		strings.ToLower(strings.TrimSpace(credentials.Email)),
		string(passwordHash),
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("register user: %w", err)
	}
	user.PasswordHash = ""
	return user, nil
}

func (service *Service) Login(ctx context.Context, credentials domain.Credentials) (domain.User, error) {
	if err := credentials.Validate(); err != nil {
		return domain.User{}, coreerrors.ErrInvalidCredentials
	}

	user, err := service.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(credentials.Email)))
	if err != nil {
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	if err = service.hasher.Compare([]byte(credentials.Password), []byte(user.PasswordHash)); err != nil {
		return domain.User{}, coreerrors.ErrInvalidCredentials
	}

	user.PasswordHash = ""
	return user, nil
}

func (service *Service) Refresh(_ context.Context, _ string) (domain.TokenPair, error) {
	return domain.TokenPair{}, coreerrors.ErrNotImplemented
}

func (service *Service) Logout(_ context.Context, _ string) error {
	return coreerrors.ErrNotImplemented
}

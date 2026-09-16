package auth

import (
	"SPOproject/internal/core/auth"
	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
	"context"
	"fmt"
)

type Service struct {
	txManager      txManager
	usersRepo      userRepository
	hasher         hasher
	tokensProvider tokensProvider
}

func NewService(u userRepository, h hasher, p tokensProvider, t txManager) *Service {
	return &Service{
		usersRepo:      u,
		hasher:         h,
		tokensProvider: p,
		txManager:      t,
	}
}

func (s *Service) Register(ctx context.Context, credentials domain.Credentials) (domain.TokenPair, error) {
	if err := credentials.Validate(); err != nil {
		return domain.TokenPair{}, err
	}

	passwordHash, err := s.hasher.HashPassword([]byte(credentials.Password))
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("hash password: %w", err)
	}
	var tokens domain.TokenPair
	err = s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		user, err := s.usersRepo.CreateUser(
			txCtx,
			credentials.Email,
			string(passwordHash),
		)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		user.PasswordHash = ""
		tokens.AccessToken, err = s.tokensProvider.NewToken(user, auth.AccessType)
		if err != nil {
			return fmt.Errorf("failed to generate access token: %w", err)
		}
		refreshToken, claims, err := s.tokensProvider.NewTokenWithClaims(user, auth.RefreshType)
		if err != nil {
			return fmt.Errorf("failed to generate refresh token: %w", err)
		}
		tokens.RefreshToken = refreshToken
		err = s.usersRepo.CreateRefreshSession(
			txCtx,
			user.ID.String(),
			s.hasher.HashToken(refreshToken),
			claims.ExpiresAt.Time,
		)
		if err != nil {
			return fmt.Errorf("failed to save refresh token: %w", err)
		}
		return nil
	})
	if err != nil {
		return domain.TokenPair{}, err
	}
	return tokens, nil
}

func (s *Service) Login(ctx context.Context, credentials domain.Credentials) (domain.TokenPair, error) {
	if err := credentials.Validate(); err != nil {
		return domain.TokenPair{}, coreerrors.ErrInvalidCredentials
	}

	var tokens domain.TokenPair
	err := s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		user, err := s.usersRepo.GetUserByEmail(txCtx, credentials.Email)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		if err = s.hasher.Compare([]byte(credentials.Password), []byte(user.PasswordHash)); err != nil {
			return coreerrors.ErrInvalidCredentials
		}

		user.PasswordHash = ""
		tokens.AccessToken, err = s.tokensProvider.NewToken(user, auth.AccessType)
		if err != nil {
			return fmt.Errorf("failed to generate access token: %w", err)
		}
		refreshToken, claims, err := s.tokensProvider.NewTokenWithClaims(user, auth.RefreshType)
		if err != nil {
			return fmt.Errorf("failed to generate refresh token: %w", err)
		}
		tokens.RefreshToken = refreshToken
		if err = s.usersRepo.CreateRefreshSession(
			txCtx,
			user.ID.String(),
			s.hasher.HashToken(refreshToken),
			claims.ExpiresAt.Time,
		); err != nil {
			return fmt.Errorf("failed to save refresh token: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.TokenPair{}, err
	}

	return tokens, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	claims, err := s.tokensProvider.ParseToken(refreshToken, auth.RefreshType)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("validate refresh token: %w", err)
	}

	currentTokenHash := s.hasher.HashToken(refreshToken)
	var tokens domain.TokenPair
	err = s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		user, err := s.usersRepo.GetByID(txCtx, claims.UserID)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		user.PasswordHash = ""

		tokens.AccessToken, err = s.tokensProvider.NewToken(user, auth.AccessType)
		if err != nil {
			return fmt.Errorf("failed to generate access token: %w", err)
		}
		newRefreshToken, newRefreshClaims, err := s.tokensProvider.NewTokenWithClaims(user, auth.RefreshType)
		if err != nil {
			return fmt.Errorf("failed to generate refresh token: %w", err)
		}
		if newRefreshClaims.ExpiresAt == nil {
			return fmt.Errorf("refresh token expiration is missing: %w", coreerrors.ErrInternal)
		}

		if err = s.usersRepo.RotateRefreshSession(
			txCtx,
			user.ID.String(),
			currentTokenHash,
			s.hasher.HashToken(newRefreshToken),
			newRefreshClaims.ExpiresAt.Time,
		); err != nil {
			return fmt.Errorf("rotate refresh session: %w", err)
		}

		tokens.RefreshToken = newRefreshToken
		return nil
	})
	if err != nil {
		return domain.TokenPair{}, err
	}

	return tokens, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.tokensProvider.ParseToken(refreshToken, auth.RefreshType)
	if err != nil {
		return fmt.Errorf("validate refresh token: %w", err)
	}

	if err = s.usersRepo.DeleteRefreshSession(ctx, claims.UserID, s.hasher.HashToken(refreshToken)); err != nil {
		return fmt.Errorf("delete refresh session: %w", err)
	}

	return nil
}

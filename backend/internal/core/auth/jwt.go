package auth

import (
	"errors"
	"fmt"
	"time"

	"SPOproject/internal/core/domain"
	core_errors "SPOproject/internal/core/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AccessType  TokenType = "access"
	RefreshType TokenType = "refresh"
)

type Claims struct {
	UserID string          `json:"user_id"`
	Role   domain.UserRole `json:"role"`
	Type   TokenType       `json:"type"`
	jwt.RegisteredClaims
}

type JWTProvider struct {
	signingKey      []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTProvider(signingKey string, accessTokenTTL, refreshTokenTTL time.Duration) *JWTProvider {
	return &JWTProvider{
		signingKey:      []byte(signingKey),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}
func (p *JWTProvider) NewTokenWithClaims(user domain.User, tokenType TokenType) (string, Claims, error) {
	if tokenType == AccessType {
		return p.newToken(user, tokenType, p.accessTokenTTL)
	} else if tokenType == RefreshType {
		return p.newToken(user, tokenType, p.refreshTokenTTL)
	}
	return "", Claims{}, fmt.Errorf("wrong token type: %w", core_errors.ErrInternal)
}
func (p *JWTProvider) NewToken(user domain.User, tokenType TokenType) (string, error) {
	if tokenType == AccessType {
		token, _, err := p.newToken(user, tokenType, p.accessTokenTTL)
		return token, err
	} else if tokenType == RefreshType {
		token, _, err := p.newToken(user, tokenType, p.refreshTokenTTL)
		return token, err
	}
	return "", fmt.Errorf("wrong token type: %w", core_errors.ErrInternal)
}
func (p *JWTProvider) ParseToken(tokenString string, tokenType TokenType) (Claims, error) {

	tokenClaims := Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &tokenClaims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("wrong signing method: %v: %w", token.Header["alg"], core_errors.ErrInvalidToken)
		}
		return p.signingKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) && !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return Claims{}, fmt.Errorf("invalid token: %w", core_errors.ErrExpiredToken)
		}
		return Claims{}, fmt.Errorf("parse token: %v: %w", err, core_errors.ErrInvalidToken)
	}
	if !token.Valid {
		return Claims{}, core_errors.ErrInvalidToken
	}
	if tokenClaims.Type != tokenType {
		return Claims{}, fmt.Errorf("wrong token type: expected %q, got %q: %w", tokenType, tokenClaims.Type, core_errors.ErrInvalidToken)
	}
	return tokenClaims, nil
}

func (p *JWTProvider) newToken(user domain.User, tokenType TokenType, ttl time.Duration) (string, Claims, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: user.ID.String(),
		Role:   user.Role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   user.ID.String(),
			ID:        uuid.NewString(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(p.signingKey)
	if err != nil {
		return "", Claims{}, fmt.Errorf("sign token: %w", err)
	}
	return token, claims, nil
}

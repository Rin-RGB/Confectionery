package auth

import (
	core_errors "SPOproject/internal/core/errors"
	"errors"
	"fmt"
	"time"

	"SPOproject/internal/core/domain"

	"github.com/golang-jwt/jwt/v5"
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
func (p *JWTProvider) NewToken(user domain.User, tokenType TokenType) (string, error) {
	if tokenType == AccessType {
		return p.newToken(user, tokenType, p.accessTokenTTL)
	} else if tokenType == RefreshType {
		return p.newToken(user, tokenType, p.refreshTokenTTL)
	}
	return "", fmt.Errorf("wrong token type: %w", core_errors.ErrInternal)
}
func (p *JWTProvider) ParseToken(tokenString string, tokenType TokenType) (Claims, error) {

	tokenClaims := Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &tokenClaims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("wrong signing method: %v: %w", token.Header["alg"], core_errors.ErrInvalidRequest)
		}
		return p.signingKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Claims{}, fmt.Errorf("invalid token: %w", core_errors.ErrExpiredToken)
		}
		return Claims{}, fmt.Errorf("parse token: %w: %w", err, core_errors.ErrInvalidRequest)
	}
	if !token.Valid {
		return Claims{}, fmt.Errorf("invalid token: %w", core_errors.ErrInvalidRequest)
	}
	if tokenClaims.Type != tokenType {
		return Claims{}, fmt.Errorf("invalid token: wrong token type: %w", core_errors.ErrInvalidRequest)
	}
	return tokenClaims, nil
}

func (p *JWTProvider) newToken(user domain.User, tokenType TokenType, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: user.ID.String(),
		Role:   user.Role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   user.ID.String(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(p.signingKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return token, nil
}

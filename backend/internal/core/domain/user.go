package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	coreerrors "SPOproject/internal/core/errors"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Role         UserRole
	CreatedAt    time.Time
}

type Credentials struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (credentials Credentials) Validate() error {
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailPattern.MatchString(strings.TrimSpace(credentials.Email)) {
		return fmt.Errorf("invalid email: %w", coreerrors.ErrInvalidRequest)
	}
	if len(credentials.Password) < 3 {
		return fmt.Errorf("password must contain at least 3 characters: %w", coreerrors.ErrInvalidRequest)
	}
	return nil
}

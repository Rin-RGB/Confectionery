package core_middleware

import (
	"context"
	"fmt"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"

	"github.com/google/uuid"
)

type userIdKey struct{}

func GetUserIdFromCtx(ctx context.Context) (uuid.UUID, error) {
	userId, ok := ctx.Value(userIdKey{}).(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user id is missing in context: %w", coreerrors.ErrNotAuthorized)
	}
	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse user id from context: %w: %w", err, coreerrors.ErrNotAuthorized)
	}
	return userIdUUID, nil
}

type roleKey struct{}

func GetRoleFromCtx(ctx context.Context) (domain.UserRole, error) {
	role, ok := ctx.Value(roleKey{}).(domain.UserRole)
	if !ok {
		return "", fmt.Errorf("user role is missing in context: %w", coreerrors.ErrNotAuthorized)
	}
	return role, nil
}

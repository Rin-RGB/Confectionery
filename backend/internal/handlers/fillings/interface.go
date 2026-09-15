package fillings

import (
	"context"

	"SPOproject/internal/core/domain"
)

type fillingService interface {
	List(ctx context.Context) ([]domain.Filling, error)
	GetByID(ctx context.Context, fillingID string) (domain.Filling, error)
	Create(ctx context.Context, filling domain.Filling) (domain.Filling, error)
	Update(ctx context.Context, fillingID string, patch domain.FillingPatch) (domain.Filling, error)
	Hide(ctx context.Context, fillingID string) error
}

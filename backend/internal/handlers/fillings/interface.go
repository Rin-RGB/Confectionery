package fillings

import (
	"context"

	"SPOproject/internal/core/domain"
)

type fillingService interface {
	GetFillings(ctx context.Context) ([]domain.Filling, error)
	GetFillingByID(ctx context.Context, fillingID string) (domain.Filling, error)
	CreateFilling(ctx context.Context, filling domain.Filling, image []byte) (domain.Filling, error)
	UpdateFilling(ctx context.Context, fillingID string, patch domain.FillingPatch) (domain.Filling, error)
	HideFilling(ctx context.Context, fillingID string) error
}

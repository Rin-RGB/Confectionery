package fillings

import (
	"context"

	"SPOproject/internal/core/domain"
)

type fillingRepository interface {
	GetFillings(ctx context.Context, onlyActive bool) ([]domain.Filling, error)
	GetFillingByID(ctx context.Context, fillingID string) (domain.Filling, error)
	CreateFilling(ctx context.Context, filling domain.Filling) (domain.Filling, error)
	UpdateFilling(ctx context.Context, fillingID string, patch domain.FillingPatch) (domain.Filling, error)
	HideFilling(ctx context.Context, fillingID string) error
}

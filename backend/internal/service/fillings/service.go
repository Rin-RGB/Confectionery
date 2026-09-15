package fillings

import (
	"context"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
)

type Service struct {
	fillings fillingRepository
}

func NewService(fillings fillingRepository) *Service {
	return &Service{fillings: fillings}
}

func (service *Service) List(_ context.Context) ([]domain.Filling, error) {
	return nil, coreerrors.ErrNotImplemented
}

func (service *Service) GetByID(_ context.Context, _ string) (domain.Filling, error) {
	return domain.Filling{}, coreerrors.ErrNotImplemented
}

func (service *Service) Create(_ context.Context, _ domain.Filling) (domain.Filling, error) {
	return domain.Filling{}, coreerrors.ErrNotImplemented
}

func (service *Service) Update(_ context.Context, _ string, _ domain.FillingPatch) (domain.Filling, error) {
	return domain.Filling{}, coreerrors.ErrNotImplemented
}

func (service *Service) Hide(_ context.Context, _ string) error {
	return coreerrors.ErrNotImplemented
}

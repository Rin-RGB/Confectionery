package repository_fillings

import (
	"context"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (repository *Repository) List(_ context.Context, _ bool) ([]domain.Filling, error) {
	return nil, coreerrors.ErrNotImplemented
}

func (repository *Repository) GetByID(_ context.Context, _ string) (domain.Filling, error) {
	return domain.Filling{}, coreerrors.ErrNotImplemented
}

func (repository *Repository) Create(_ context.Context, _ domain.Filling) (domain.Filling, error) {
	return domain.Filling{}, coreerrors.ErrNotImplemented
}

func (repository *Repository) Update(_ context.Context, _ string, _ domain.FillingPatch) (domain.Filling, error) {
	return domain.Filling{}, coreerrors.ErrNotImplemented
}

func (repository *Repository) Hide(_ context.Context, _ string) error {
	return coreerrors.ErrNotImplemented
}

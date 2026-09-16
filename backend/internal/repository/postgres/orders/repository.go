package repository_orders

import (
	"context"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (repository *Repository) Create(_ context.Context, _ domain.Order) (domain.Order, error) {
	return domain.Order{}, coreerrors.ErrNotImplemented
}

func (repository *Repository) GetByID(_ context.Context, _ string) (domain.Order, error) {
	return domain.Order{}, coreerrors.ErrNotImplemented
}

func (repository *Repository) ListByUser(_ context.Context, _ string) ([]domain.Order, error) {
	return nil, coreerrors.ErrNotImplemented
}

func (repository *Repository) List(_ context.Context) ([]domain.Order, error) {
	return nil, coreerrors.ErrNotImplemented
}

func (repository *Repository) UpdateStatus(_ context.Context, _ string, _ domain.OrderStatus) error {
	return coreerrors.ErrNotImplemented
}

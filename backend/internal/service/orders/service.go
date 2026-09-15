package orders

import (
	"context"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
)

type Service struct {
	orders   orderRepository
	fillings fillingRepository
	payments paymentService
}

func NewService(
	orders orderRepository,
	fillings fillingRepository,
	payments paymentService,
) *Service {
	return &Service{
		orders:   orders,
		fillings: fillings,
		payments: payments,
	}
}

func (service *Service) CalculatePrice(_ context.Context, _ domain.PriceCalculation) (domain.Price, error) {
	return domain.Price{}, coreerrors.ErrNotImplemented
}

func (service *Service) Create(_ context.Context, _ string, _ domain.OrderDraft) (domain.Order, error) {
	return domain.Order{}, coreerrors.ErrNotImplemented
}

func (service *Service) GetMy(_ context.Context, _ string) ([]domain.Order, error) {
	return nil, coreerrors.ErrNotImplemented
}

func (service *Service) GetByID(_ context.Context, _, _ string, _ domain.UserRole) (domain.Order, error) {
	return domain.Order{}, coreerrors.ErrNotImplemented
}

func (service *Service) List(_ context.Context) ([]domain.Order, error) {
	return nil, coreerrors.ErrNotImplemented
}

func (service *Service) UpdateStatus(_ context.Context, _ string, _ domain.OrderStatus) error {
	return coreerrors.ErrNotImplemented
}

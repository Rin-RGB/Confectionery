package orders

import (
	"context"

	"SPOproject/internal/core/domain"
)

type orderRepository interface {
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
	GetByID(ctx context.Context, orderID string) (domain.Order, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Order, error)
	List(ctx context.Context) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error
}

type fillingRepository interface {
	GetByID(ctx context.Context, fillingID string) (domain.Filling, error)
}

type paymentService interface {
	Pay(ctx context.Context, orderID string, price domain.Price) (domain.Payment, error)
}

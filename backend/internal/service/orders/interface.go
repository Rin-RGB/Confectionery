package orders

import (
	"context"

	"SPOproject/internal/core/domain"
)

type orderRepository interface {
	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error)
	CreateOrderFillings(ctx context.Context, orderID string, fillings []domain.OrderFilling) error
	GetOrderByID(ctx context.Context, orderID string) (domain.Order, error)
	GetOrderFillings(ctx context.Context, orderID string) ([]domain.OrderFilling, error)
	GetOrdersByUser(ctx context.Context, userID string) ([]domain.Order, error)
	GetOrders(ctx context.Context, status *domain.OrderStatus) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error
}

type txManager interface {
	WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type fillingRepository interface {
	GetFillings(ctx context.Context, onlyActive bool) ([]domain.Filling, error)
	GetFillingByID(ctx context.Context, fillingID string) (domain.Filling, error)
}

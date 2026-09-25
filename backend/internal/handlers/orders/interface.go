package orders

import (
	"context"

	"SPOproject/internal/core/domain"
)

type orderService interface {
	CalculatePrice(ctx context.Context, calculation domain.PriceCalculation) (domain.Price, error)
	CreateOrder(ctx context.Context, userID string, draft domain.OrderDraft) (domain.Order, error)
	GetMyOrders(ctx context.Context, userID string) ([]domain.Order, error)
	GetOrderByID(ctx context.Context, userID, orderID string) (domain.Order, error)
	GetOrders(ctx context.Context, status *domain.OrderStatus) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error
}

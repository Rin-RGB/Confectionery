package orders

import (
	"context"

	"SPOproject/internal/core/domain"
)

type orderService interface {
	CalculatePrice(ctx context.Context, calculation domain.PriceCalculation) (domain.Price, error)
	Create(ctx context.Context, userID string, draft domain.OrderDraft) (domain.Order, error)
	GetMy(ctx context.Context, userID string) ([]domain.Order, error)
	GetByID(ctx context.Context, actorID, orderID string, role domain.UserRole) (domain.Order, error)
	List(ctx context.Context) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error
}

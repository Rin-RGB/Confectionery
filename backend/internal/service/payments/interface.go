package payments

import (
	"context"

	"SPOproject/internal/core/domain"
)

type paymentClient interface {
	Pay(ctx context.Context, orderID string, price domain.Price) (domain.Payment, error)
}

package payments

import (
	"context"

	"SPOproject/internal/core/domain"
)

type Service struct {
	client paymentClient
}

func NewService(client paymentClient) *Service {
	return &Service{client: client}
}

func (service *Service) Pay(ctx context.Context, orderID string, price domain.Price) (domain.Payment, error) {
	return service.client.Pay(ctx, orderID, price)
}

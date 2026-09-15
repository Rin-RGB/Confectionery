package payment

import (
	"context"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (client *Client) Pay(_ context.Context, _ string, _ domain.Price) (domain.Payment, error) {
	return domain.Payment{}, coreerrors.ErrNotImplemented
}

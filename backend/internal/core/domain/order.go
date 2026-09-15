package domain

import "time"

type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "new"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusReady     OrderStatus = "ready"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderDraft struct {
	FillingID  string
	WeightKG   int
	Decoration string
	Address    string
}

type PriceCalculation struct {
	FillingID string
	WeightKG  int
}

type Price struct {
	Amount   int64
	Currency string
}

type Order struct {
	ID         string
	UserID     string
	FillingID  string
	WeightKG   int
	Decoration string
	Address    string
	Price      Price
	Status     OrderStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

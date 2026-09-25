package domain

import "time"

type OrderStatus string

const (
	OrderStatusAccepted   OrderStatus = "accepted"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusReady      OrderStatus = "ready"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type OrderDraft struct {
	IdempotencyKey   string
	Fillings         []FillingWeight
	DecorationWishes string
	DeliveryAddress  string
}

type PriceCalculation struct {
	Fillings []FillingWeight
}

type FillingWeight struct {
	Name        string
	WeightGrams int
}

type Price struct {
	Amount float64
}

type Order struct {
	ID               string
	UserID           string
	IdempotencyKey   string
	Fillings         []OrderFilling
	WeightGrams      int
	DecorationWishes string
	DeliveryAddress  string
	TotalPrice       float64
	Status           OrderStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type OrderFilling struct {
	ID          string
	Name        string
	PricePerKG  float64
	WeightGrams int
}

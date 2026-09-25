package dto

import (
	"time"

	"SPOproject/internal/core/domain"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" format:"email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=3" minLength:"3" example:"password123"`
}

type LoginRequest = RegisterRequest

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type FillingResponse struct {
	ID          string  `json:"id" format:"uuid"`
	Name        string  `json:"name" maxLength:"150"`
	Description string  `json:"description"`
	Price       float64 `json:"price" minimum:"0.01" example:"1200"`
	ImageName   string  `json:"image_name" example:"filling1.png"`
	IsActive    bool    `json:"is_active"`
}

type CreateFillingRequest struct {
	Name        string  `json:"name" validate:"required,max=150" minLength:"1" maxLength:"150"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0" minimum:"0.01" example:"1200"`
	ImageName   string  `json:"image_name" example:"filling1.png"`
}

type UpdateFillingRequest struct {
	Name        *string  `json:"name" validate:"omitempty,max=150" minLength:"1" maxLength:"150"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" validate:"omitempty,gt=0" minimum:"0.01" example:"1200"`
	ImageName   *string  `json:"image_name" example:"filling1.png"`
	IsActive    *bool    `json:"is_active"`
}

type CalculatePriceRequest struct {
	Fillings []PriceFillingRequest `json:"fillings" validate:"required,min=1,dive" minItems:"1"`
}

type PriceFillingRequest struct {
	Name        string `json:"name" validate:"required" minLength:"1" example:"Шоколадная"`
	WeightGrams int    `json:"weight_grams" validate:"required,gt=0" minimum:"1" maximum:"7000" example:"1500"`
}

type PriceResponse struct {
	Amount float64 `json:"amount" minimum:"0" example:"1800"`
}

type CreateOrderRequest struct {
	Fillings         []PriceFillingRequest `json:"fillings" validate:"required,min=1,dive" minItems:"1"`
	DecorationWishes string                `json:"decoration_wishes"`
	DeliveryAddress  string                `json:"delivery_address" validate:"required" minLength:"1" example:"Москва, ул. Примерная, 10"`
}

type UpdateOrderStatusRequest struct {
	Status domain.OrderStatus `json:"status" validate:"required,oneof=accepted processing ready cancelled" enums:"accepted,processing,ready,cancelled"`
}

type OrderSummaryResponse struct {
	ID         string             `json:"id" format:"uuid"`
	TotalPrice float64            `json:"total_price" minimum:"0"`
	Status     domain.OrderStatus `json:"status" enums:"accepted,processing,ready,cancelled"`
}

type OrderFillingResponse struct {
	ID          string  `json:"id" format:"uuid"`
	Name        string  `json:"name"`
	PricePerKG  float64 `json:"price_per_kg" minimum:"0.01"`
	WeightGrams int     `json:"weight_grams" minimum:"1" maximum:"7000"`
}

type OrderResponse struct {
	ID               string                 `json:"id" format:"uuid"`
	UserID           string                 `json:"user_id" format:"uuid"`
	Fillings         []OrderFillingResponse `json:"fillings"`
	WeightGrams      int                    `json:"weight_grams" minimum:"1000" maximum:"7000"`
	DecorationWishes string                 `json:"decoration_wishes"`
	DeliveryAddress  string                 `json:"delivery_address"`
	TotalPrice       float64                `json:"total_price" minimum:"0"`
	Status           domain.OrderStatus     `json:"status" enums:"accepted,processing,ready,cancelled"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3"`
}

type LoginRequest = RegisterRequest

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type CreateFillingRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	ImageURL    string `json:"image_url"`
}

type UpdateFillingRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Price       *int64  `json:"price"`
	ImageURL    *string `json:"image_url"`
	IsActive    *bool   `json:"is_active"`
}

type CalculatePriceRequest struct {
	FillingID string `json:"filling_id"`
	WeightKG  int    `json:"weight_kg"`
}

type CreateOrderRequest struct {
	FillingID  string `json:"filling_id"`
	WeightKG   int    `json:"weight_kg"`
	Decoration string `json:"decoration"`
	Address    string `json:"address"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

package dto

import "SPOproject/internal/core/domain"

func FillingResponseFromDomain(filling domain.Filling) FillingResponse {
	return FillingResponse{
		ID:          filling.ID,
		Name:        filling.Name,
		Description: filling.Description,
		Price:       filling.Price,
		ImageName:   filling.ImageName,
		IsActive:    filling.IsActive,
	}
}

func FillingResponsesFromDomain(fillings []domain.Filling) []FillingResponse {
	result := make([]FillingResponse, len(fillings))
	for i := range fillings {
		result[i] = FillingResponseFromDomain(fillings[i])
	}
	return result
}

func CreateFillingRequestToDomain(request CreateFillingRequest) domain.Filling {
	return domain.Filling{
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
	}
}

func UpdateFillingRequestToDomain(request UpdateFillingRequest) domain.FillingPatch {
	return domain.FillingPatch{
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		ImageName:   request.ImageName,
		IsActive:    request.IsActive,
	}
}

func CalculatePriceRequestToDomain(request CalculatePriceRequest) domain.PriceCalculation {
	fillings := make([]domain.FillingWeight, len(request.Fillings))
	for i := range request.Fillings {
		fillings[i] = domain.FillingWeight{
			Name:        request.Fillings[i].Name,
			WeightGrams: request.Fillings[i].WeightGrams,
		}
	}

	return domain.PriceCalculation{Fillings: fillings}
}

func PriceResponseFromDomain(price domain.Price) PriceResponse {
	return PriceResponse{
		Amount: price.Amount,
	}
}

func CreateOrderRequestToDomain(request CreateOrderRequest) domain.OrderDraft {
	fillings := make([]domain.FillingWeight, len(request.Fillings))
	for i := range request.Fillings {
		fillings[i] = domain.FillingWeight{
			Name:        request.Fillings[i].Name,
			WeightGrams: request.Fillings[i].WeightGrams,
		}
	}

	return domain.OrderDraft{
		Fillings:         fillings,
		DecorationWishes: request.DecorationWishes,
		DeliveryAddress:  request.DeliveryAddress,
	}
}

func OrderSummaryResponsesFromDomain(orders []domain.Order) []OrderSummaryResponse {
	result := make([]OrderSummaryResponse, len(orders))
	for i := range orders {
		result[i] = OrderSummaryResponse{
			ID:         orders[i].ID,
			TotalPrice: orders[i].TotalPrice,
			Status:     orders[i].Status,
		}
	}
	return result
}

func OrderResponseFromDomain(order domain.Order) OrderResponse {
	fillings := make([]OrderFillingResponse, len(order.Fillings))
	for i := range order.Fillings {
		fillings[i] = OrderFillingResponse{
			ID:          order.Fillings[i].ID,
			Name:        order.Fillings[i].Name,
			PricePerKG:  order.Fillings[i].PricePerKG,
			WeightGrams: order.Fillings[i].WeightGrams,
		}
	}

	return OrderResponse{
		ID:               order.ID,
		UserID:           order.UserID,
		Fillings:         fillings,
		WeightGrams:      order.WeightGrams,
		DecorationWishes: order.DecorationWishes,
		DeliveryAddress:  order.DeliveryAddress,
		TotalPrice:       order.TotalPrice,
		Status:           order.Status,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	}
}

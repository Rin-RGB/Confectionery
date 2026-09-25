package orders

import (
	"context"
	"fmt"
	"math"
	"strings"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
	core_middleware "SPOproject/internal/core/http/middleware"

	"github.com/google/uuid"
)

type Service struct {
	orders    orderRepository
	fillings  fillingRepository
	txManager txManager
}

func NewService(
	orders orderRepository,
	fillings fillingRepository,
	txManager txManager,
) *Service {
	return &Service{
		orders:    orders,
		fillings:  fillings,
		txManager: txManager,
	}
}

func (s *Service) CalculatePrice(ctx context.Context, calculation domain.PriceCalculation) (domain.Price, error) {
	order, err := s.prepareOrder(ctx, calculation.Fillings)
	if err != nil {
		return domain.Price{}, err
	}

	return domain.Price{Amount: order.TotalPrice}, nil
}

func (s *Service) CreateOrder(ctx context.Context, userID string, draft domain.OrderDraft) (domain.Order, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return domain.Order{}, fmt.Errorf("invalid user id: %w", coreerrors.ErrInvalidRequest)
	}
	draft.IdempotencyKey = strings.TrimSpace(draft.IdempotencyKey)
	if _, err := uuid.Parse(draft.IdempotencyKey); err != nil {
		return domain.Order{}, fmt.Errorf("invalid idempotency key: %w", coreerrors.ErrInvalidRequest)
	}

	draft.DeliveryAddress = strings.TrimSpace(draft.DeliveryAddress)
	draft.DecorationWishes = strings.TrimSpace(draft.DecorationWishes)
	if draft.DeliveryAddress == "" {
		return domain.Order{}, coreerrors.ErrInvalidRequest
	}

	order, err := s.prepareOrder(ctx, draft.Fillings)
	if err != nil {
		return domain.Order{}, err
	}

	order.UserID = userID
	order.IdempotencyKey = draft.IdempotencyKey
	order.DecorationWishes = draft.DecorationWishes
	order.DeliveryAddress = draft.DeliveryAddress
	order.Status = domain.OrderStatusAccepted

	var createdOrder domain.Order
	err = s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		createdOrder, err = s.orders.CreateOrder(txCtx, order)
		if err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		savedFillings, err := s.orders.GetOrderFillings(txCtx, createdOrder.ID)
		if err != nil {
			return fmt.Errorf("get saved order fillings: %w", err)
		}
		if len(savedFillings) > 0 {
			createdOrder.Fillings = savedFillings
			return nil
		}

		if err = s.orders.CreateOrderFillings(txCtx, createdOrder.ID, order.Fillings); err != nil {
			return fmt.Errorf("create order fillings: %w", err)
		}
		createdOrder.Fillings = order.Fillings

		return nil
	})
	if err != nil {
		return domain.Order{}, err
	}

	return createdOrder, nil
}

func (s *Service) prepareOrder(
	ctx context.Context,
	requestedFillings []domain.FillingWeight,
) (domain.Order, error) {
	if len(requestedFillings) == 0 {
		return domain.Order{}, coreerrors.ErrInvalidRequest
	}

	availableFillings, err := s.fillings.GetFillings(ctx, true)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get fillings: %w", err)
	}

	fillingsByName := make(map[string]domain.Filling, len(availableFillings))
	for _, filling := range availableFillings {
		fillingsByName[filling.Name] = filling
	}

	orderFillings := make([]domain.OrderFilling, 0, len(requestedFillings))
	indexesByID := make(map[string]int, len(requestedFillings))
	var weightedPrice float64
	var totalWeightGrams int
	for _, requestedFilling := range requestedFillings {
		name := strings.TrimSpace(requestedFilling.Name)
		if name == "" || requestedFilling.WeightGrams <= 0 {
			return domain.Order{}, coreerrors.ErrInvalidRequest
		}

		filling, found := fillingsByName[name]
		if !found {
			return domain.Order{}, fmt.Errorf("wrong filling in request: %w", coreerrors.ErrInvalidRequest)
		}

		if index, exists := indexesByID[filling.ID]; exists {
			orderFillings[index].WeightGrams += requestedFilling.WeightGrams
		} else {
			indexesByID[filling.ID] = len(orderFillings)
			orderFillings = append(orderFillings, domain.OrderFilling{
				ID:          filling.ID,
				Name:        filling.Name,
				PricePerKG:  filling.Price,
				WeightGrams: requestedFilling.WeightGrams,
			})
		}

		totalWeightGrams += requestedFilling.WeightGrams
		weightedPrice += filling.Price * float64(requestedFilling.WeightGrams)
	}
	if totalWeightGrams < 1000 || totalWeightGrams > 7000 {
		return domain.Order{}, coreerrors.ErrInvalidRequest
	}

	return domain.Order{
		Fillings:    orderFillings,
		WeightGrams: totalWeightGrams,
		TotalPrice:  math.Round(weightedPrice/1000*100) / 100,
	}, nil
}

func (s *Service) GetMyOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("invalid user id: %w", coreerrors.ErrInvalidRequest)
	}

	orders, err := s.orders.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user orders: %w", err)
	}

	return orders, nil
}

func (s *Service) GetOrderByID(ctx context.Context, userID, orderID string) (domain.Order, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return domain.Order{}, fmt.Errorf("invalid user id: %w", coreerrors.ErrInvalidRequest)
	}
	if _, err := uuid.Parse(orderID); err != nil {
		return domain.Order{}, fmt.Errorf("invalid order id: %w", coreerrors.ErrInvalidRequest)
	}

	order, err := s.orders.GetOrderByID(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}

	userRole, err := core_middleware.GetRoleFromCtx(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("failed to get user role: %w", err)
	}
	if userRole != domain.RoleAdmin {
		if order.UserID != userID {
			return domain.Order{}, fmt.Errorf("this is not your order: %w", coreerrors.ErrForbidden)
		}
	}

	fillings, err := s.orders.GetOrderFillings(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order fillings: %w", err)
	}
	order.Fillings = fillings

	return order, nil
}

func (s *Service) GetOrders(ctx context.Context, status *domain.OrderStatus) ([]domain.Order, error) {
	if status != nil && !isValidOrderStatus(*status) {
		return nil, fmt.Errorf("invalid order status: %w", coreerrors.ErrInvalidRequest)
	}

	orders, err := s.orders.GetOrders(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("get orders: %w", err)
	}

	return orders, nil
}

func (s *Service) UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	if _, err := uuid.Parse(orderID); err != nil {
		return fmt.Errorf("invalid order id: %w", coreerrors.ErrInvalidRequest)
	}
	if !isValidOrderStatus(status) {
		return fmt.Errorf("invalid order status: %w", coreerrors.ErrInvalidRequest)
	}

	if err := s.orders.UpdateStatus(ctx, orderID, status); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	return nil
}

func isValidOrderStatus(status domain.OrderStatus) bool {
	switch status {
	case domain.OrderStatusAccepted,
		domain.OrderStatusProcessing,
		domain.OrderStatusReady,
		domain.OrderStatusCancelled:
		return true
	default:
		return false
	}
}

package orders

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
	core_middleware "SPOproject/internal/core/http/middleware"

	"github.com/google/uuid"
)

const (
	tempPhotoDirectory = "temp_photo"
	maxReferencePhotos = 3
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
	if len(draft.ReferencePhotos) > maxReferencePhotos {
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
	var createdPhotoPaths []string
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
			createdOrder.ReferencePhotos, err = s.orders.GetOrderPhotoNames(txCtx, createdOrder.ID)
			if err != nil {
				return fmt.Errorf("get saved order photo names: %w", err)
			}
			return nil
		}

		if err = s.orders.CreateOrderFillings(txCtx, createdOrder.ID, order.Fillings); err != nil {
			return fmt.Errorf("create order fillings: %w", err)
		}
		createdOrder.Fillings = order.Fillings

		photoNames, photoPaths, err := saveReferencePhotos(createdOrder.ID, draft.ReferencePhotos)
		createdPhotoPaths = photoPaths
		if err != nil {
			return fmt.Errorf("save reference photos: %w", err)
		}
		if err = s.orders.CreateOrderPhotos(txCtx, createdOrder.ID, photoNames); err != nil {
			return fmt.Errorf("create order photos: %w", err)
		}
		createdOrder.ReferencePhotos = photoNames

		return nil
	})
	if err != nil {
		for _, photoPath := range createdPhotoPaths {
			if removeErr := os.Remove(photoPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				err = errors.Join(err, fmt.Errorf("remove reference photo after rollback: %w", removeErr))
			}
		}
		return domain.Order{}, err
	}

	return createdOrder, nil
}

func saveReferencePhotos(orderID string, photos []domain.OrderPhotoUpload) ([]string, []string, error) {
	if len(photos) == 0 {
		return make([]string, 0), nil, nil
	}
	if err := os.MkdirAll(tempPhotoDirectory, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create temp photo directory: %w", err)
	}

	photoNames := make([]string, 0, len(photos))
	photoPaths := make([]string, 0, len(photos))
	for index, photo := range photos {
		photoName := orderID + "_" + strconv.Itoa(index+1) + photo.Extension
		photoPath := filepath.Join(tempPhotoDirectory, photoName)
		photoFile, err := os.OpenFile(photoPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return photoNames, photoPaths, fmt.Errorf("create reference photo: %w", err)
		}
		photoNames = append(photoNames, photoName)
		photoPaths = append(photoPaths, photoPath)

		if _, err = photoFile.Write(photo.Data); err != nil {
			closeErr := photoFile.Close()
			return photoNames, photoPaths, errors.Join(fmt.Errorf("write reference photo: %w", err), closeErr)
		}
		if err = photoFile.Close(); err != nil {
			return photoNames, photoPaths, fmt.Errorf("close reference photo: %w", err)
		}
	}

	return photoNames, photoPaths, nil
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
	order, err := s.getAuthorizedOrder(ctx, userID, orderID)
	if err != nil {
		return domain.Order{}, err
	}

	fillings, err := s.orders.GetOrderFillings(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order fillings: %w", err)
	}
	order.Fillings = fillings
	order.ReferencePhotos, err = s.orders.GetOrderPhotoNames(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order photo names: %w", err)
	}

	return order, nil
}

func (s *Service) getAuthorizedOrder(ctx context.Context, userID, orderID string) (domain.Order, error) {
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

	return order, nil
}

func (s *Service) GetOrderPhoto(ctx context.Context, userID, orderID, photoName string) (string, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return "", fmt.Errorf("invalid user id: %w", coreerrors.ErrInvalidRequest)
	}
	if _, err := uuid.Parse(orderID); err != nil {
		return "", fmt.Errorf("invalid order id: %w", coreerrors.ErrInvalidRequest)
	}
	photoName = strings.TrimSpace(photoName)
	if photoName == "" || filepath.Base(photoName) != photoName {
		return "", fmt.Errorf("invalid order photo name: %w", coreerrors.ErrInvalidRequest)
	}

	ownerID, err := s.orders.GetOrderPhotoOwnerID(ctx, orderID, photoName)
	if err != nil {
		return "", fmt.Errorf("get order photo owner: %w", err)
	}
	userRole, err := core_middleware.GetRoleFromCtx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	if userRole != domain.RoleAdmin && ownerID != userID {
		return "", fmt.Errorf("this is not your order photo: %w", coreerrors.ErrForbidden)
	}

	photoPath := filepath.Join(tempPhotoDirectory, photoName)
	fileInfo, err := os.Stat(photoPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", coreerrors.ErrOrderPhotoNotFound
		}
		return "", fmt.Errorf("stat order photo: %w", err)
	}
	if !fileInfo.Mode().IsRegular() {
		return "", coreerrors.ErrOrderPhotoNotFound
	}

	return photoPath, nil
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

func (s *Service) UpdateStatus(
	ctx context.Context,
	orderID string,
	status domain.OrderStatus,
	declineReason *string,
) error {
	if _, err := uuid.Parse(orderID); err != nil {
		return fmt.Errorf("invalid order id: %w", coreerrors.ErrInvalidRequest)
	}
	if !isValidOrderStatus(status) {
		return fmt.Errorf("invalid order status: %w", coreerrors.ErrInvalidRequest)
	}

	if err := s.orders.UpdateStatus(ctx, orderID, status, declineReason); err != nil {
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

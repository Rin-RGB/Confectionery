package orders

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
	"SPOproject/internal/core/http/middleware"
	core_request "SPOproject/internal/core/http/request"
	"SPOproject/internal/core/http/response"
	"SPOproject/internal/core/http/server"
	"SPOproject/internal/handlers/dto"
)

const (
	maxReferencePhotoSize = 8 << 20
	maxReferencePhotos    = 3
	maxOrderBodySize      = maxReferencePhotoSize*maxReferencePhotos + 1<<20
)

type Handler struct {
	service   orderService
	provider  core_middleware.Parser
	validator core_request.Validator
}

func NewHandler(service orderService, provider core_middleware.Parser, validator core_request.Validator) *Handler {
	return &Handler{service: service, provider: provider, validator: validator}
}

func (h *Handler) Routes() []core_server.Route {
	userOnly := []core_middleware.Middleware{core_middleware.Auth(h.provider, domain.RoleUser)}
	adminOnly := []core_middleware.Middleware{core_middleware.Auth(h.provider, domain.RoleAdmin)}
	userOrAdmin := []core_middleware.Middleware{core_middleware.Auth(h.provider, domain.RoleUser, domain.RoleAdmin)}
	return []core_server.Route{
		{Method: http.MethodPost, Path: "/orders/price", Handler: h.CalculatePrice},
		{Method: http.MethodPost, Path: "/orders", Handler: h.CreateOrder, Middlewares: userOnly},
		{Method: http.MethodGet, Path: "/orders/my", Handler: h.GetMyOrders, Middlewares: userOnly},
		{Method: http.MethodGet, Path: "/orders/{orderId}", Handler: h.GetOrder, Middlewares: userOrAdmin},
		{Method: http.MethodGet, Path: "/orders", Handler: h.GetOrders, Middlewares: adminOnly},
		{Method: http.MethodGet, Path: "/orders/{orderId}/photos/{fileName}", Handler: h.GetOrderPhoto, Middlewares: userOrAdmin},
		{Method: http.MethodPatch, Path: "/orders/{orderId}/status", Handler: h.UpdateOrderStatus, Middlewares: adminOnly},
	}
}

// CalculatePrice рассчитывает стоимость заказа.
// @Summary Расчёт стоимости заказа
// @Description Суммарный вес начинок должен составлять от 1000 до 7000 граммов.
// @Tags orders
// @Accept json
// @Produce json
// @Param request body dto.CalculatePriceRequest true "Начинки и их вес"
// @Success 200 {object} dto.PriceResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/orders/price [post]
func (h *Handler) CalculatePrice(writer *core_response.Writer, request *http.Request) {
	var body dto.CalculatePriceRequest
	if err := core_request.DecodeAndValidate(request, &body, h.validator); err != nil {
		writer.ErrorResponse(err)
		return
	}

	price, err := h.service.CalculatePrice(request.Context(), dto.CalculatePriceRequestToDomain(body))
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, dto.PriceResponseFromDomain(price))
}

// CreateOrder создаёт заказ пользователя.
// @Summary Создание заказа
// @Description Цена рассчитывается на сервере. При ретраях клиент повторно использует тот же Idempotency-Key.
// @Tags orders
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string true "UUID идемпотентности, созданный клиентом" Format(uuid)
// @Param fillings formData string true "JSON-массив начинок и их веса"
// @Param decoration_wishes formData string false "Пожелания по оформлению"
// @Param delivery_address formData string true "Адрес доставки"
// @Param reference_photos formData file false "До трёх фотографий-референсов"
// @Success 201 {object} dto.OrderResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 403 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/orders [post]
func (h *Handler) CreateOrder(writer *core_response.Writer, request *http.Request) {
	userID, err := core_middleware.GetUserIdFromCtx(request.Context())
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, maxOrderBodySize)
	body, err := h.parseCreateOrderRequest(request)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}
	draft := dto.CreateOrderRequestToDomain(body)
	draft.IdempotencyKey = request.Header.Get("Idempotency-Key")

	order, err := h.service.CreateOrder(
		request.Context(),
		userID.String(),
		draft,
	)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusCreated, dto.OrderResponseFromDomain(order))
}

func (h *Handler) parseCreateOrderRequest(request *http.Request) (dto.CreateOrderRequest, error) {
	if err := request.ParseMultipartForm(maxReferencePhotoSize); err != nil {
		return dto.CreateOrderRequest{}, fmt.Errorf("parse order multipart form: %v: %w", err, coreerrors.ErrInvalidRequest)
	}
	defer request.MultipartForm.RemoveAll()

	var fillings []dto.PriceFillingRequest
	if err := json.Unmarshal([]byte(request.FormValue("fillings")), &fillings); err != nil {
		return dto.CreateOrderRequest{}, fmt.Errorf("parse order fillings: %v: %w", err, coreerrors.ErrInvalidRequest)
	}

	photoHeaders := request.MultipartForm.File["reference_photos"]
	if len(photoHeaders) > maxReferencePhotos {
		return dto.CreateOrderRequest{}, fmt.Errorf("too many reference photos: %w", coreerrors.ErrInvalidRequest)
	}

	photos := make([]domain.OrderPhotoUpload, 0, len(photoHeaders))
	for _, photoHeader := range photoHeaders {
		photo, err := readReferencePhoto(photoHeader)
		if err != nil {
			return dto.CreateOrderRequest{}, err
		}
		photos = append(photos, photo)
	}

	body := dto.CreateOrderRequest{
		Fillings:         fillings,
		DecorationWishes: request.FormValue("decoration_wishes"),
		DeliveryAddress:  request.FormValue("delivery_address"),
		ReferencePhotos:  photos,
	}
	if err := h.validator.Struct(body); err != nil {
		return dto.CreateOrderRequest{}, fmt.Errorf("validate order request: %v: %w", err, coreerrors.ErrInvalidRequest)
	}

	return body, nil
}

func readReferencePhoto(header *multipart.FileHeader) (domain.OrderPhotoUpload, error) {
	file, err := header.Open()
	if err != nil {
		return domain.OrderPhotoUpload{}, fmt.Errorf("open reference photo: %v: %w", err, coreerrors.ErrInvalidRequest)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxReferencePhotoSize+1))
	if err != nil {
		return domain.OrderPhotoUpload{}, fmt.Errorf("read reference photo: %v: %w", err, coreerrors.ErrInvalidRequest)
	}
	if len(data) == 0 || len(data) > maxReferencePhotoSize {
		return domain.OrderPhotoUpload{}, fmt.Errorf("invalid reference photo size: %w", coreerrors.ErrInvalidRequest)
	}

	var extension string
	switch http.DetectContentType(data) {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"
	default:
		return domain.OrderPhotoUpload{}, fmt.Errorf("reference photo has unsupported format: %w", coreerrors.ErrInvalidRequest)
	}

	return domain.OrderPhotoUpload{Data: data, Extension: extension}, nil
}

// GetMyOrders возвращает заказы текущего пользователя.
// @Summary Получение заказов текущего пользователя
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.OrderSummaryResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 403 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/orders/my [get]
func (h *Handler) GetMyOrders(writer *core_response.Writer, request *http.Request) {
	userID, err := core_middleware.GetUserIdFromCtx(request.Context())
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	orders, err := h.service.GetMyOrders(request.Context(), userID.String())
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, dto.OrderSummaryResponsesFromDomain(orders))
}

// GetOrder возвращает полный заказ пользователя.
// @Summary Получение заказа
// @Description Пользователь может получить только собственный заказ.
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param orderId path string true "UUID заказа" Format(uuid)
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 403 {object} core_response.ErrorResponse
// @Failure 404 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/orders/{orderId} [get]
func (h *Handler) GetOrder(writer *core_response.Writer, request *http.Request) {
	userID, err := core_middleware.GetUserIdFromCtx(request.Context())
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	order, err := h.service.GetOrderByID(
		request.Context(),
		userID.String(),
		request.PathValue("orderId"),
	)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, dto.OrderResponseFromDomain(order))
}

// GetOrderPhoto возвращает фотографию-референс заказа.
// @Summary Получение фотографии заказа
// @Description Пользователь может получить фотографию только собственного заказа. Администратор может получить фотографию любого заказа.
// @Tags orders
// @Produce image/jpeg,image/png,image/gif,image/webp
// @Security BearerAuth
// @Param orderId path string true "UUID заказа" Format(uuid)
// @Param fileName path string true "Имя фотографии из reference_photos"
// @Success 200 {file} file
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 403 {object} core_response.ErrorResponse
// @Failure 404 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/orders/{orderId}/photos/{fileName} [get]
func (h *Handler) GetOrderPhoto(writer *core_response.Writer, request *http.Request) {
	userID, err := core_middleware.GetUserIdFromCtx(request.Context())
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	photoPath, err := h.service.GetOrderPhoto(
		request.Context(),
		userID.String(),
		request.PathValue("orderId"),
		request.PathValue("fileName"),
	)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	http.ServeFile(writer, request, photoPath)
}

// GetOrders возвращает заказы с необязательной фильтрацией по статусу.
// @Summary Получение всех заказов
// @Description Доступно только администратору. Без type возвращаются все заказы.
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param type query string false "Фильтр по статусу заказа" Enums(accepted,processing,ready,cancelled)
// @Success 200 {array} dto.OrderSummaryResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 403 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/orders [get]
func (h *Handler) GetOrders(writer *core_response.Writer, request *http.Request) {
	var status *domain.OrderStatus
	query := request.URL.Query()
	if _, exists := query["type"]; exists {
		orderStatus := domain.OrderStatus(query.Get("type"))
		status = &orderStatus
	}

	orders, err := h.service.GetOrders(request.Context(), status)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, dto.OrderSummaryResponsesFromDomain(orders))
}

// UpdateOrderStatus изменяет статус заказа.
// @Summary Изменение статуса заказа
// @Description Доступно только администратору.
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param orderId path string true "UUID заказа" Format(uuid)
// @Param request body dto.UpdateOrderStatusRequest true "Новый статус"
// @Success 204
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 403 {object} core_response.ErrorResponse
// @Failure 404 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/orders/{orderId}/status [patch]
func (h *Handler) UpdateOrderStatus(writer *core_response.Writer, request *http.Request) {
	var body dto.UpdateOrderStatusRequest
	if err := core_request.DecodeAndValidate(request, &body, h.validator); err != nil {
		writer.ErrorResponse(err)
		return
	}

	if err := h.service.UpdateStatus(
		request.Context(),
		request.PathValue("orderId"),
		body.Status,
		body.DeclineReason,
	); err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

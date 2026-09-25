package orders

import (
	"net/http"

	"SPOproject/internal/core/domain"
	"SPOproject/internal/core/http/middleware"
	core_request "SPOproject/internal/core/http/request"
	"SPOproject/internal/core/http/response"
	"SPOproject/internal/core/http/server"
	"SPOproject/internal/handlers/dto"
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
	return []core_server.Route{
		{Method: http.MethodPost, Path: "/orders/price", Handler: h.CalculatePrice},
		{Method: http.MethodPost, Path: "/orders", Handler: h.CreateOrder, Middlewares: userOnly},
		{Method: http.MethodGet, Path: "/orders/my", Handler: h.GetMyOrders, Middlewares: userOnly},
		{Method: http.MethodGet, Path: "/orders/{orderId}", Handler: h.GetOrder, Middlewares: userOnly},
		{Method: http.MethodGet, Path: "/orders", Handler: h.GetOrders, Middlewares: adminOnly},
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
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string true "UUID идемпотентности, созданный клиентом" Format(uuid)
// @Param request body dto.CreateOrderRequest true "Данные заказа"
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

	var body dto.CreateOrderRequest
	if err = core_request.DecodeAndValidate(request, &body, h.validator); err != nil {
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
	); err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

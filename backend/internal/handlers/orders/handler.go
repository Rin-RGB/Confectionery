package orders

import (
	"net/http"

	"SPOproject/internal/core/domain"
	"SPOproject/internal/core/http/middleware"
	"SPOproject/internal/core/http/response"
	"SPOproject/internal/core/http/server"
)

type Handler struct {
	service  orderService
	provider core_middleware.Parser
}

func NewHandler(service orderService, provider core_middleware.Parser) *Handler {
	return &Handler{service: service, provider: provider}
}

func (h *Handler) Routes() []core_server.Route {
	userOnly := []core_middleware.Middleware{core_middleware.Auth(h.provider, domain.RoleUser)}
	adminOnly := []core_middleware.Middleware{core_middleware.Auth(h.provider, domain.RoleAdmin)}
	return []core_server.Route{
		{Method: http.MethodPost, Path: "/orders/price", Handler: h.CalculatePrice, Middlewares: userOnly},
		{Method: http.MethodPost, Path: "/orders", Handler: h.CreateOrder, Middlewares: userOnly},
		{Method: http.MethodGet, Path: "/orders/my", Handler: h.GetMyOrders, Middlewares: userOnly},
		{Method: http.MethodGet, Path: "/orders/{orderId}", Handler: h.GetOrder, Middlewares: userOnly},
		{Method: http.MethodGet, Path: "/orders", Handler: h.GetOrders, Middlewares: adminOnly},
		{Method: http.MethodPatch, Path: "/orders/{orderId}/status", Handler: h.UpdateOrderStatus, Middlewares: adminOnly},
	}
}

func (h *Handler) CalculatePrice(writer *core_response.Writer, r *http.Request) {
}

func (h *Handler) CreateOrder(writer *core_response.Writer, r *http.Request) {
}

func (h *Handler) GetMyOrders(writer *core_response.Writer, r *http.Request) {
}

func (h *Handler) GetOrder(writer *core_response.Writer, r *http.Request) {
}

func (h *Handler) GetOrders(writer *core_response.Writer, r *http.Request) {
}

func (h *Handler) UpdateOrderStatus(writer *core_response.Writer, r *http.Request) {
}

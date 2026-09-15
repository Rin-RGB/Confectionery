package fillings

import (
	core_response "SPOproject/internal/core/http/response"
	"net/http"

	"SPOproject/internal/core/domain"
	"SPOproject/internal/core/http/middleware"
	"SPOproject/internal/core/http/server"
)

type Handler struct {
	service  fillingService
	provider core_middleware.Parser
}

func NewHandler(service fillingService, provider core_middleware.Parser) *Handler {
	return &Handler{service: service, provider: provider}
}

func (h *Handler) Routes() []core_server.Route {
	adminOnly := []core_middleware.Middleware{core_middleware.Auth(h.provider, domain.RoleAdmin)}
	return []core_server.Route{
		{Method: http.MethodGet, Path: "/fillings", Handler: h.GetFillings},
		{Method: http.MethodGet, Path: "/fillings/{fillingId}", Handler: h.GetFilling},
		{Method: http.MethodPost, Path: "/fillings", Handler: h.CreateFilling, Middlewares: adminOnly},
		{Method: http.MethodPatch, Path: "/fillings/{fillingId}", Handler: h.UpdateFilling, Middlewares: adminOnly},
		{Method: http.MethodDelete, Path: "/fillings/{fillingId}", Handler: h.DeleteFilling, Middlewares: adminOnly},
	}
}

func (h *Handler) GetFillings(writer *core_response.Writer, r *http.Request) {
}

func (h *Handler) GetFilling(writer *core_response.Writer, r *http.Request) {
}

func (h *Handler) CreateFilling(writer *core_response.Writer, r *http.Request) {
}

func (h *Handler) UpdateFilling(writer *core_response.Writer, r *http.Request) {
}

func (h *Handler) DeleteFilling(writer *core_response.Writer, r *http.Request) {
}

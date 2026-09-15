package auth

import (
	"fmt"
	"net/http"

	coreauth "SPOproject/internal/core/auth"
	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
	"SPOproject/internal/core/http/middleware"
	"SPOproject/internal/core/http/request"
	"SPOproject/internal/core/http/response"
	"SPOproject/internal/core/http/server"
	"SPOproject/internal/handlers/dto"
)

type Handler struct {
	service   authService
	provider  tokenProvider
	validator core_request.Validator
}

func NewHandler(service authService, provider tokenProvider, validator core_request.Validator) *Handler {
	return &Handler{
		service:   service,
		provider:  provider,
		validator: validator,
	}
}

func (h *Handler) Routes() []core_server.Route {
	return []core_server.Route{
		{Method: http.MethodPost, Path: "/register", Handler: h.Register},
		{Method: http.MethodPost, Path: "/login", Handler: h.Login},
		{Method: http.MethodPost, Path: "/refresh", Handler: h.Refresh},
		{
			Method:      http.MethodPost,
			Path:        "/logout",
			Handler:     h.Logout,
			Middlewares: []core_middleware.Middleware{core_middleware.Auth(h.provider, domain.RoleUser)},
		},
	}
}

func (h *Handler) Register(writer *core_response.Writer, httpRequest *http.Request) {
	var body dto.RegisterRequest
	if err := core_request.DecodeAndValidate(httpRequest, &body, h.validator); err != nil {
		writer.ErrorResponse(err)
		return
	}

	user, err := h.service.Register(httpRequest.Context(), domain.Credentials{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	tokens, err := h.newTokenPair(user)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusCreated, tokens)
}

func (h *Handler) Login(writer *core_response.Writer, httpRequest *http.Request) {
	var body dto.LoginRequest
	if err := core_request.DecodeAndValidate(httpRequest, &body, h.validator); err != nil {
		writer.ErrorResponse(err)
		return
	}

	user, err := h.service.Login(httpRequest.Context(), domain.Credentials{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	tokens, err := h.newTokenPair(user)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, tokens)
}

func (h *Handler) newTokenPair(user domain.User) (dto.TokenPairResponse, error) {
	accessToken, err := h.provider.NewToken(user, coreauth.AccessType)
	if err != nil {
		return dto.TokenPairResponse{}, fmt.Errorf("create access token: %w", err)
	}
	refreshToken, err := h.provider.NewToken(user, coreauth.RefreshType)
	if err != nil {
		return dto.TokenPairResponse{}, fmt.Errorf("create refresh token: %w", err)
	}

	return dto.TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) Refresh(writer *core_response.Writer, r *http.Request) {
	writer.ErrorResponse(coreerrors.ErrNotImplemented)
}

func (h *Handler) Logout(writer *core_response.Writer, r *http.Request) {
	writer.ErrorResponse(coreerrors.ErrNotImplemented)
}

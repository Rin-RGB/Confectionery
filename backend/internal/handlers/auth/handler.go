package auth

import (
	"net/http"

	"SPOproject/internal/core/domain"
	"SPOproject/internal/core/http/request"
	"SPOproject/internal/core/http/response"
	"SPOproject/internal/core/http/server"
	"SPOproject/internal/handlers/dto"
)

type Handler struct {
	service   authService
	validator core_request.Validator
}

func NewHandler(service authService, validator core_request.Validator) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
	}
}

func (h *Handler) Routes() []core_server.Route {
	return []core_server.Route{
		{Method: http.MethodPost, Path: "/register", Handler: h.Register},
		{Method: http.MethodPost, Path: "/login", Handler: h.Login},
		{Method: http.MethodPost, Path: "/refresh", Handler: h.Refresh},
		{Method: http.MethodPost, Path: "/logout", Handler: h.Logout},
	}
}

// Register регистрирует нового пользователя.
// @Summary Регистрация пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Учётные данные"
// @Success 201 {object} dto.TokenPairResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 409 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/register [post]
func (h *Handler) Register(writer *core_response.Writer, httpRequest *http.Request) {
	var body dto.RegisterRequest
	if err := core_request.DecodeAndValidate(httpRequest, &body, h.validator); err != nil {
		writer.ErrorResponse(err)
		return
	}

	tokensPair, err := h.service.Register(httpRequest.Context(), domain.Credentials{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusCreated, dto.TokenPairResponse{
		AccessToken:  tokensPair.AccessToken,
		RefreshToken: tokensPair.RefreshToken,
	})
}

// Login выполняет вход пользователя.
// @Summary Вход пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Учётные данные"
// @Success 200 {object} dto.TokenPairResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/login [post]
func (h *Handler) Login(writer *core_response.Writer, httpRequest *http.Request) {
	var body dto.LoginRequest
	if err := core_request.DecodeAndValidate(httpRequest, &body, h.validator); err != nil {
		writer.ErrorResponse(err)
		return
	}

	tokensPair, err := h.service.Login(httpRequest.Context(), domain.Credentials{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, dto.TokenPairResponse{
		AccessToken:  tokensPair.AccessToken,
		RefreshToken: tokensPair.RefreshToken,
	})
}

// Refresh обновляет пару JWT-токенов.
// @Summary Обновление пары токенов
// @Description Валидирует refresh-токен, инвалидирует использованную сессию и возвращает новую пару токенов.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh-токен"
// @Success 200 {object} dto.TokenPairResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/refresh [post]
func (h *Handler) Refresh(writer *core_response.Writer, httpRequest *http.Request) {
	var body dto.RefreshRequest
	if err := core_request.DecodeAndValidate(httpRequest, &body, h.validator); err != nil {
		writer.ErrorResponse(err)
		return
	}

	tokensPair, err := h.service.Refresh(httpRequest.Context(), body.RefreshToken)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, dto.TokenPairResponse{
		AccessToken:  tokensPair.AccessToken,
		RefreshToken: tokensPair.RefreshToken,
	})
}

// Logout отзывает refresh-сессию пользователя.
// @Summary Выход пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh-токен"
// @Success 204
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/logout [post]
func (h *Handler) Logout(writer *core_response.Writer, httpRequest *http.Request) {
	var body dto.RefreshRequest
	if err := core_request.DecodeAndValidate(httpRequest, &body, h.validator); err != nil {
		writer.ErrorResponse(err)
		return
	}

	if err := h.service.Logout(httpRequest.Context(), body.RefreshToken); err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

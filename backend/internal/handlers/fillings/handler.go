package fillings

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"strconv"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
	"SPOproject/internal/core/http/middleware"
	core_request "SPOproject/internal/core/http/request"
	core_response "SPOproject/internal/core/http/response"
	"SPOproject/internal/core/http/server"
	"SPOproject/internal/handlers/dto"
)

const (
	maxFillingImageSize  = 8 << 20
	maxMultipartBodySize = 10 << 20
)

type Handler struct {
	service   fillingService
	provider  core_middleware.Parser
	validator core_request.Validator
}

func NewHandler(service fillingService, provider core_middleware.Parser, validator core_request.Validator) *Handler {
	return &Handler{service: service, provider: provider, validator: validator}
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

// GetFillings возвращает список активных начинок.
// @Summary Получение активных начинок
// @Tags fillings
// @Produce json
// @Success 200 {array} dto.FillingResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/fillings [get]
func (h *Handler) GetFillings(writer *core_response.Writer, request *http.Request) {
	fillings, err := h.service.GetFillings(request.Context())
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, dto.FillingResponsesFromDomain(fillings))
}

// GetFilling возвращает активную начинку по UUID.
// @Summary Получение начинки
// @Tags fillings
// @Produce json
// @Param fillingId path string true "UUID начинки" Format(uuid)
// @Success 200 {object} dto.FillingResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 404 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/fillings/{fillingId} [get]
func (h *Handler) GetFilling(writer *core_response.Writer, request *http.Request) {
	filling, err := h.service.GetFillingByID(request.Context(), request.PathValue("fillingId"))
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, dto.FillingResponseFromDomain(filling))
}

// CreateFilling создаёт начинку.
// @Summary Создание начинки
// @Description Доступно только администратору.
// @Tags fillings
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Название начинки"
// @Param description formData string false "Описание начинки"
// @Param price formData number true "Цена за килограмм"
// @Param image formData file true "Изображение начинки в формате PNG (до 8 МБ)"
// @Success 201 {object} dto.FillingResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 403 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/fillings [post]
func (h *Handler) CreateFilling(writer *core_response.Writer, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maxMultipartBodySize)
	body, err := h.parseCreateFillingRequest(request)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	filling, err := h.service.CreateFilling(
		request.Context(),
		dto.CreateFillingRequestToDomain(body),
		body.Image,
	)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusCreated, dto.FillingResponseFromDomain(filling))
}

func (h *Handler) parseCreateFillingRequest(request *http.Request) (dto.CreateFillingRequest, error) {
	if err := request.ParseMultipartForm(maxFillingImageSize); err != nil {
		return dto.CreateFillingRequest{}, fmt.Errorf(
			"parse multipart form: %v: %w",
			err,
			coreerrors.ErrInvalidRequest,
		)
	}
	defer request.MultipartForm.RemoveAll()

	price, err := strconv.ParseFloat(request.FormValue("price"), 64)
	if err != nil {
		return dto.CreateFillingRequest{}, fmt.Errorf(
			"parse filling price: %v: %w",
			err,
			coreerrors.ErrInvalidRequest,
		)
	}

	imageFile, _, err := request.FormFile("image")
	if err != nil {
		return dto.CreateFillingRequest{}, fmt.Errorf(
			"get filling image: %v: %w",
			err,
			coreerrors.ErrInvalidRequest,
		)
	}
	defer imageFile.Close()

	image, err := io.ReadAll(io.LimitReader(imageFile, maxFillingImageSize+1))
	if err != nil {
		return dto.CreateFillingRequest{}, fmt.Errorf(
			"read filling image: %v: %w",
			err,
			coreerrors.ErrInvalidRequest,
		)
	}
	if len(image) == 0 || len(image) > maxFillingImageSize {
		return dto.CreateFillingRequest{}, fmt.Errorf(
			"filling image must be between 1 byte and 8 MiB: %w",
			coreerrors.ErrInvalidRequest,
		)
	}
	if _, err = png.DecodeConfig(bytes.NewReader(image)); err != nil {
		return dto.CreateFillingRequest{}, fmt.Errorf(
			"filling image is not a valid PNG: %v: %w",
			err,
			coreerrors.ErrInvalidRequest,
		)
	}

	body := dto.CreateFillingRequest{
		Name:        request.FormValue("name"),
		Description: request.FormValue("description"),
		Price:       price,
		Image:       image,
	}
	if err = h.validator.Struct(body); err != nil {
		return dto.CreateFillingRequest{}, fmt.Errorf(
			"validate filling request: %v: %w",
			err,
			coreerrors.ErrInvalidRequest,
		)
	}

	return body, nil
}

// UpdateFilling частично обновляет начинку.
// @Summary Изменение начинки
// @Description Доступно только администратору. Пустой JSON-объект возвращает начинку без изменений.
// @Tags fillings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param fillingId path string true "UUID начинки" Format(uuid)
// @Param request body dto.UpdateFillingRequest true "Изменяемые поля"
// @Success 200 {object} dto.FillingResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 403 {object} core_response.ErrorResponse
// @Failure 404 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/fillings/{fillingId} [patch]
func (h *Handler) UpdateFilling(writer *core_response.Writer, request *http.Request) {
	var body dto.UpdateFillingRequest
	if err := core_request.DecodeAndValidate(request, &body, h.validator); err != nil {
		writer.ErrorResponse(err)
		return
	}

	filling, err := h.service.UpdateFilling(
		request.Context(),
		request.PathValue("fillingId"),
		dto.UpdateFillingRequestToDomain(body),
	)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, dto.FillingResponseFromDomain(filling))
}

// DeleteFilling скрывает начинку.
// @Summary Скрытие начинки
// @Description Доступно только администратору. Начинка помечается неактивной.
// @Tags fillings
// @Produce json
// @Security BearerAuth
// @Param fillingId path string true "UUID начинки" Format(uuid)
// @Success 204
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 403 {object} core_response.ErrorResponse
// @Failure 404 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/fillings/{fillingId} [delete]
func (h *Handler) DeleteFilling(writer *core_response.Writer, request *http.Request) {
	if err := h.service.HideFilling(request.Context(), request.PathValue("fillingId")); err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

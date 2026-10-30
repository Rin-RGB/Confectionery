package fillings

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"SPOproject/internal/core/auth"
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

// GetFillings возвращает начинки с учётом роли пользователя.
// @Summary Получение начинок
// @Description Без токена и с токеном пользователя возвращаются только активные начинки. С действительным Bearer-токеном администратора возвращаются все начинки. Переданный токен проверяется.
// @Tags fillings
// @Produce json
// @Success 200 {array} dto.FillingResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/fillings [get]
func (h *Handler) GetFillings(writer *core_response.Writer, request *http.Request) {
	onlyActive := true
	token := strings.TrimSpace(strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer "))
	if token != "" {
		claims, err := h.provider.ParseToken(token, auth.AccessType)
		if err != nil {
			writer.ErrorResponse(err)
			return
		}
		onlyActive = claims.Role != domain.RoleAdmin
	}

	fillings, err := h.service.GetFillings(request.Context(), onlyActive)
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
// @Description Доступно только администратору. Изображение передаётся файлом image (PNG до 8 МиБ) в multipart/form-data. Для остальных полей также поддерживается JSON. Не переданные поля сохраняются; пустое description очищает описание. image_name напрямую не изменяется.
// @Tags fillings
// @Accept multipart/form-data,json
// @Produce json
// @Security BearerAuth
// @Param fillingId path string true "UUID начинки" Format(uuid)
// @Param name formData string false "Название начинки"
// @Param description formData string false "Описание начинки"
// @Param price formData number false "Цена за килограмм"
// @Param is_active formData boolean false "Активность начинки"
// @Param image formData file false "Новое изображение PNG (до 8 МиБ)"
// @Success 200 {object} dto.FillingResponse
// @Failure 400 {object} core_response.ErrorResponse
// @Failure 401 {object} core_response.ErrorResponse
// @Failure 403 {object} core_response.ErrorResponse
// @Failure 404 {object} core_response.ErrorResponse
// @Failure 500 {object} core_response.ErrorResponse
// @Router /api/v1/fillings/{fillingId} [patch]
func (h *Handler) UpdateFilling(writer *core_response.Writer, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maxMultipartBodySize)
	body, image, err := h.parseUpdateFillingRequest(request)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	filling, err := h.service.UpdateFilling(
		request.Context(),
		request.PathValue("fillingId"),
		dto.UpdateFillingRequestToDomain(body),
		image,
	)
	if err != nil {
		writer.ErrorResponse(err)
		return
	}

	writer.WriteJson(http.StatusOK, dto.FillingResponseFromDomain(filling))
}

func (h *Handler) parseUpdateFillingRequest(request *http.Request) (dto.UpdateFillingRequest, []byte, error) {
	var body dto.UpdateFillingRequest
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil {
		return body, nil, fmt.Errorf("invalid content type: %w", coreerrors.ErrInvalidRequest)
	}
	if mediaType == "application/json" {
		err := core_request.DecodeAndValidate(request, &body, h.validator)
		return body, nil, err
	}
	if mediaType != "multipart/form-data" {
		return body, nil, fmt.Errorf("expected JSON or multipart form: %w", coreerrors.ErrInvalidRequest)
	}
	if err := request.ParseMultipartForm(maxFillingImageSize); err != nil {
		return body, nil, fmt.Errorf("parse filling patch: %v: %w", err, coreerrors.ErrInvalidRequest)
	}
	defer request.MultipartForm.RemoveAll()
	optionalString := func(key string) *string {
		if values, ok := request.MultipartForm.Value[key]; ok && len(values) > 0 {
			return &values[0]
		}
		return nil
	}
	if optionalString("image_name") != nil {
		return body, nil, fmt.Errorf("upload image instead of image_name: %w", coreerrors.ErrInvalidRequest)
	}
	body.Name = optionalString("name")
	body.Description = optionalString("description")
	if value := optionalString("price"); value != nil {
		price, err := strconv.ParseFloat(*value, 64)
		if err != nil {
			return body, nil, fmt.Errorf("invalid price: %w", coreerrors.ErrInvalidRequest)
		}
		body.Price = &price
	}
	if value := optionalString("is_active"); value != nil {
		active, err := strconv.ParseBool(*value)
		if err != nil {
			return body, nil, fmt.Errorf("invalid is_active: %w", coreerrors.ErrInvalidRequest)
		}
		body.IsActive = &active
	}
	var image []byte
	if photos := request.MultipartForm.File["image"]; len(photos) > 0 {
		if len(photos) != 1 {
			return body, nil, fmt.Errorf("expected one image: %w", coreerrors.ErrInvalidRequest)
		}
		file, err := photos[0].Open()
		if err != nil {
			return body, nil, fmt.Errorf("open image: %v: %w", err, coreerrors.ErrInvalidRequest)
		}
		defer file.Close()
		image, err = io.ReadAll(io.LimitReader(file, maxFillingImageSize+1))
		if err != nil || len(image) == 0 || len(image) > maxFillingImageSize {
			return body, nil, fmt.Errorf("image must be between 1 byte and 8 MiB: %w", coreerrors.ErrInvalidRequest)
		}
		if _, err := png.DecodeConfig(bytes.NewReader(image)); err != nil {
			return body, nil, fmt.Errorf("image must be PNG: %w", coreerrors.ErrInvalidRequest)
		}
	}
	if err := h.validator.Struct(body); err != nil {
		return body, nil, fmt.Errorf("validate filling patch: %v: %w", err, coreerrors.ErrInvalidRequest)
	}
	return body, image, nil
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

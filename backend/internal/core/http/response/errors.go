package core_response

import (
	coreerrors "SPOproject/internal/core/errors"
	"net/http"
)

type errorValue struct {
	mapError   error
	statusCode int
	error      string
	logLevel   string
}

type ErrorResponse struct {
	Error Error `json:"error"`
}

type Error struct {
	Code    string `json:"code" example:"invalid request"`
	Message string `json:"message" example:"invalid companyId path parameter"`
}

var errorSlice = []errorValue{
	{mapError: coreerrors.ErrInvalidRequest, statusCode: http.StatusBadRequest, logLevel: "WARN", error: coreerrors.ErrInvalidRequest.Error()},
	{mapError: coreerrors.ErrInvalidCredentials, statusCode: http.StatusUnauthorized, logLevel: "WARN", error: coreerrors.ErrInvalidCredentials.Error()},
	{mapError: coreerrors.ErrNotAuthorized, statusCode: http.StatusUnauthorized, logLevel: "WARN", error: coreerrors.ErrNotAuthorized.Error()},
	{mapError: coreerrors.ErrForbidden, statusCode: http.StatusForbidden, logLevel: "WARN", error: coreerrors.ErrForbidden.Error()},
	{mapError: coreerrors.ErrEmailExists, statusCode: http.StatusConflict, logLevel: "WARN", error: coreerrors.ErrEmailExists.Error()},
	{mapError: coreerrors.ErrNotImplemented, statusCode: http.StatusNotImplemented, logLevel: "WARN", error: coreerrors.ErrNotImplemented.Error()},
	{mapError: coreerrors.ErrInternal, statusCode: http.StatusInternalServerError, logLevel: "ERROR", error: coreerrors.ErrInternal.Error()},
} // ErrExpiredToken

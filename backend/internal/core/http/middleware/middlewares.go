package core_middleware

import (
	"SPOproject/internal/core/auth"
	"SPOproject/internal/core/domain"
	core_logger "SPOproject/internal/core/logger"
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	core_errors "SPOproject/internal/core/errors"
	"SPOproject/internal/core/http/response"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requestID := request.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.NewString()
			}
			request.Header.Set("X-Request-ID", requestID)
			writer.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(writer, request)
		})
	}
}

func Logger(logger *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requestLogger := logger.With(
				zap.String("request_id", request.Header.Get("X-Request-ID")),
				zap.String("method", request.Method),
				zap.String("path", request.URL.Path),
			)
			ctx := core_logger.CtxWithLogger(request.Context(), requestLogger)
			next.ServeHTTP(writer, request.WithContext(ctx))
		})
	}
}

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			wrappedWriter := core_response.WrapWriter(writer, request.Context())
			logger := core_logger.FromContext(request.Context())
			startedAt := time.Now()
			logger.Debug("starting request")

			next.ServeHTTP(wrappedWriter, request)

			logger.Debug(
				"request completed",
				zap.Int("status", wrappedWriter.StatusCode),
				zap.Duration("duration", time.Since(startedAt)),
			)
		})
	}
}

func PanicRecoverer() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				panicValue := recover()
				if panicValue == nil {
					return
				}

				core_logger.FromContext(request.Context()).Error(
					"panic recovered",
					zap.Any("panic", panicValue),
					zap.ByteString("stack", debug.Stack()),
				)
				core_response.WrapWriter(writer, request.Context()).ErrorResponse(core_errors.ErrInternal)
			}()

			next.ServeHTTP(writer, request)
		})
	}
}

func Auth(roleChecker Parser, roles ...domain.UserRole) Middleware {
	allowedRoles := make(map[domain.UserRole]bool, len(roles))
	for i := 0; i < len(roles); i++ {
		allowedRoles[roles[i]] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := core_response.WrapWriter(w, r.Context())

			token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
			if token == "" {
				writer.ErrorResponse(fmt.Errorf("authorization token is required: %w", core_errors.ErrNotAuthorized))
				return
			}
			tokenClaims, err := roleChecker.ParseToken(token, auth.AccessType)
			if err != nil {
				writer.ErrorResponse(err)
				return
			}
			if _, ok := allowedRoles[tokenClaims.Role]; !ok {
				writer.ErrorResponse(core_errors.ErrForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), userIdKey{}, tokenClaims.UserID)
			ctx = context.WithValue(r.Context(), roleKey{}, tokenClaims.Role)
			next.ServeHTTP(writer, r.WithContext(ctx))
		})
	}
}

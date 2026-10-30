package fillings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"SPOproject/internal/core/auth"
	"SPOproject/internal/core/domain"
	corelogger "SPOproject/internal/core/logger"
	"SPOproject/internal/handlers/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type listServiceStub struct {
	fillingService
	called bool
}

func (s *listServiceStub) GetFillings(_ context.Context, onlyActive bool) ([]domain.Filling, error) {
	s.called = true
	items := []domain.Filling{{ID: uuid.NewString(), Name: "Active", IsActive: true}}
	if !onlyActive {
		items = append(items, domain.Filling{ID: uuid.NewString(), Name: "Hidden", IsActive: false})
	}
	return items, nil
}

func TestGetFillingsOptionalToken(t *testing.T) {
	provider := auth.NewJWTProvider("test-secret", time.Hour, time.Hour)
	token := func(p *auth.JWTProvider, role domain.UserRole, kind auth.TokenType) string {
		t.Helper()
		value, err := p.NewToken(domain.User{ID: uuid.New(), Role: role}, kind)
		if err != nil {
			t.Fatal(err)
		}
		return "Bearer " + value
	}
	tests := []struct {
		name   string
		header string
		status int
		count  int
	}{
		{"anonymous", "", http.StatusOK, 1},
		{"user", token(provider, domain.RoleUser, auth.AccessType), http.StatusOK, 1},
		{"admin", token(provider, domain.RoleAdmin, auth.AccessType), http.StatusOK, 2},
		{"invalid", "Bearer invalid", http.StatusBadRequest, 0},
		{"expired", token(auth.NewJWTProvider("test-secret", -time.Hour, time.Hour), domain.RoleAdmin, auth.AccessType), http.StatusUnauthorized, 0},
		{"wrong signature", token(auth.NewJWTProvider("other-secret", time.Hour, time.Hour), domain.RoleAdmin, auth.AccessType), http.StatusBadRequest, 0},
		{"refresh token", token(provider, domain.RoleAdmin, auth.RefreshType), http.StatusBadRequest, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &listServiceStub{}
			handler := NewHandler(service, provider, nil)
			mux := http.NewServeMux()
			for _, route := range handler.Routes() {
				if route.Method == http.MethodGet && route.Path == "/fillings" {
					if len(route.Middlewares) != 0 {
						t.Fatal("list route should be public")
					}
					mux.Handle("GET /fillings", route.Handler)
				}
			}
			request := httptest.NewRequest(http.MethodGet, "/fillings", nil)
			request.Header.Set("Authorization", tt.header)
			request = request.WithContext(corelogger.CtxWithLogger(request.Context(), &corelogger.Logger{Logger: zap.NewNop()}))
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tt.status, response.Body)
			}
			if tt.status != http.StatusOK {
				if service.called {
					t.Fatal("invalid token reached service")
				}
				return
			}
			var result []dto.FillingResponse
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if len(result) != tt.count {
				t.Fatalf("got %d fillings, want %d", len(result), tt.count)
			}
		})
	}
}

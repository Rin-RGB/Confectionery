package core_server_test

import (
	core_middleware "SPOproject/internal/core/http/middleware"
	core_response "SPOproject/internal/core/http/response"
	core_server "SPOproject/internal/core/http/server"
	"net/http"
	"net/http/httptest"
	"testing"

	"SPOproject/internal/core/config"
	corelogger "SPOproject/internal/core/logger"
)

func TestServerRoutesVersionedAPI(t *testing.T) {
	router := core_server.NewAPIVersionRouter(core_server.Version1)
	router.ChainRoutes(core_server.Route{
		Method: http.MethodGet,
		Path:   "/health",
		Handler: func(writer *core_response.Writer, _ *http.Request) {
			writer.WriteJson(http.StatusOK, map[string]string{"status": "ok"})
		},
	})

	logger, err := corelogger.NewLogger(config.LoggerConfig{Level: "ERROR", Folder: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()
	httpServer := core_server.New(
		config.ServerConfig{},
		[]*core_server.Router{router},
		core_middleware.RequestID(),
		core_middleware.Logger(logger),
		core_middleware.Trace(),
		core_middleware.PanicRecoverer(),
	)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	recorder := httptest.NewRecorder()
	httpServer.Handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected response request ID")
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("unexpected content type %q", contentType)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	logger, err := corelogger.NewLogger(config.LoggerConfig{Level: "ERROR", Folder: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Close() }()
	handler := core_middleware.Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("test panic")
		}),
		core_middleware.RequestID(),
		core_middleware.Logger(logger),
		core_middleware.PanicRecoverer(),
	)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

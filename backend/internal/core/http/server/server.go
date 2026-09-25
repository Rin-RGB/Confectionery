package core_server

import (
	"net/http"

	"SPOproject/internal/core/config"
	corehandler "SPOproject/internal/core/http/handler"
	core_middleware "SPOproject/internal/core/http/middleware"
	corestatic "SPOproject/internal/static"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Route struct {
	Method      string
	Path        string
	Handler     corehandler.HandlerFunc
	Middlewares []core_middleware.Middleware
}

func New(config config.ServerConfig, routers []*Router, middlewares ...core_middleware.Middleware) *http.Server {
	mux := registerRouters(routers...)

	return &http.Server{
		Addr:              config.Addr,
		Handler:           core_middleware.Chain(mux, middlewares...),
		ReadHeaderTimeout: config.ReadHeaderTimeout,
	}
}

func registerRouters(routers ...*Router) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", corestatic.Handler()))
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	for _, router := range routers {
		prefix := "/api/" + string(router.version)
		handler := core_middleware.Chain(router.mux, router.middlewares...)
		mux.Handle(prefix+"/", http.StripPrefix(prefix, handler))
	}

	return mux
}

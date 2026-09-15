package core_server

import (
	core_middleware "SPOproject/internal/core/http/middleware"
	"net/http"

	"SPOproject/internal/core/config"
	corehandler "SPOproject/internal/core/http/handler"
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

	for _, router := range routers {
		prefix := "/api/" + string(router.version)
		handler := core_middleware.Chain(router.mux, router.middlewares...)
		mux.Handle(prefix+"/", http.StripPrefix(prefix, handler))
	}

	return mux
}

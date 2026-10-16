package core_server

import (
	core_middleware "SPOproject/internal/core/http/middleware"
	"fmt"
	"net/http"
)

type APIVersion string

const Version1 APIVersion = "v1"

type Router struct {
	mux         *http.ServeMux
	version     APIVersion
	middlewares []core_middleware.Middleware
}

func NewAPIVersionRouter(version APIVersion, middlewares ...core_middleware.Middleware) *Router {
	return &Router{
		mux:         http.NewServeMux(),
		version:     version,
		middlewares: middlewares,
	}
}

func (router *Router) ChainRoutes(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)
		handler := core_middleware.Chain(route.Handler, route.Middlewares...)
		router.mux.Handle(pattern, handler)
	}
}

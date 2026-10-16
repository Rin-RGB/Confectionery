package core_handler

import (
	"net/http"

	"SPOproject/internal/core/http/response"
)

type HandlerFunc func(writer *core_response.Writer, request *http.Request)

func (handler HandlerFunc) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	handler(core_response.WrapWriter(writer, request.Context()), request)
}

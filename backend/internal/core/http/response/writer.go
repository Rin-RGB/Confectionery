package core_response

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	corelogger "SPOproject/internal/core/logger"

	"go.uber.org/zap"
)

type Writer struct {
	http.ResponseWriter
	StatusCode int
	Logger     *corelogger.Logger
}

func WrapWriter(writer http.ResponseWriter, ctx context.Context) *Writer {
	if wrappedWriter, ok := writer.(*Writer); ok {
		return wrappedWriter
	}

	return &Writer{
		ResponseWriter: writer,
		StatusCode:     http.StatusOK,
		Logger:         corelogger.FromContext(ctx),
	}
}

func (writer *Writer) WriteHeader(statusCode int) {
	writer.StatusCode = statusCode
	writer.ResponseWriter.WriteHeader(statusCode)
}

func (writer *Writer) ErrorResponse(err error) {
	var value errorValue
	found := false
	for i := 0; i < len(errorSlice); i++ {
		if errors.Is(err, errorSlice[i].mapError) {
			value = errorSlice[i]
			found = true
			break
		}
	}

	if !found {
		writer.Logger.Error("got INTERNAL error", zap.Error(err))
		writer.writeErrorJson(http.StatusInternalServerError, "", err.Error())
		return
	}

	switch value.logLevel {
	case "DEBUG":
		writer.Logger.Debug("got error", zap.Error(err))
	case "WARN":
		writer.Logger.Warn("got error", zap.Error(err))
	case "ERROR":
		writer.Logger.Error("got INTERNAL error", zap.Error(err))
	}

	writer.writeErrorJson(value.statusCode, value.error, err.Error())
}

func (writer *Writer) writeErrorJson(statusCode int, errorCode, message string) {
	response := ErrorResponse{
		Error: Error{
			Code:    errorCode,
			Message: message,
		},
	}
	writer.WriteJson(statusCode, response)
}

func (writer *Writer) WriteJson(statusCode int, body any) {
	bytes, err := json.MarshalIndent(body, "", "    ")
	if err != nil {
		writer.Logger.Error("failed to marshal data into json", zap.Error(err))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	if _, err = writer.Write(bytes); err != nil {
		writer.Logger.Error("failed to write data", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

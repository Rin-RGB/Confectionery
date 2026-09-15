package core_request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	coreerrors "SPOproject/internal/core/errors"
)

type Validator interface {
	Struct(s any) error
}

func DecodeAndValidate(r *http.Request, dest any, validator Validator) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return fmt.Errorf("failed to decode request: %s: %w", err.Error(), coreerrors.ErrInvalidRequest)
	}

	if err := validator.Struct(dest); err != nil {
		return fmt.Errorf("failed to validate request: %s: %w", err.Error(), coreerrors.ErrInvalidRequest)
	}

	return nil
}

func DecodeJSON(request *http.Request, destination any) error {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode request body: %v: %w", err, coreerrors.ErrInvalidRequest)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("request body must contain one JSON object: %w", coreerrors.ErrInvalidRequest)
	}
	return nil
}

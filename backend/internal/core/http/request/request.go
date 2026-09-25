package core_request

import (
	"encoding/json"
	"fmt"
	"net/http"

	coreerrors "SPOproject/internal/core/errors"
)

type Validator interface {
	Struct(s any) error
}

func DecodeAndValidate(r *http.Request, dest any, validator Validator) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("failed to decode request: %s: %w", err.Error(), coreerrors.ErrInvalidRequest)
	}

	if err := validator.Struct(dest); err != nil {
		return fmt.Errorf("failed to validate request: %s: %w", err.Error(), coreerrors.ErrInvalidRequest)
	}

	return nil
}

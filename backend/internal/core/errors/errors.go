package core_errors

import "errors"

var (
	ErrNotImplemented     = errors.New("not implemented")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrEmailExists        = errors.New("email is already in use")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrNotAuthorized      = errors.New("not authorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInternal           = errors.New("internal server error")
	ErrExpiredToken       = errors.New("token is expired")
)

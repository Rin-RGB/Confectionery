package core_middleware

import "SPOproject/internal/core/auth"

type Parser interface {
	ParseToken(tokenString string, tokenType auth.TokenType) (auth.Claims, error)
}

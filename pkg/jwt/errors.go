package jwt

import "errors"

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrUnknownClaimsType = errors.New("unknown claims type")
	ErrTokensNotPaired   = errors.New("tokens is not paired")
)

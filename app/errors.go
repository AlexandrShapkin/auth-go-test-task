package app

import "errors"

var (
	ErrInvalidRequestData = errors.New("invalid request data")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound = errors.New("user not found")
	ErrRefreshTokenRequired = errors.New("refresh token is required")
	ErrAccessTokenRequired = errors.New("access token is required")
	ErrIncorrectRefreshToken = errors.New("incorrect refresh token")
)
package client

import "errors"

var (
	ErrServerFailure = errors.New("server error")
	ErrInvalidCookie = errors.New("invalid cookie")
)
package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenNotFound      = errors.New("refresh token not found")
	ErrTokenExpired       = errors.New("refresh token expired")
)

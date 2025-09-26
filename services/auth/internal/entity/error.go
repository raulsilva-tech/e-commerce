package entity

import "errors"

var (
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrEmailPasswordRequired     = errors.New("email and password required")
	ErrNameEmailPasswordRequired = errors.New("name,email and password required")
	ErrEmailAlreadyUsed          = errors.New("email already used")
)

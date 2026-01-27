package domain

import "errors"

var (
	ErrInvalidID = errors.New("invalid id")
	ErrNotFound  = errors.New("entity not found")
)

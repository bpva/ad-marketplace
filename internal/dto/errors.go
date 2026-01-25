package dto

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrForbidden         = errors.New("forbidden")
	ErrUserNotRegistered = errors.New("user not registered")
	ErrCannotRemoveOwner = errors.New("cannot remove owner")
)

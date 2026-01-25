package dto

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrForbidden         = errors.New("forbidden")
	ErrCannotRemoveOwner = errors.New("cannot remove owner")
)

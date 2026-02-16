package dto

import "net/http"

type APIError struct {
	status  int
	code    string
	details map[string]any
	cause   error
}

type ErrorResponse struct {
	ErrorCode string         `json:"error_code"`
	Details   map[string]any `json:"details,omitempty"`
}

func new(status int, code string) *APIError {
	return &APIError{status: status, code: code}
}

var (
	// 400 Bad Request
	ErrBadRequest           = new(http.StatusBadRequest, "bad_request")
	ErrValidation           = new(http.StatusBadRequest, "invalid_request")
	ErrInvalidChannelID     = new(http.StatusBadRequest, "invalid_channel_id")
	ErrInvalidTelegramID    = new(http.StatusBadRequest, "invalid_telegram_id")
	ErrTelegramIDRequired   = new(http.StatusBadRequest, "telegram_id_required")
	ErrCannotRemoveOwner    = new(http.StatusBadRequest, "cannot_remove_owner")
	ErrFormatTypeNotAllowed = new(http.StatusBadRequest, "format_type_not_allowed")
	ErrInvalidFormatID      = new(http.StatusBadRequest, "invalid_format_id")
	ErrTooManyCategories    = new(http.StatusBadRequest, "too_many_categories")
	ErrInvalidCategory      = new(http.StatusBadRequest, "invalid_category")

	// 401 Unauthorized
	ErrUnauthorized = new(http.StatusUnauthorized, "unauthorized")

	// 403 Forbidden
	ErrForbidden = new(http.StatusForbidden, "forbidden")

	// 404 Not Found
	ErrNotFound = new(http.StatusNotFound, "not_found")

	// 422 Unprocessable Entity
	ErrNoPayoutMethod = new(http.StatusUnprocessableEntity, "no_payout_method")

	// 409 Conflict
	ErrAdFormatExists = new(http.StatusConflict, "ad_format_exists")

	// 500 Internal Server Error
	ErrInternalError = new(http.StatusInternalServerError, "internal_error")
)

func (e *APIError) Error() string {
	if e.cause != nil {
		return e.code + ": " + e.cause.Error()
	}
	return e.code
}

func (e *APIError) Unwrap() error {
	return e.cause
}

func (e *APIError) Status() int             { return e.status }
func (e *APIError) Code() string            { return e.code }
func (e *APIError) Details() map[string]any { return e.details }

func (e *APIError) Wrap(cause error) *APIError {
	return &APIError{
		status:  e.status,
		code:    e.code,
		details: e.details,
		cause:   cause,
	}
}

func (e *APIError) WithDetails(details map[string]any) *APIError {
	return &APIError{
		status:  e.status,
		code:    e.code,
		details: details,
		cause:   e.cause,
	}
}

package respond

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/bpva/ad-marketplace/internal/dto"
)

func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, data)
}

func Created(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, data)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Err(w http.ResponseWriter, log *slog.Logger, err error) {
	var apiErr *dto.APIError
	isAPIError := errors.As(err, &apiErr)

	if !isAPIError {
		log.Error("internal error", "error", err)
		apiErr = dto.ErrInternalError
	}

	if isAPIError && apiErr.Status() >= 500 {
		log.Error("server error", "error", err)
	}

	if isAPIError && apiErr.Status() < 500 {
		log.Warn("client error", "error", err)
	}

	resp := dto.ErrorResponse{
		ErrorCode: apiErr.Code(),
		Details:   apiErr.Details(),
	}
	writeJSON(w, apiErr.Status(), resp)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

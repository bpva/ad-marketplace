package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	nethttpmw "github.com/oapi-codegen/nethttp-middleware"

	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/respond"
)

func NewOpenAPIValidator(log *slog.Logger) (func(http.Handler) http.Handler, error) {
	spec, err := openapi3.NewLoader().LoadFromFile("docs/openapi.json")
	if err != nil {
		return nil, fmt.Errorf("load OpenAPI spec: %w", err)
	}
	spec.Servers = nil

	return nethttpmw.OapiRequestValidatorWithOptions(spec, &nethttpmw.Options{
		Options: openapi3filter.Options{
			MultiError: true,
		},
		ErrorHandlerWithOpts: func(
			_ context.Context,
			err error,
			w http.ResponseWriter,
			_ *http.Request,
			_ nethttpmw.ErrorHandlerOpts,
		) {
			respond.Err(w, log, dto.ErrValidation.Wrap(err))
		},
	}), nil
}

package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	nethttpmw "github.com/oapi-codegen/nethttp-middleware"

	"github.com/bpva/ad-marketplace/docs"
	"github.com/bpva/ad-marketplace/internal/dto"
	"github.com/bpva/ad-marketplace/internal/http/respond"
)

func NewOpenAPIValidator(log *slog.Logger) (func(http.Handler) http.Handler, error) {
	spec, err := openapi3.NewLoader().LoadFromData(docs.OpenAPISpec)
	if err != nil {
		return nil, fmt.Errorf("parse OpenAPI spec: %w", err)
	}

	return nethttpmw.OapiRequestValidatorWithOptions(spec, &nethttpmw.Options{
		SilenceServersWarning: true,
		Options: openapi3filter.Options{
			MultiError:         true,
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
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

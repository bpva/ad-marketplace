package bind

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/bpva/ad-marketplace/internal/dto"
)

type Validatable interface {
	Valid() error
}

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}

func JSON(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return dto.ErrBadRequest
	}
	if err := validate.Struct(v); err != nil {
		return validationError(err)
	}
	return validateCustom(reflect.ValueOf(v))
}

func validationError(err error) error {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return dto.ErrBadRequest
	}
	fields := make(map[string]string, len(ve))
	for _, fe := range ve {
		fields[fe.Field()] = fe.Tag()
	}
	return dto.ErrValidation.WithDetails(map[string]any{"fields": fields})
}

func validateCustom(v reflect.Value) error {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if val, ok := v.Interface().(Validatable); ok {
		if err := val.Valid(); err != nil {
			return dto.ErrValidation.WithDetails(map[string]any{"reason": err.Error()})
		}
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue
		}
		if err := validateCustom(field); err != nil {
			return err
		}
	}
	return nil
}

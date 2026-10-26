package server

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate checks request bodies. It is safe for concurrent use.
var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	// Report fields by their JSON name, not the Go name.
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}

// validationMessage joins all field errors into one message, in struct field order.
func validationMessage(err error) string {
	var errs validator.ValidationErrors
	if !errors.As(err, &errs) {
		return err.Error()
	}
	msgs := make([]string, len(errs))
	for i, fe := range errs {
		msgs[i] = fieldMessage(fe)
	}
	return strings.Join(msgs, "; ")
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "oneof":
		return fe.Field() + " must be one of: " + strings.Join(strings.Fields(fe.Param()), ", ")
	case "required_if", "excluded_if":
		// Param is "<GoField> <value>", for example "Scope project".
		other, value, _ := strings.Cut(fe.Param(), " ")
		if fe.Tag() == "required_if" {
			return fmt.Sprintf("%s is required for %s %s", fe.Field(), strings.ToLower(other), value)
		}
		return fmt.Sprintf("%s is not allowed for %s %s", fe.Field(), strings.ToLower(other), value)
	default:
		return fmt.Sprintf("%s is not valid (%s)", fe.Field(), fe.Tag())
	}
}

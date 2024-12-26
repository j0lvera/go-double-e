package server

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string
	Message string
}

func ParseValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		for _, valErr := range valErrs {
			validationErrors = append(validationErrors, ValidationError{
				Field:   valErr.Field(),
				Message: getErrorMessage(valErr),
			})
		}
	}

	return validationErrors
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email address"
	case "min":
		return fmt.Sprintf("This field must be at least %s characters long", err.Param())
	case "max":
		return fmt.Sprintf("This field must be at most %s characters long", err.Param())
	default:
		return "Invalid value"
	}
}

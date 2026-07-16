package utils

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func InitValidator() {
	Validate = validator.New()
}

// ValidationError represents a single field validation error
type ValidationError struct {
	FailedField string
	Tag         string
	Value       interface{}
}

/*
Package utils provides validation utilities for struct-based data.

This file wraps the go-playground/validator package, offering:
- A centralized validator instance
- Conversion of validation errors into a consistent application-level format
- Support for custom field/tag/value mapping
*/

// Validator wraps the go-playground validator to add custom behavior
type Validator struct {
	validator *validator.Validate
}

// NewValidator returns a new Validator instance with struct-level validation enabled
func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(validator.WithRequiredStructEnabled()),
	}
}

// Validator exposes the raw *validator.Validate instance
func (v *Validator) Validator() *validator.Validate {
	return v.validator
}

// Validate runs struct validation and returns a slice of custom ValidationError objects
func (v Validator) Validate(data interface{}) []ValidationError {
	validationErrors := []ValidationError{}

	errs := v.validator.Struct(data)
	if errs != nil {
		// Handle invalid validator usage (e.g., nil input)
		if _, ok := errs.(*validator.InvalidValidationError); ok {
			return []ValidationError{{
				FailedField: "validation",
				Tag:         "invalid",
				Value:       "validation error",
			}}
		}
		// Convert each validation error to app-specific format
		for _, err := range errs.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, NewError(err))
		}
	}

	return validationErrors
}

// NewError converts a validator.FieldError into a models.ValidationError
func NewError(err validator.FieldError) ValidationError {
	return ValidationError{
		FailedField: err.Field(),
		Tag:         err.Tag(),
		Value:       err.Value(),
	}
}

// AggregateValidationErrors aggregates multiple ValidationError objects into a single error
func AggregateValidationErrors(errs []ValidationError) error {
	if len(errs) == 0 {
		return nil
	}

	errorsSlice := make([]error, 0, len(errs))
	for _, e := range errs {
		errorsSlice = append(errorsSlice, errors.New(e.FailedField+":"+e.Tag))
	}

	return errors.Join(errorsSlice...)
}

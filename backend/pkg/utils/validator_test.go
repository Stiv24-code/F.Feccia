package utils

import (
	"strings"
	"testing"
)

// Simple struct for validation tests.
type testUser struct {
	Name string `validate:"required"`
	Age  int    `validate:"gte=18"`
}

func TestInitValidatorSetsGlobal(t *testing.T) {
	// Do not parallelize: modifies global Validate.
	InitValidator()
	if Validate == nil {
		t.Fatalf("expected global Validate to be initialized, got nil")
	}
}

func TestNewValidator_ValidateValidStruct(t *testing.T) {
	t.Parallel()
	v := NewValidator()

	u := testUser{
		Name: "John",
		Age:  25,
	}

	errs := v.Validate(u)
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %d: %+v", len(errs), errs)
	}
}

func TestNewValidator_ValidateSingleError(t *testing.T) {
	t.Parallel()
	v := NewValidator()

	// Missing Name, Age is valid.
	u := testUser{
		Name: "",
		Age:  25,
	}

	errs := v.Validate(u)
	if len(errs) != 1 {
		t.Fatalf("expected 1 validation error, got %d: %+v", len(errs), errs)
	}

	e := errs[0]
	if e.FailedField != "Name" {
		t.Fatalf("expected FailedField %q, got %q", "Name", e.FailedField)
	}
	if e.Tag != "required" {
		t.Fatalf("expected Tag %q, got %q", "required", e.Tag)
	}
	if e.Value != "" {
		t.Fatalf("expected Value to be empty string, got %#v", e.Value)
	}
}

func TestNewValidator_ValidateMultipleErrors(t *testing.T) {
	t.Parallel()
	v := NewValidator()

	// Both fields invalid: Name required, Age must be >= 18.
	u := testUser{
		Name: "",
		Age:  10,
	}

	errs := v.Validate(u)
	if len(errs) != 2 {
		t.Fatalf("expected 2 validation errors, got %d: %+v", len(errs), errs)
	}

	// Do not rely on order; build a map for checks.
	fieldTags := map[string]string{}
	for _, e := range errs {
		fieldTags[e.FailedField] = e.Tag
	}

	if tag, ok := fieldTags["Name"]; !ok || tag != "required" {
		t.Fatalf("expected Name to have tag %q, got %q (present=%v)", "required", tag, ok)
	}
	if tag, ok := fieldTags["Age"]; !ok || tag != "gte" {
		t.Fatalf("expected Age to have tag %q, got %q (present=%v)", "gte", tag, ok)
	}
}

func TestValidator_InvalidValidationErrorNilInput(t *testing.T) {
	t.Parallel()
	v := NewValidator()

	errs := v.Validate(nil)
	if len(errs) != 1 {
		t.Fatalf("expected 1 validation error for invalid input, got %d: %+v", len(errs), errs)
	}

	e := errs[0]
	if e.FailedField != "validation" {
		t.Fatalf("expected FailedField %q, got %q", "validation", e.FailedField)
	}
	if e.Tag != "invalid" {
		t.Fatalf("expected Tag %q, got %q", "invalid", e.Tag)
	}
	if e.Value != "validation error" {
		t.Fatalf("expected Value %q, got %#v", "validation error", e.Value)
	}
}

func TestAggregateValidationErrors_EmptyAndNil(t *testing.T) {
	t.Parallel()
	if err := AggregateValidationErrors(nil); err != nil {
		t.Fatalf("expected nil error for nil slice, got %v", err)
	}

	if err := AggregateValidationErrors([]ValidationError{}); err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestAggregateValidationErrors_Multiple(t *testing.T) {
	t.Parallel()
	errs := []ValidationError{
		{FailedField: "Name", Tag: "required"},
		{FailedField: "Age", Tag: "gte"},
	}

	err := AggregateValidationErrors(errs)
	if err == nil {
		t.Fatalf("expected non-nil aggregated error")
	}

	msg := err.Error()
	if !strings.Contains(msg, "Name:required") {
		t.Fatalf("expected aggregated error message to contain %q, got %q", "Name:required", msg)
	}
	if !strings.Contains(msg, "Age:gte") {
		t.Fatalf("expected aggregated error message to contain %q, got %q", "Age:gte", msg)
	}
}

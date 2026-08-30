package productconfig

import (
	"errors"
	"regexp"
	"strings"
)

type DiagnosticKind string

const (
	DiagnosticReadFailed   DiagnosticKind = "read_failed"
	DiagnosticYAMLInvalid  DiagnosticKind = "yaml_invalid"
	DiagnosticUnknownField DiagnosticKind = "unknown_field"
	DiagnosticMissingField DiagnosticKind = "missing_field"
	DiagnosticInvalidValue DiagnosticKind = "invalid_value"
)

type Diagnostic struct {
	Kind  DiagnosticKind
	Field string
}

type diagnosticError struct {
	kind  DiagnosticKind
	field string
	cause error
}

func (err *diagnosticError) Error() string { return err.cause.Error() }
func (err *diagnosticError) Unwrap() error { return err.cause }

type validationError struct {
	field   string
	message string
	cause   error
}

func (err *validationError) Error() string {
	if err.message == "" {
		return "configuration field " + err.field + ": " + err.cause.Error()
	}
	return "configuration field " + err.field + " " + err.message + ": " + err.cause.Error()
}

func (err *validationError) Unwrap() error { return err.cause }

var diagnosticField = regexp.MustCompile(`^[a-z_][a-z0-9_]*(\[[0-9]+\])?(\.[a-z_][a-z0-9_]*(\[[0-9]+\])?)*$`)

func Diagnose(err error) Diagnostic {
	if err == nil {
		return Diagnostic{}
	}
	var detailed *diagnosticError
	if errors.As(err, &detailed) {
		return Diagnostic{Kind: detailed.kind, Field: safeDiagnosticField(detailed.field)}
	}
	var validation *validationError
	if errors.As(err, &validation) {
		return Diagnostic{Kind: DiagnosticInvalidValue, Field: safeDiagnosticField(validation.field)}
	}
	if errors.Is(err, ErrNotFound) {
		return Diagnostic{Kind: DiagnosticReadFailed}
	}
	return Diagnostic{Kind: DiagnosticInvalidValue}
}

func withDiagnostic(err error, kind DiagnosticKind, field string) error {
	return &diagnosticError{kind: kind, field: safeDiagnosticField(field), cause: err}
}

func validationField(err error) string {
	var validation *validationError
	if !errors.As(err, &validation) {
		return ""
	}
	return validation.field
}

func safeDiagnosticField(field string) string {
	field = strings.TrimSpace(field)
	if len(field) > 256 || !diagnosticField.MatchString(field) {
		return ""
	}
	return field
}

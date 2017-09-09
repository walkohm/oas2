package oas2

import (
	"fmt"
	"net/url"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// ValidateQuery validates request query parameters by spec and returns errors
// if any.
func ValidateQuery(ps []spec.Parameter, q url.Values) []error {
	errs := make(ValidationErrors, 0)

	// Iterate over spec parameters and validate each against the spec.
	for _, p := range ps {
		if p.In != "query" {
			// Validating only "query" parameters.
			continue
		}

		errs = append(errs, validateQueryParam(p, q)...)

		delete(q, p.Name) // to check not described parameters passed
	}

	// Check that no additional parameters passed.
	for name := range q {
		errs = append(errs, ValidationErrorf(name, q.Get(name), "parameter %s is unknown", name))
	}

	return errs.Errors()
}

func validateQueryParam(p spec.Parameter, q url.Values) (errs ValidationErrors) {
	_, ok := q[p.Name]
	if !ok {
		if p.Required {
			errs = append(errs, ValidationErrorf(p.Name, nil, "param %s is required", p.Name))
		}
		return errs
	}

	value, err := ConvertParameter(q[p.Name], p.Type, p.Format)
	if err != nil {
		// TODO: q.Get(p.Name) relies on type that is not array/file.
		return append(errs, ValidationErrorf(p.Name, q.Get(p.Name), "param %s: %s", p.Name, err))
	}

	if result := validate.NewParamValidator(&p, strfmt.Default).Validate(value); result != nil {
		for _, e := range result.Errors {
			errs = append(errs, ValidationErrorf(p.Name, value, e.Error()))
		}
	}

	return errs
}

// ValidationError describes validation error.
type ValidationError interface {
	error

	// Field returns field name where error occurred.
	Field() string

	// Value returns original value passed by client on field where error
	// occurred.
	Value() interface{}
}

// ValidationErrorf returns a new formatted ValidationError.
func ValidationErrorf(field string, value interface{}, format string, args ...interface{}) ValidationError {
	return valErr{
		message: fmt.Sprintf(format, args...),
		field:   field,
		value:   value,
	}
}

// ValidationErrors is a set of validation errors.
type ValidationErrors []ValidationError

// Errors returns ValidationErrors in form of Go builtin errors.
func (es ValidationErrors) Errors() []error {
	if len(es) == 0 {
		return nil
	}

	errs := make([]error, len(es))
	for i, e := range es {
		errs[i] = e
	}
	return errs
}

// valErr implements ValidationError.
type valErr struct {
	message string
	field   string
	value   interface{}
}

func (v valErr) Error() string {
	return v.message
}

func (v valErr) Field() string {
	return v.field
}

func (v valErr) Value() interface{} {
	return v.value
}

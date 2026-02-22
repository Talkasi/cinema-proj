package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field  string `json:"field"`
	Rule   string `json:"rule"`
	Param  string `json:"param,omitempty"`
	Reason string `json:"reason"`
}

type Error struct {
	Msg    string       `json:"message"`
	Fields []FieldError `json:"fields,omitempty"`
	Err    error        `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}

	return e.Msg
}

func (e *Error) Unwrap() error { return e.Err }

var v = validator.New(validator.WithRequiredStructEnabled())

func ValidateDTO(dto any) error {
	if err := v.Struct(dto); err != nil {
		ve, ok := err.(validator.ValidationErrors)
		if !ok {
			return &Error{Msg: "Invalid input", Err: err}
		}

		fields := make([]FieldError, 0, len(ve))
		for _, fe := range ve {
			fields = append(fields, FieldError{
				Field:  toLowerFirst(fe.Field()),
				Rule:   fe.Tag(),
				Param:  fe.Param(),
				Reason: buildReason(fe),
			})
		}

		return &Error{
			Msg:    "Invalid input",
			Fields: fields,
			Err:    err,
		}
	}

	return nil
}

func toLowerFirst(s string) string {
	if s == "" {
		return s
	}

	return strings.ToLower(s[:1]) + s[1:]
}

func buildReason(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "invalid email format"
	case "min":
		return "too small"
	case "max":
		return "too large"
	default:
		return "validation failed"
	}
}

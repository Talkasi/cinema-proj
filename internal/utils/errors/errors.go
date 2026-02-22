package errors

import (
	"fmt"
	"net/http"
)

type Error struct {
	Message string
	Code    int
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

func NewBadRequest(msg string, err error) *Error {
	return &Error{Message: msg, Code: http.StatusBadRequest, Err: err}
}

func NewNotFound(msg string, err error) *Error {
	return &Error{Message: msg, Code: http.StatusNotFound, Err: err}
}

func NewConflict(msg string, err error) *Error {
	return &Error{Message: msg, Code: http.StatusConflict, Err: err}
}

func NewForbidden(msg string, err error) *Error {
	return &Error{Message: msg, Code: http.StatusForbidden, Err: err}
}

func NewInternal(msg string, err error) *Error {
	return &Error{Message: msg, Code: http.StatusInternalServerError, Err: err}
}

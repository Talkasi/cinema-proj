package utils

import (
	uerrors "cw/internal/utils/errors"
	"cw/internal/utils/httperr"
	"cw/internal/utils/pgerr"
	"net/http"
)

type Error = uerrors.Error

func NewBadRequest(msg string, err error) *Error { return uerrors.NewBadRequest(msg, err) }
func NewNotFound(msg string, err error) *Error   { return uerrors.NewNotFound(msg, err) }
func NewConflict(msg string, err error) *Error   { return uerrors.NewConflict(msg, err) }
func NewForbidden(msg string, err error) *Error  { return uerrors.NewForbidden(msg, err) }
func NewInternal(msg string, err error) *Error   { return uerrors.NewInternal(msg, err) }

func ConvertError(err error) *Error {
	return pgerr.ConvertError(err)
}

func IsDuplicateKeyError(err error) bool {
	return pgerr.IsDuplicateKeyError(err)
}

func WriteError(w http.ResponseWriter, err *Error) {
	httperr.WriteError(w, err)
}

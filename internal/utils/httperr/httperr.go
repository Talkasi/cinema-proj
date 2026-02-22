package httperr

import (
	dto "cw/internal/dto/models"
	uerrors "cw/internal/utils/errors"
	validation "cw/internal/utils/validation"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func WriteError(w http.ResponseWriter, err *uerrors.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)

	message := err.Message
	if err.Err != nil {
		var validationErr *validation.Error
		if !errors.As(err.Err, &validationErr) {
			message += " " + err.Err.Error()
		}
	}
	if encodeErr := json.NewEncoder(w).Encode(dto.ErrorResponse{Message: message}); encodeErr != nil {
		fmt.Printf("Failed to encode error response: %v\n", encodeErr)
	}
}

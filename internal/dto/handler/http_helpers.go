package handler

import (
	"bytes"
	"cw/internal/utils"
	"cw/internal/utils/validation"
	"encoding/json"
	"net/http"
	"strconv"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxPageLimit = 100
)

func parsePaginationParams(r *http.Request) (int, int) {
	page := parsePositiveIntOrDefault(r.URL.Query().Get("page"), defaultPage)
	limit := parsePositiveIntOrDefault(r.URL.Query().Get("limit"), defaultLimit)
	if limit > maxPageLimit {
		limit = defaultLimit
	}

	return page, limit
}

func parsePositiveIntOrDefault(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}

	return value
}

func decodeJSONBody(r *http.Request, dst any) *utils.Error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return utils.NewBadRequest("Invalid input", err)
	}

	return nil
}

func decodeAndValidateJSONBody(r *http.Request, dst any) *utils.Error {
	if err := decodeJSONBody(r, dst); err != nil {
		return err
	}

	if err := validation.ValidateDTO(dst); err != nil {
		return utils.NewBadRequest("Invalid input", err)
	}

	return nil
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) *utils.Error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return utils.NewInternal("failed to encode JSON", err)
	}

	w.Header().Set("Content-Type", "application/json")
	if statusCode > 0 {
		w.WriteHeader(statusCode)
	}

	if _, err := w.Write(buf.Bytes()); err != nil {
		return utils.NewInternal("failed to write response body", err)
	}

	return nil
}

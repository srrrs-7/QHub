package requtil

import (
	"api/src/domain/apperror"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
)

var Validate = validator.New()
var Sanitize = bluemonday.StrictPolicy()

// Decode decodes a JSON request body into T and validates it.
func Decode[T any](r *http.Request, sanitize func(*T)) (T, error) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, apperror.NewBadRequestError(err, "request")
	}
	if sanitize != nil {
		sanitize(&req)
	}
	if err := Validate.Struct(req); err != nil {
		return req, apperror.NewValidationError(err, "request")
	}
	return req, nil
}

// ParseUUID extracts a chi URL parameter and parses it as UUID.
func ParseUUID(r *http.Request, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		return uuid.UUID{}, apperror.NewValidationError(err, param)
	}
	return id, nil
}

// ValidateParams validates a struct built from URL parameters.
func ValidateParams[T any](v T) (T, error) {
	if err := Validate.Struct(v); err != nil {
		return v, apperror.NewValidationError(err, "params")
	}
	return v, nil
}

// MergeField returns the new value if non-empty, otherwise the existing value.
func MergeField[T any](existing T, raw string, constructor func(string) (T, error)) (T, error) {
	if raw == "" {
		return existing, nil
	}
	return constructor(raw)
}

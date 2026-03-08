package requtil

import (
	"api/src/domain/apperror"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

var Validate = validator.New()
var Sanitize = bluemonday.StrictPolicy()

// Decode decodes a JSON request body into T and validates it.
// The sanitize function is called before validation to clean inputs.
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

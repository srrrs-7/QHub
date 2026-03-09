package users

import (
	"api/src/domain/apperror"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

// Global validator instance for performance.
var validate = validator.New()

// Global sanitizer instance.
var sanitize = bluemonday.StrictPolicy()

type getRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

func newGetRequest(r *http.Request) getRequest {
	return getRequest{
		ID: chi.URLParam(r, "id"),
	}
}

func (req getRequest) validate() (getRequest, error) {
	if err := validate.Struct(req); err != nil {
		return getRequest{}, apperror.NewValidationError(err, "GetRequest")
	}
	return req, nil
}

type postRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
	Name  string `json:"name" validate:"required,min=1,max=100"`
}

func newPostRequest(r *http.Request) (postRequest, error) {
	var req postRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return postRequest{}, apperror.NewBadRequestError(err, "postRequest")
	}
	return req, nil
}

func (req postRequest) sanitizeAndValidate() (postRequest, error) {
	req.Email = sanitize.Sanitize(req.Email)
	req.Name = sanitize.Sanitize(req.Name)

	if err := validate.Struct(req); err != nil {
		return postRequest{}, apperror.NewValidationError(err, "postRequest")
	}
	return req, nil
}

package tasks

import (
	"api/src/domain/apperror"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

// Global validator instance for performance
var validate = validator.New()

// Global sanitizer instance
var sanitize = bluemonday.StrictPolicy()

type getRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

func newGetRequest(r *http.Request) getRequest {
	return getRequest{
		ID: chi.URLParam(r, "id"),
	}
}

func (r getRequest) validate() (getRequest, error) {
	if err := validate.Struct(r); err != nil {
		return getRequest{}, apperror.NewValidationError(err, "GetRequest")
	}
	return r, nil
}

type listRequest struct {
	ID          string `json:"id" validate:"omitempty,uuid"`
	Title       string `json:"title" validate:"omitempty,min=3,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Status      string `json:"status" validate:"omitempty"`
}

func newListRequest(r *http.Request) listRequest {
	return listRequest{
		ID:          r.URL.Query().Get("id"),
		Title:       r.URL.Query().Get("title"),
		Description: r.URL.Query().Get("description"),
		Status:      r.URL.Query().Get("status"),
	}
}

func (r listRequest) validate() (listRequest, error) {
	r.Title = sanitize.Sanitize(r.Title)
	r.Description = sanitize.Sanitize(r.Description)

	if err := validate.Struct(r); err != nil {
		return listRequest{}, apperror.NewValidationError(err, "listRequest")
	}
	return r, nil
}

type postRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
}

func newPostRequest(r *http.Request) (postRequest, error) {
	var req postRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return postRequest{}, apperror.NewBadRequestError(err, "postRequest")
	}
	return req, nil
}

func (r postRequest) validate() (postRequest, error) {
	r.Title = sanitize.Sanitize(r.Title)
	r.Description = sanitize.Sanitize(r.Description)

	if err := validate.Struct(r); err != nil {
		return postRequest{}, apperror.NewValidationError(err, "postRequest")
	}
	return r, nil
}

type putRequest struct {
	ID          string `json:"-" validate:"required,uuid"`
	Title       string `json:"title" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	Status      string `json:"status" validate:"omitempty,oneof=pending completed"`
}

func newPutRequest(r *http.Request) (putRequest, error) {
	var req putRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return putRequest{}, apperror.NewBadRequestError(err, "putRequest")
	}
	req.ID = chi.URLParam(r, "id")
	return req, nil
}

func (r putRequest) validate() (putRequest, error) {
	r.Title = sanitize.Sanitize(r.Title)
	r.Description = sanitize.Sanitize(r.Description)

	if err := validate.Struct(r); err != nil {
		return putRequest{}, apperror.NewValidationError(err, "putRequest")
	}
	return r, nil
}

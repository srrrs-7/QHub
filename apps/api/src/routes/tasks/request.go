package tasks

import (
	"api/src/domain/apperror"
	"encoding/json"
	"net/http"
	"utils/types"

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

func (r getRequest) validate() types.Result[getRequest, apperror.AppError] {
	if err := validate.Struct(r); err != nil {
		return types.Err[getRequest, apperror.AppError](
			apperror.NewValidationError(err, "GetRequest"),
		)
	}
	return types.Ok[getRequest, apperror.AppError](r)
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

func (r listRequest) validate() types.Result[listRequest, apperror.AppError] {
	r.Title = sanitize.Sanitize(r.Title)
	r.Description = sanitize.Sanitize(r.Description)

	if err := validate.Struct(r); err != nil {
		return types.Err[listRequest, apperror.AppError](
			apperror.NewValidationError(err, "listRequest"),
		)
	}
	return types.Ok[listRequest, apperror.AppError](r)
}

type postRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
}

func newPostRequest(r *http.Request) types.Result[postRequest, apperror.AppError] {
	var req postRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return types.Err[postRequest, apperror.AppError](
			apperror.NewBadRequestError(err, "postRequest"),
		)
	}
	return types.Ok[postRequest, apperror.AppError](req)
}

func (r postRequest) validate() types.Result[postRequest, apperror.AppError] {
	r.Title = sanitize.Sanitize(r.Title)
	r.Description = sanitize.Sanitize(r.Description)

	if err := validate.Struct(r); err != nil {
		return types.Err[postRequest, apperror.AppError](
			apperror.NewValidationError(err, "postRequest"),
		)
	}
	return types.Ok[postRequest, apperror.AppError](r)
}

type putRequest struct {
	ID          string `json:"-" validate:"required,uuid"`
	Title       string `json:"title" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	Status      string `json:"status" validate:"omitempty,oneof=pending completed"`
}

func newPutRequest(r *http.Request) types.Result[putRequest, apperror.AppError] {
	var req putRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return types.Err[putRequest, apperror.AppError](
			apperror.NewBadRequestError(err, "putRequest"),
		)
	}
	// Get ID from URL path parameter
	req.ID = chi.URLParam(r, "id")
	return types.Ok[putRequest, apperror.AppError](req)
}

func (r putRequest) validate() types.Result[putRequest, apperror.AppError] {
	r.Title = sanitize.Sanitize(r.Title)
	r.Description = sanitize.Sanitize(r.Description)

	if err := validate.Struct(r); err != nil {
		return types.Err[putRequest, apperror.AppError](
			apperror.NewValidationError(err, "putRequest"),
		)
	}
	return types.Ok[putRequest, apperror.AppError](r)
}

package organizations

import (
	"api/src/domain/apperror"
	"api/src/routes/requtil"
	"net/http"
)

// --- POST ---

type postRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Slug string `json:"slug" validate:"required,min=2,max=50"`
}

func decodePostRequest(r *http.Request) (postRequest, error) {
	return requtil.Decode[postRequest](r, func(req *postRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
	})
}

// --- GET (by slug) ---

type getRequest struct {
	Slug string `validate:"required,min=2,max=50"`
}

func newGetRequest(slug string) (getRequest, error) {
	req := getRequest{Slug: slug}
	if err := requtil.Validate.Struct(req); err != nil {
		return getRequest{}, apperror.NewValidationError(err, "getRequest")
	}
	return req, nil
}

// --- PUT ---

type putRequest struct {
	Name string `json:"name" validate:"omitempty,min=2,max=100"`
	Slug string `json:"slug" validate:"omitempty,min=2,max=50"`
	Plan string `json:"plan" validate:"omitempty,oneof=free pro team enterprise"`
}

func decodePutRequest(r *http.Request) (putRequest, error) {
	return requtil.Decode[putRequest](r, func(req *putRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
	})
}

package organizations

import (
	"api/src/routes/requtil"
	"net/http"
)

// --- POST ---

type postRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Slug string `json:"slug" validate:"required,min=2,max=50"`
}

func decodePostRequest(r *http.Request) (postRequest, error) {
	return requtil.Decode(r, func(req *postRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
	})
}

// --- PUT ---

type putRequest struct {
	Name string `json:"name" validate:"omitempty,min=2,max=100"`
	Slug string `json:"slug" validate:"omitempty,min=2,max=50"`
	Plan string `json:"plan" validate:"omitempty,oneof=free pro team enterprise"`
}

func decodePutRequest(r *http.Request) (putRequest, error) {
	return requtil.Decode(r, func(req *putRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
	})
}

package projects

import (
	"api/src/routes/requtil"
	"net/http"
)

// --- POST ---

type postRequest struct {
	OrganizationID string `json:"organization_id" validate:"required,uuid"`
	Name           string `json:"name" validate:"required,min=2,max=100"`
	Slug           string `json:"slug" validate:"required,min=2,max=50"`
	Description    string `json:"description" validate:"omitempty,max=500"`
}

func decodePostRequest(r *http.Request) (postRequest, error) {
	return requtil.Decode(r, func(req *postRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
		req.Description = requtil.Sanitize.Sanitize(req.Description)
	})
}

// --- PUT ---

type putRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=100"`
	Slug        string `json:"slug" validate:"omitempty,min=2,max=50"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

func decodePutRequest(r *http.Request) (putRequest, error) {
	return requtil.Decode(r, func(req *putRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
		req.Description = requtil.Sanitize.Sanitize(req.Description)
	})
}

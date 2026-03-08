package apikeys

import (
	"api/src/routes/requtil"
	"net/http"
)

// postRequest is the request body for creating an API key.
type postRequest struct {
	OrganizationID string `json:"organization_id" validate:"required,uuid"`
	Name           string `json:"name" validate:"required,min=1,max=100"`
}

func decodePostRequest(r *http.Request) (postRequest, error) {
	return requtil.Decode(r, func(req *postRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
	})
}

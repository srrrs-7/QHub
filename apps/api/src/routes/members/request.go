package members

import (
	"api/src/routes/requtil"
	"net/http"
)

// postRequest is the request body for adding a member to an organization.
type postRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Role   string `json:"role" validate:"required,oneof=owner admin member viewer"`
}

func decodePostRequest(r *http.Request) (postRequest, error) {
	return requtil.Decode(r, func(req *postRequest) {
		req.Role = requtil.Sanitize.Sanitize(req.Role)
	})
}

// putRequest is the request body for updating a member's role.
type putRequest struct {
	Role string `json:"role" validate:"required,oneof=owner admin member viewer"`
}

func decodePutRequest(r *http.Request) (putRequest, error) {
	return requtil.Decode(r, func(req *putRequest) {
		req.Role = requtil.Sanitize.Sanitize(req.Role)
	})
}

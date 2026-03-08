package projects

import (
	"api/src/domain/apperror"
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
	return requtil.Decode[postRequest](r, func(req *postRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
		req.Description = requtil.Sanitize.Sanitize(req.Description)
	})
}

// --- GET (by slug) ---

type getRequest struct {
	OrgID string `validate:"required,uuid"`
	Slug  string `validate:"required,min=2,max=50"`
}

func newGetRequest(orgID, slug string) (getRequest, error) {
	req := getRequest{OrgID: orgID, Slug: slug}
	if err := requtil.Validate.Struct(req); err != nil {
		return getRequest{}, apperror.NewValidationError(err, "getRequest")
	}
	return req, nil
}

// --- LIST ---

type listRequest struct {
	OrgID string `validate:"required,uuid"`
}

func newListRequest(orgID string) (listRequest, error) {
	req := listRequest{OrgID: orgID}
	if err := requtil.Validate.Struct(req); err != nil {
		return listRequest{}, apperror.NewValidationError(err, "listRequest")
	}
	return req, nil
}

// --- PUT ---

type putRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=100"`
	Slug        string `json:"slug" validate:"omitempty,min=2,max=50"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

func decodePutRequest(r *http.Request) (putRequest, error) {
	return requtil.Decode[putRequest](r, func(req *putRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
		req.Description = requtil.Sanitize.Sanitize(req.Description)
	})
}

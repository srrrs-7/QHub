package tags

import (
	"api/src/routes/requtil"
	"net/http"
)

type postTagRequest struct {
	OrgID string `json:"org_id" validate:"required,uuid"`
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Color string `json:"color" validate:"required,min=1,max=20"`
}

func decodePostTagRequest(r *http.Request) (postTagRequest, error) {
	return requtil.Decode(r, func(req *postTagRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Color = requtil.Sanitize.Sanitize(req.Color)
	})
}

type addPromptTagRequest struct {
	TagID string `json:"tag_id" validate:"required,uuid"`
}

func decodeAddPromptTagRequest(r *http.Request) (addPromptTagRequest, error) {
	return requtil.Decode[addPromptTagRequest](r, nil)
}

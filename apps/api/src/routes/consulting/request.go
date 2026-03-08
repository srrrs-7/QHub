package consulting

import (
	"api/src/routes/requtil"
	"encoding/json"
	"net/http"
)

type postSessionRequest struct {
	OrgID            string `json:"org_id" validate:"required,uuid"`
	Title            string `json:"title" validate:"required,min=1,max=200"`
	IndustryConfigID string `json:"industry_config_id" validate:"omitempty,uuid"`
}

func decodePostSessionRequest(r *http.Request) (postSessionRequest, error) {
	return requtil.Decode(r, func(req *postSessionRequest) {
		req.Title = requtil.Sanitize.Sanitize(req.Title)
	})
}

type postMessageRequest struct {
	Role         string          `json:"role" validate:"required,oneof=user assistant system"`
	Content      string          `json:"content" validate:"required,min=1"`
	Citations    json.RawMessage `json:"citations"`
	ActionsTaken json.RawMessage `json:"actions_taken"`
}

func decodePostMessageRequest(r *http.Request) (postMessageRequest, error) {
	return requtil.Decode(r, func(req *postMessageRequest) {
		req.Content = requtil.Sanitize.Sanitize(req.Content)
	})
}

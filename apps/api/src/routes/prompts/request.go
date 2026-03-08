package prompts

import (
	"api/src/routes/requtil"
	"encoding/json"
	"net/http"
)

// --- POST Prompt ---

type postPromptRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=200"`
	Slug        string `json:"slug" validate:"required,min=2,max=80"`
	PromptType  string `json:"prompt_type" validate:"required,oneof=system user combined"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}

func decodePostPromptRequest(r *http.Request) (postPromptRequest, error) {
	return requtil.Decode(r, func(req *postPromptRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
		req.Description = requtil.Sanitize.Sanitize(req.Description)
	})
}

// --- PUT Prompt ---

type putPromptRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=200"`
	Slug        string `json:"slug" validate:"omitempty,min=2,max=80"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}

func decodePutPromptRequest(r *http.Request) (putPromptRequest, error) {
	return requtil.Decode(r, func(req *putPromptRequest) {
		req.Name = requtil.Sanitize.Sanitize(req.Name)
		req.Slug = requtil.Sanitize.Sanitize(req.Slug)
		req.Description = requtil.Sanitize.Sanitize(req.Description)
	})
}

// --- POST Version ---

type postVersionRequest struct {
	Content           json.RawMessage `json:"content" validate:"required"`
	Variables         json.RawMessage `json:"variables"`
	ChangeDescription string          `json:"change_description" validate:"omitempty,max=500"`
	AuthorID          string          `json:"author_id" validate:"required,uuid"`
}

func decodePostVersionRequest(r *http.Request) (postVersionRequest, error) {
	return requtil.Decode(r, func(req *postVersionRequest) {
		req.ChangeDescription = requtil.Sanitize.Sanitize(req.ChangeDescription)
	})
}

// --- PUT Version Status ---

type putVersionStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft review production archived"`
}

func decodePutVersionStatusRequest(r *http.Request) (putVersionStatusRequest, error) {
	return requtil.Decode[putVersionStatusRequest](r, nil)
}

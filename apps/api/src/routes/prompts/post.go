package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *PromptHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requtil.ParseUUID(r, "project_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePostPromptRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		name, err := prompt.NewPromptName(req.Name)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		slug, err := prompt.NewPromptSlug(req.Slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		promptType, err := prompt.NewPromptType(req.PromptType)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		desc, err := prompt.NewPromptDescription(req.Description)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		p, err := h.promptRepo.Create(r.Context(), prompt.PromptCmd{
			ProjectID: projectID, Name: name, Slug: slug, PromptType: promptType, Description: desc,
		})
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toPromptResponse(p))
	}
}

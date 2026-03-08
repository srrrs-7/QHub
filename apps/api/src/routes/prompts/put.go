package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *PromptHandler) Put() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requtil.ParseUUID(r, "project_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		slug, err := prompt.NewPromptSlug(chi.URLParam(r, "prompt_slug"))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		existing, err := h.promptRepo.FindByProjectAndSlug(r.Context(), projectID, slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePutPromptRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		name, err := requtil.MergeField(existing.Name, req.Name, prompt.NewPromptName)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		newSlug, err := requtil.MergeField(existing.Slug, req.Slug, prompt.NewPromptSlug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		desc, err := requtil.MergeField(existing.Description, req.Description, prompt.NewPromptDescription)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		updated, err := h.promptRepo.Update(r.Context(), existing.ID, prompt.PromptCmd{
			ProjectID: projectID, Name: name, Slug: newSlug, PromptType: existing.PromptType, Description: desc,
		})
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toPromptResponse(updated))
	}
}

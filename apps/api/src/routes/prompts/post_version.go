package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PromptHandler) PostVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")

		parsedPromptID, err := uuid.Parse(promptID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePostVersionRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		changeDesc, err := prompt.NewChangeDescription(req.ChangeDescription)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		authorID, err := uuid.Parse(req.AuthorID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Get prompt to determine next version number
		p, err := h.promptRepo.FindByID(r.Context(), prompt.PromptIDFromUUID(parsedPromptID))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		cmd := prompt.VersionCmd{
			PromptID:          prompt.PromptIDFromUUID(parsedPromptID),
			Content:           req.Content,
			Variables:         req.Variables,
			ChangeDescription: changeDesc,
			AuthorID:          authorID,
		}

		nextVersion := p.LatestVersion + 1
		v, err := h.versionRepo.Create(r.Context(), cmd, nextVersion)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Update prompt's latest_version
		_, err = h.promptRepo.UpdateLatestVersion(r.Context(), p.ID, nextVersion)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toVersionResponse(v))
	}
}

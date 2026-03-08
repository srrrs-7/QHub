package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"github.com/google/uuid"
)

func (h *PromptHandler) PostVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
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

		p, err := h.promptRepo.FindByID(r.Context(), prompt.PromptIDFromUUID(promptID))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		nextVersion := p.LatestVersion + 1
		v, err := h.versionRepo.Create(r.Context(), prompt.VersionCmd{
			PromptID: prompt.PromptIDFromUUID(promptID), Content: req.Content,
			Variables: req.Variables, ChangeDescription: changeDesc, AuthorID: authorID,
		}, nextVersion)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		if _, err = h.promptRepo.UpdateLatestVersion(r.Context(), p.ID, nextVersion); err != nil {
			response.HandleError(w, err)
			return
		}

		// Generate embedding asynchronously (fire-and-forget)
		if h.embeddingSvc != nil {
			h.embeddingSvc.EmbedVersionAsync(v.ID.UUID(), v.Content)
		}

		response.Created(w, toVersionResponse(v))
	}
}

package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PromptHandler) PutVersionStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")
		versionParam := chi.URLParam(r, "version")

		parsedPromptID, err := uuid.Parse(promptID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		number, err := strconv.Atoi(versionParam)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePutVersionStatusRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Get existing version to validate transition
		existing, err := h.versionRepo.FindByPromptAndNumber(r.Context(), prompt.PromptIDFromUUID(parsedPromptID), number)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		newStatus, err := prompt.NewVersionStatus(req.Status)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Validate status transition
		if err := prompt.ValidateStatusTransition(existing.Status, newStatus); err != nil {
			response.HandleError(w, err)
			return
		}

		// If promoting to production, archive current production version
		if newStatus == prompt.StatusProduction {
			if err := h.versionRepo.ArchiveProduction(r.Context(), prompt.PromptIDFromUUID(parsedPromptID)); err != nil {
				response.HandleError(w, err)
				return
			}
		}

		updated, err := h.versionRepo.UpdateStatus(r.Context(), existing.ID, newStatus)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Update prompt's production_version
		if newStatus == prompt.StatusProduction {
			v := existing.VersionNumber
			_, err = h.promptRepo.UpdateProductionVersion(r.Context(), prompt.PromptIDFromUUID(parsedPromptID), &v)
			if err != nil {
				response.HandleError(w, err)
				return
			}
		}

		response.OK(w, toVersionResponse(updated))
	}
}

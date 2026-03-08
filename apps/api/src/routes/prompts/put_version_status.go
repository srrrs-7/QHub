package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *PromptHandler) PutVersionStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		number, err := strconv.Atoi(chi.URLParam(r, "version"))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePutVersionStatusRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		pid := prompt.PromptIDFromUUID(promptID)
		existing, err := h.versionRepo.FindByPromptAndNumber(r.Context(), pid, number)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		newStatus, err := prompt.NewVersionStatus(req.Status)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		if err := prompt.ValidateStatusTransition(existing.Status, newStatus); err != nil {
			response.HandleError(w, err)
			return
		}

		if newStatus == prompt.StatusProduction {
			if err := h.versionRepo.ArchiveProduction(r.Context(), pid); err != nil {
				response.HandleError(w, err)
				return
			}
		}

		updated, err := h.versionRepo.UpdateStatus(r.Context(), existing.ID, newStatus)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		if newStatus == prompt.StatusProduction {
			v := existing.VersionNumber
			if _, err = h.promptRepo.UpdateProductionVersion(r.Context(), pid, &v); err != nil {
				response.HandleError(w, err)
				return
			}
		}

		response.OK(w, toVersionResponse(updated))
	}
}

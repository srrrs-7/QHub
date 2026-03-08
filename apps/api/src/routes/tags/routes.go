package tags

import (
	"api/src/domain/tag"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *TagHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodePostTagRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		orgID, err := uuid.Parse(req.OrgID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		t, err := h.tagRepo.Create(r.Context(), tag.Tag{
			OrgID: orgID,
			Name:  req.Name,
			Color: req.Color,
		})
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toTagResponse(t))
	}
}

func (h *TagHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgIDStr := chi.URLParam(r, "org_id")
		if orgIDStr == "" {
			orgIDStr = r.URL.Query().Get("org_id")
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		tags, err := h.tagRepo.FindAllByOrg(r.Context(), orgID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(tags, toTagResponse))
	}
}

func (h *TagHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requtil.ParseUUID(r, "id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		if err := h.tagRepo.Delete(r.Context(), id); err != nil {
			response.HandleError(w, err)
			return
		}

		response.NoContent(w)
	}
}

func (h *TagHandler) AddToPrompt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodeAddPromptTagRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		tagID, err := uuid.Parse(req.TagID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		if err := h.tagRepo.AddToPrompt(r.Context(), promptID, tagID); err != nil {
			response.HandleError(w, err)
			return
		}

		response.NoContent(w)
	}
}

func (h *TagHandler) RemoveFromPrompt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		tagID, err := requtil.ParseUUID(r, "tag_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		if err := h.tagRepo.RemoveFromPrompt(r.Context(), promptID, tagID); err != nil {
			response.HandleError(w, err)
			return
		}

		response.NoContent(w)
	}
}

func (h *TagHandler) ListByPrompt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		tags, err := h.tagRepo.FindByPrompt(r.Context(), promptID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(tags, toTagResponse))
	}
}

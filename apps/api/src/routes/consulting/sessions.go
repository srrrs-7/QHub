package consulting

import (
	domain "api/src/domain/consulting"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *ConsultingHandler) PostSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodePostSessionRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		orgID, err := uuid.Parse(req.OrgID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		session := domain.Session{
			OrgID: orgID,
			Title: req.Title,
		}

		if req.IndustryConfigID != "" {
			id, err := uuid.Parse(req.IndustryConfigID)
			if err != nil {
				response.HandleError(w, err)
				return
			}
			session.IndustryConfigID = &id
		}

		created, err := h.sessionRepo.Create(r.Context(), session)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toSessionResponse(created))
	}
}

func (h *ConsultingHandler) GetSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requtil.ParseUUID(r, "id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		session, err := h.sessionRepo.FindByID(r.Context(), id)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toSessionResponse(session))
	}
}

func (h *ConsultingHandler) ListSessions() http.HandlerFunc {
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

		limit := 20
		offset := 0
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		sessions, err := h.sessionRepo.FindAllByOrg(r.Context(), orgID, limit, offset)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(sessions, toSessionResponse))
	}
}

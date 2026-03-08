package members

import (
	"api/src/domain/apperror"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

// List returns all members for the given organization.
func (h *MemberHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID, err := requtil.ParseUUID(r, "org_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		members, err := h.q.ListOrganizationMembers(r.Context(), orgID)
		if err != nil {
			response.HandleError(w, apperror.NewDatabaseError(err, "member"))
			return
		}

		response.OK(w, response.MapSlice(members, toMemberResponse))
	}
}

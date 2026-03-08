package members

import (
	"api/src/domain/apperror"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"utils/db/db"
)

// Delete removes a member from an organization.
func (h *MemberHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID, err := requtil.ParseUUID(r, "org_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		userID, err := requtil.ParseUUID(r, "user_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		err = h.q.RemoveOrganizationMember(r.Context(), db.RemoveOrganizationMemberParams{
			OrganizationID: orgID,
			UserID:         userID,
		})
		if err != nil {
			response.HandleError(w, apperror.NewDatabaseError(err, "member"))
			return
		}

		response.NoContent(w)
	}
}

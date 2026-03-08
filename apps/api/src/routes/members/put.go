package members

import (
	"api/src/domain/apperror"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"utils/db/db"
)

// Put updates a member's role in an organization.
func (h *MemberHandler) Put() http.HandlerFunc {
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

		req, err := decodePutRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		member, err := h.q.UpdateOrganizationMemberRole(r.Context(), db.UpdateOrganizationMemberRoleParams{
			OrganizationID: orgID,
			UserID:         userID,
			Role:           req.Role,
		})
		if err != nil {
			response.HandleError(w, apperror.NewDatabaseError(err, "member"))
			return
		}

		response.OK(w, toMemberResponse(member))
	}
}

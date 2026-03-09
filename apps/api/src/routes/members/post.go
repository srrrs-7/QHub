package members

import (
	"api/src/domain/apperror"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"github.com/google/uuid"

	"utils/db/db"
)

// Post adds a member to an organization.
func (h *MemberHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID, err := requtil.ParseUUID(r, "org_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePostRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			response.HandleError(w, apperror.NewValidationError(err, "user_id"))
			return
		}

		member, err := h.q.AddOrganizationMember(r.Context(), db.AddOrganizationMemberParams{
			OrganizationID: orgID,
			UserID:         userID,
			Role:           req.Role,
		})
		if err != nil {
			response.HandleError(w, apperror.NewDatabaseError(err, "member"))
			return
		}

		response.Created(w, toMemberResponse(member))
	}
}

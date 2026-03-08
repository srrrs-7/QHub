package members

import (
	"time"
	"utils/db/db"
)

// memberResponse is the response for an organization member.
type memberResponse struct {
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	Role           string `json:"role"`
	JoinedAt       string `json:"joined_at"`
}

func toMemberResponse(m db.OrganizationMember) memberResponse {
	return memberResponse{
		OrganizationID: m.OrganizationID.String(),
		UserID:         m.UserID.String(),
		Role:           m.Role,
		JoinedAt:       m.JoinedAt.Format(time.RFC3339),
	}
}

package members

import (
	"utils/db/db"
)

// MemberHandler handles organization member management endpoints.
type MemberHandler struct {
	q db.Querier
}

// NewMemberHandler creates a new MemberHandler with the given querier.
func NewMemberHandler(q db.Querier) *MemberHandler {
	return &MemberHandler{q: q}
}

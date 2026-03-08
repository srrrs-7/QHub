package tags

import (
	"api/src/domain/tag"
	"time"
)

type tagResponse struct {
	ID        string `json:"id"`
	OrgID     string `json:"org_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
}

func toTagResponse(t tag.Tag) tagResponse {
	return tagResponse{
		ID:        t.ID.String(),
		OrgID:     t.OrgID.String(),
		Name:      t.Name,
		Color:     t.Color,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
}

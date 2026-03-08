// Package tag defines the Tag entity used for categorising prompts.
//
// Tags belong to an organization and can be attached to or detached from
// prompts via the TagRepository.
package tag

import (
	"time"

	"github.com/google/uuid"
)

// Tag represents a label that can be attached to prompts.
type Tag struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	Name      string
	Color     string // hex colour code, e.g. "#ff5733"
	CreatedAt time.Time
}

package tag

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	Name      string
	Color     string
	CreatedAt time.Time
}

package consulting

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID               uuid.UUID
	OrgID            uuid.UUID
	Title            string
	IndustryConfigID *uuid.UUID
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Message struct {
	ID           uuid.UUID
	SessionID    uuid.UUID
	Role         string
	Content      string
	Citations    json.RawMessage
	ActionsTaken json.RawMessage
	CreatedAt    time.Time
}

type IndustryConfig struct {
	ID              uuid.UUID
	Slug            string
	Name            string
	Description     string
	KnowledgeBase   json.RawMessage
	ComplianceRules json.RawMessage
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

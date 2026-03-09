// Package consulting defines domain entities for the AI consulting feature.
//
// Sessions represent individual consulting conversations. Each session
// contains messages (user and assistant turns) and may be associated
// with an IndustryConfig for domain-specific knowledge and compliance.
package consulting

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Session represents an AI consulting conversation.
type Session struct {
	ID               uuid.UUID
	OrgID            uuid.UUID
	Title            string
	IndustryConfigID *uuid.UUID // nil when no industry is associated
	Status           string     // "active", "closed"
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Message represents a single turn in a consulting session.
type Message struct {
	ID           uuid.UUID
	SessionID    uuid.UUID
	Role         string // "user" or "assistant"
	Content      string
	Citations    json.RawMessage // optional reference citations
	ActionsTaken json.RawMessage // optional actions performed
	CreatedAt    time.Time
}

// IndustryConfig stores domain-specific knowledge and compliance rules
// for a particular industry vertical (e.g. healthcare, finance).
type IndustryConfig struct {
	ID              uuid.UUID
	Slug            string
	Name            string
	Description     string
	KnowledgeBase   json.RawMessage // industry knowledge corpus
	ComplianceRules json.RawMessage // regulatory constraints
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

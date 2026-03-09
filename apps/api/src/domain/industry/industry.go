package industry

import (
	"encoding/json"
	"time"
)

// IndustryConfig is the aggregate representing an industry configuration
// with domain-specific knowledge and compliance rules.
type IndustryConfig struct {
	ID              IndustryID
	Slug            IndustrySlug
	Name            IndustryName
	Description     IndustryDescription
	KnowledgeBase   json.RawMessage
	ComplianceRules json.RawMessage
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewIndustryConfig constructs an IndustryConfig from validated value objects.
// Typically used when reconstructing from DB results.
func NewIndustryConfig(
	id IndustryID,
	slug IndustrySlug,
	name IndustryName,
	description IndustryDescription,
	knowledgeBase json.RawMessage,
	complianceRules json.RawMessage,
	createdAt time.Time,
	updatedAt time.Time,
) IndustryConfig {
	return IndustryConfig{
		ID:              id,
		Slug:            slug,
		Name:            name,
		Description:     description,
		KnowledgeBase:   knowledgeBase,
		ComplianceRules: complianceRules,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}

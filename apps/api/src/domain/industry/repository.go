package industry

import "context"

// IndustryConfigRepository defines persistence operations for industry configurations.
// Implementations live in the infra layer (dependency inversion).
type IndustryConfigRepository interface {
	FindBySlug(ctx context.Context, slug IndustrySlug) (IndustryConfig, error)
	FindAll(ctx context.Context) ([]IndustryConfig, error)
	Create(ctx context.Context, slug IndustrySlug, name IndustryName, description IndustryDescription) (IndustryConfig, error)
}

package apikeys

import (
	"time"
	"utils/db/db"
)

// apiKeyResponse is the standard response for an API key (does not include the raw key).
type apiKeyResponse struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	Name           string  `json:"name"`
	KeyPrefix      string  `json:"key_prefix"`
	LastUsedAt     *string `json:"last_used_at"`
	ExpiresAt      *string `json:"expires_at"`
	RevokedAt      *string `json:"revoked_at"`
	CreatedAt      string  `json:"created_at"`
}

// apiKeyCreatedResponse is returned when a new API key is created.
// It includes the raw key which is only shown once.
type apiKeyCreatedResponse struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	Name           string  `json:"name"`
	Key            string  `json:"key"`
	KeyPrefix      string  `json:"key_prefix"`
	ExpiresAt      *string `json:"expires_at"`
	CreatedAt      string  `json:"created_at"`
}

func toApiKeyResponse(k db.ApiKey) apiKeyResponse {
	resp := apiKeyResponse{
		ID:             k.ID.String(),
		OrganizationID: k.OrganizationID.String(),
		Name:           k.Name,
		KeyPrefix:      k.KeyPrefix,
		CreatedAt:      k.CreatedAt.Format(time.RFC3339),
	}
	if k.LastUsedAt.Valid {
		s := k.LastUsedAt.Time.Format(time.RFC3339)
		resp.LastUsedAt = &s
	}
	if k.ExpiresAt.Valid {
		s := k.ExpiresAt.Time.Format(time.RFC3339)
		resp.ExpiresAt = &s
	}
	if k.RevokedAt.Valid {
		s := k.RevokedAt.Time.Format(time.RFC3339)
		resp.RevokedAt = &s
	}
	return resp
}

func toApiKeyCreatedResponse(k db.ApiKey, rawKey string) apiKeyCreatedResponse {
	resp := apiKeyCreatedResponse{
		ID:             k.ID.String(),
		OrganizationID: k.OrganizationID.String(),
		Name:           k.Name,
		Key:            rawKey,
		KeyPrefix:      k.KeyPrefix,
		CreatedAt:      k.CreatedAt.Format(time.RFC3339),
	}
	if k.ExpiresAt.Valid {
		s := k.ExpiresAt.Time.Format(time.RFC3339)
		resp.ExpiresAt = &s
	}
	return resp
}

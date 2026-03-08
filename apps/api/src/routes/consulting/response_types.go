package consulting

import (
	domain "api/src/domain/consulting"
	"encoding/json"
	"time"
)

type sessionResponse struct {
	ID               string  `json:"id"`
	OrgID            string  `json:"org_id"`
	Title            string  `json:"title"`
	IndustryConfigID *string `json:"industry_config_id"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func toSessionResponse(s domain.Session) sessionResponse {
	var industryConfigID *string
	if s.IndustryConfigID != nil {
		id := s.IndustryConfigID.String()
		industryConfigID = &id
	}
	return sessionResponse{
		ID:               s.ID.String(),
		OrgID:            s.OrgID.String(),
		Title:            s.Title,
		IndustryConfigID: industryConfigID,
		Status:           s.Status,
		CreatedAt:        s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        s.UpdatedAt.Format(time.RFC3339),
	}
}

type messageResponse struct {
	ID           string          `json:"id"`
	SessionID    string          `json:"session_id"`
	Role         string          `json:"role"`
	Content      string          `json:"content"`
	Citations    json.RawMessage `json:"citations"`
	ActionsTaken json.RawMessage `json:"actions_taken"`
	CreatedAt    string          `json:"created_at"`
}

func toMessageResponse(m domain.Message) messageResponse {
	return messageResponse{
		ID:           m.ID.String(),
		SessionID:    m.SessionID.String(),
		Role:         m.Role,
		Content:      m.Content,
		Citations:    m.Citations,
		ActionsTaken: m.ActionsTaken,
		CreatedAt:    m.CreatedAt.Format(time.RFC3339),
	}
}

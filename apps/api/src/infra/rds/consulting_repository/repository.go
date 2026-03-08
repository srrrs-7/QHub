package consulting_repository

import (
	"api/src/domain/consulting"
	"api/src/infra/rds/repoerr"
	"context"
	"database/sql"
	"encoding/json"
	"time"
	"utils/db/db"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

const dbTimeout = 5 * time.Second

// --- SessionRepository ---

type SessionRepository struct {
	q db.Querier
}

func NewSessionRepository(q db.Querier) *SessionRepository {
	return &SessionRepository{q: q}
}

var _ consulting.SessionRepository = (*SessionRepository)(nil)

func toSession(s db.ConsultingSession) consulting.Session {
	var industryConfigID *uuid.UUID
	if s.IndustryConfigID.Valid {
		industryConfigID = &s.IndustryConfigID.UUID
	}
	return consulting.Session{
		ID:               s.ID,
		OrgID:            s.OrganizationID,
		Title:            s.Title,
		IndustryConfigID: industryConfigID,
		Status:           s.Status,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}

func (r *SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (consulting.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	s, err := r.q.GetConsultingSession(ctx, id)
	if err != nil {
		return consulting.Session{}, repoerr.Handle(err, "SessionRepository", "ConsultingSession")
	}
	return toSession(s), nil
}

func (r *SessionRepository) FindAllByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]consulting.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	sessions, err := r.q.ListConsultingSessionsByOrg(ctx, db.ListConsultingSessionsByOrgParams{
		OrganizationID: orgID,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, repoerr.Handle(err, "SessionRepository", "")
	}

	result := make([]consulting.Session, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, toSession(s))
	}
	return result, nil
}

func (r *SessionRepository) Create(ctx context.Context, session consulting.Session) (consulting.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var industryConfigID uuid.NullUUID
	if session.IndustryConfigID != nil {
		industryConfigID = uuid.NullUUID{UUID: *session.IndustryConfigID, Valid: true}
	}

	s, err := r.q.CreateConsultingSession(ctx, db.CreateConsultingSessionParams{
		OrganizationID:   session.OrgID,
		Title:            session.Title,
		IndustryConfigID: industryConfigID,
	})
	if err != nil {
		return consulting.Session{}, repoerr.Handle(err, "SessionRepository", "")
	}
	return toSession(s), nil
}

func (r *SessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (consulting.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	s, err := r.q.UpdateConsultingSessionStatus(ctx, db.UpdateConsultingSessionStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return consulting.Session{}, repoerr.Handle(err, "SessionRepository", "ConsultingSession")
	}
	return toSession(s), nil
}

// --- MessageRepository ---

type MessageRepository struct {
	q db.Querier
}

func NewMessageRepository(q db.Querier) *MessageRepository {
	return &MessageRepository{q: q}
}

var _ consulting.MessageRepository = (*MessageRepository)(nil)

func toMessage(m db.ConsultingMessage) consulting.Message {
	return consulting.Message{
		ID:           m.ID,
		SessionID:    m.SessionID,
		Role:         m.Role,
		Content:      m.Content,
		Citations:    m.Citations.RawMessage,
		ActionsTaken: m.ActionsTaken.RawMessage,
		CreatedAt:    m.CreatedAt,
	}
}

func (r *MessageRepository) FindAllBySession(ctx context.Context, sessionID uuid.UUID) ([]consulting.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	messages, err := r.q.ListConsultingMessages(ctx, sessionID)
	if err != nil {
		return nil, repoerr.Handle(err, "MessageRepository", "")
	}

	result := make([]consulting.Message, 0, len(messages))
	for _, m := range messages {
		result = append(result, toMessage(m))
	}
	return result, nil
}

func (r *MessageRepository) Create(ctx context.Context, msg consulting.Message) (consulting.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	m, err := r.q.CreateConsultingMessage(ctx, db.CreateConsultingMessageParams{
		SessionID:    msg.SessionID,
		Role:         msg.Role,
		Content:      msg.Content,
		Citations:    pqtype.NullRawMessage{RawMessage: msg.Citations, Valid: msg.Citations != nil},
		ActionsTaken: pqtype.NullRawMessage{RawMessage: msg.ActionsTaken, Valid: msg.ActionsTaken != nil},
	})
	if err != nil {
		return consulting.Message{}, repoerr.Handle(err, "MessageRepository", "")
	}
	return toMessage(m), nil
}

// --- IndustryConfigRepository ---

type IndustryConfigRepository struct {
	q db.Querier
}

func NewIndustryConfigRepository(q db.Querier) *IndustryConfigRepository {
	return &IndustryConfigRepository{q: q}
}

var _ consulting.IndustryConfigRepository = (*IndustryConfigRepository)(nil)

func toIndustryConfig(c db.IndustryConfig) consulting.IndustryConfig {
	return consulting.IndustryConfig{
		ID:              c.ID,
		Slug:            c.Slug,
		Name:            c.Name,
		Description:     c.Description.String,
		KnowledgeBase:   c.KnowledgeBase,
		ComplianceRules: c.ComplianceRules,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

func (r *IndustryConfigRepository) FindByID(ctx context.Context, id uuid.UUID) (consulting.IndustryConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	c, err := r.q.GetIndustryConfig(ctx, id)
	if err != nil {
		return consulting.IndustryConfig{}, repoerr.Handle(err, "IndustryConfigRepository", "IndustryConfig")
	}
	return toIndustryConfig(c), nil
}

func (r *IndustryConfigRepository) FindBySlug(ctx context.Context, slug string) (consulting.IndustryConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	c, err := r.q.GetIndustryConfigBySlug(ctx, slug)
	if err != nil {
		return consulting.IndustryConfig{}, repoerr.Handle(err, "IndustryConfigRepository", "IndustryConfig")
	}
	return toIndustryConfig(c), nil
}

func (r *IndustryConfigRepository) FindAll(ctx context.Context) ([]consulting.IndustryConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	configs, err := r.q.ListIndustryConfigs(ctx)
	if err != nil {
		return nil, repoerr.Handle(err, "IndustryConfigRepository", "")
	}

	result := make([]consulting.IndustryConfig, 0, len(configs))
	for _, c := range configs {
		result = append(result, toIndustryConfig(c))
	}
	return result, nil
}

func (r *IndustryConfigRepository) Create(ctx context.Context, cfg consulting.IndustryConfig) (consulting.IndustryConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	desc := cfg.Description
	kb := cfg.KnowledgeBase
	if kb == nil {
		kb = json.RawMessage(`{}`)
	}
	cr := cfg.ComplianceRules
	if cr == nil {
		cr = json.RawMessage(`{}`)
	}

	c, err := r.q.CreateIndustryConfig(ctx, db.CreateIndustryConfigParams{
		Slug:            cfg.Slug,
		Name:            cfg.Name,
		Description:     sql.NullString{String: desc, Valid: desc != ""},
		KnowledgeBase:   kb,
		ComplianceRules: cr,
	})
	if err != nil {
		return consulting.IndustryConfig{}, repoerr.Handle(err, "IndustryConfigRepository", "")
	}
	return toIndustryConfig(c), nil
}

func (r *IndustryConfigRepository) Update(ctx context.Context, cfg consulting.IndustryConfig) (consulting.IndustryConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	c, err := r.q.UpdateIndustryConfig(ctx, db.UpdateIndustryConfigParams{
		ID:              cfg.ID,
		Name:            cfg.Name,
		Description:     cfg.Description,
		KnowledgeBase:   cfg.KnowledgeBase,
		ComplianceRules: cfg.ComplianceRules,
	})
	if err != nil {
		return consulting.IndustryConfig{}, repoerr.Handle(err, "IndustryConfigRepository", "IndustryConfig")
	}
	return toIndustryConfig(c), nil
}

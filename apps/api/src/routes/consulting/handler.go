package consulting

import (
	domain "api/src/domain/consulting"
)

// ConsultingHandler handles consulting session and message endpoints.
type ConsultingHandler struct {
	sessionRepo  domain.SessionRepository
	messageRepo  domain.MessageRepository
	industryRepo domain.IndustryConfigRepository
}

// NewConsultingHandler creates a new ConsultingHandler.
func NewConsultingHandler(
	sessionRepo domain.SessionRepository,
	messageRepo domain.MessageRepository,
	industryRepo domain.IndustryConfigRepository,
) *ConsultingHandler {
	return &ConsultingHandler{
		sessionRepo:  sessionRepo,
		messageRepo:  messageRepo,
		industryRepo: industryRepo,
	}
}

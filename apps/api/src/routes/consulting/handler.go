package consulting

import (
	domain "api/src/domain/consulting"
)

type ConsultingHandler struct {
	sessionRepo  domain.SessionRepository
	messageRepo  domain.MessageRepository
	industryRepo domain.IndustryConfigRepository
}

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

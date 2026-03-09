package consulting

import (
	domain "api/src/domain/consulting"
	"api/src/services/ragservice"
)

// ConsultingHandler handles consulting session and message endpoints.
type ConsultingHandler struct {
	sessionRepo  domain.SessionRepository
	messageRepo  domain.MessageRepository
	industryRepo domain.IndustryConfigRepository
	ragSvc       *ragservice.RAGService
}

// NewConsultingHandler creates a new ConsultingHandler.
// ragSvc may be nil if RAG is not configured.
func NewConsultingHandler(
	sessionRepo domain.SessionRepository,
	messageRepo domain.MessageRepository,
	industryRepo domain.IndustryConfigRepository,
	ragSvc *ragservice.RAGService,
) *ConsultingHandler {
	return &ConsultingHandler{
		sessionRepo:  sessionRepo,
		messageRepo:  messageRepo,
		industryRepo: industryRepo,
		ragSvc:       ragSvc,
	}
}

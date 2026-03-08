package industries

import (
	domain "api/src/domain/consulting"
	"utils/db/db"
)

type IndustryHandler struct {
	industryRepo domain.IndustryConfigRepository
	q            db.Querier
}

func NewIndustryHandler(industryRepo domain.IndustryConfigRepository, q db.Querier) *IndustryHandler {
	return &IndustryHandler{industryRepo: industryRepo, q: q}
}

package analytics

import db "utils/db/db"

type AnalyticsHandler struct {
	q db.Querier
}

func NewAnalyticsHandler(q db.Querier) *AnalyticsHandler {
	return &AnalyticsHandler{q: q}
}

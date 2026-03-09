package analytics

import (
	"api/src/services/statsservice"
	db "utils/db/db"
)

// AnalyticsHandler serves analytics and statistical comparison endpoints.
type AnalyticsHandler struct {
	q     db.Querier
	stats *statsservice.StatsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler with the given querier.
// A StatsService is automatically created from the same querier.
func NewAnalyticsHandler(q db.Querier) *AnalyticsHandler {
	return &AnalyticsHandler{
		q:     q,
		stats: statsservice.NewStatsService(q),
	}
}

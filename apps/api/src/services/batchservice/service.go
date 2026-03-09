// Package batchservice performs monthly batch aggregation of execution log
// metrics across organizations. It computes average latency, total tokens,
// average scores, and active prompt counts for a given month.
package batchservice

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"api/src/domain/apperror"

	"github.com/google/uuid"

	db "utils/db/db"
)

// domainName is the domain identifier used in error reporting.
const domainName = "BatchService"

// BatchService aggregates execution log metrics on a monthly basis.
type BatchService struct {
	q db.Querier
}

// NewBatchService creates a new BatchService with the given querier.
func NewBatchService(q db.Querier) *BatchService {
	return &BatchService{q: q}
}

// AggregationResult summarises a completed monthly aggregation run.
type AggregationResult struct {
	OrgsProcessed       int           `json:"orgs_processed"`
	PromptsProcessed    int           `json:"prompts_processed"`
	ExecutionsProcessed int           `json:"executions_processed"`
	Period              string        `json:"period"`
	Duration            time.Duration `json:"duration"`
}

// OrgMetrics holds aggregated execution metrics for a single organization.
type OrgMetrics struct {
	OrgID          uuid.UUID `json:"org_id"`
	AvgLatency     float64   `json:"avg_latency"`
	TotalTokens    int64     `json:"total_tokens"`
	AvgScore       float64   `json:"avg_score"`
	ExecutionCount int64     `json:"execution_count"`
	ActivePrompts  int64     `json:"active_prompts"`
}

// period returns the YYYY-MM formatted period string for the previous month
// relative to the given time.
func period(now time.Time) string {
	prev := now.AddDate(0, -1, 0)
	return prev.Format("2006-01")
}

// monthBounds returns the start (inclusive) and end (exclusive) timestamps for
// the month described by a YYYY-MM period string.
func monthBounds(p string) (start, end time.Time, err error) {
	t, err := time.Parse("2006-01", p)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period %q: %w", p, err)
	}
	start = t
	end = t.AddDate(0, 1, 0)
	return start, end, nil
}

// RunMonthlyAggregation computes metrics for the previous calendar month
// across all organizations. It iterates over every organization and aggregates
// execution log data for the period.
func (s *BatchService) RunMonthlyAggregation(ctx context.Context) (*AggregationResult, error) {
	startTime := time.Now()
	p := period(startTime)

	orgs, err := s.q.ListAllOrganizations(ctx)
	if err != nil {
		return nil, apperror.NewDatabaseError(
			fmt.Errorf("failed to list organizations: %w", err),
			domainName,
		)
	}

	totalPrompts := int64(0)
	totalExecutions := int64(0)
	orgsProcessed := 0

	for _, org := range orgs {
		metrics, err := s.ComputeOrgMetrics(ctx, org.ID)
		if err != nil {
			return nil, err
		}

		if metrics.ExecutionCount > 0 {
			orgsProcessed++
			totalPrompts += metrics.ActivePrompts
			totalExecutions += metrics.ExecutionCount
		}
	}

	return &AggregationResult{
		OrgsProcessed:       orgsProcessed,
		PromptsProcessed:    int(totalPrompts),
		ExecutionsProcessed: int(totalExecutions),
		Period:              p,
		Duration:            time.Since(startTime),
	}, nil
}

// ComputeOrgMetrics aggregates execution log metrics for a single organization
// over the previous calendar month.
func (s *BatchService) ComputeOrgMetrics(ctx context.Context, orgID uuid.UUID) (*OrgMetrics, error) {
	if orgID == uuid.Nil {
		return nil, apperror.NewValidationError(
			fmt.Errorf("organization ID must not be nil"),
			domainName,
		)
	}

	p := period(time.Now())
	start, end, err := monthBounds(p)
	if err != nil {
		return nil, apperror.NewInternalServerError(
			fmt.Errorf("failed to compute month bounds: %w", err),
			domainName,
		)
	}

	row, err := s.q.GetOrgMonthlyMetrics(ctx, db.GetOrgMonthlyMetricsParams{
		OrganizationID: orgID,
		ExecutedAt:     start,
		ExecutedAt_2:   end,
	})
	if err != nil {
		return nil, apperror.NewDatabaseError(
			fmt.Errorf("failed to get monthly metrics for org %s: %w", orgID, err),
			domainName,
		)
	}

	avgScore, _ := strconv.ParseFloat(row.AvgScore, 64)

	return &OrgMetrics{
		OrgID:          orgID,
		AvgLatency:     float64(row.AvgLatencyMs),
		TotalTokens:    row.TotalTokens,
		AvgScore:       avgScore,
		ExecutionCount: row.ExecutionCount,
		ActivePrompts:  row.ActivePrompts,
	}, nil
}

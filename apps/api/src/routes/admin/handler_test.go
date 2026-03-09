package admin

import (
	"api/src/services/batchservice"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	db "utils/db/db"
)

// batchFakeQuerier implements the subset of db.Querier used by BatchService.
type batchFakeQuerier struct {
	db.Querier
	orgs       []db.Organization
	orgErr     error
	metricsMap map[uuid.UUID]db.GetOrgMonthlyMetricsRow
	metricsErr error
}

func (f *batchFakeQuerier) ListAllOrganizations(_ context.Context) ([]db.Organization, error) {
	return f.orgs, f.orgErr
}

func (f *batchFakeQuerier) GetOrgMonthlyMetrics(_ context.Context, arg db.GetOrgMonthlyMetricsParams) (db.GetOrgMonthlyMetricsRow, error) {
	if f.metricsErr != nil {
		return db.GetOrgMonthlyMetricsRow{}, f.metricsErr
	}
	if row, ok := f.metricsMap[arg.OrganizationID]; ok {
		return row, nil
	}
	return db.GetOrgMonthlyMetricsRow{AvgScore: "0.00"}, nil
}

// --- PostAggregate tests ---

func TestPostAggregate(t *testing.T) {
	type expected struct {
		statusCode          int
		orgsProcessed       int
		promptsProcessed    int
		executionsProcessed int
	}

	orgA := uuid.New()

	tests := []struct {
		testName string
		querier  *batchFakeQuerier
		expected expected
	}{
		// 正常系 - successful aggregation
		{
			testName: "returns 200 with aggregation result",
			querier: &batchFakeQuerier{
				orgs: []db.Organization{{ID: orgA, Name: "Test Org", Slug: "test-org"}},
				metricsMap: map[uuid.UUID]db.GetOrgMonthlyMetricsRow{
					orgA: {ExecutionCount: 42, AvgLatencyMs: 150, TotalTokens: 5000, AvgScore: "88.00", ActivePrompts: 3},
				},
			},
			expected: expected{
				statusCode:          http.StatusOK,
				orgsProcessed:       1,
				promptsProcessed:    3,
				executionsProcessed: 42,
			},
		},
		// 正常系 - no organizations
		{
			testName: "returns 200 with zero results when no orgs exist",
			querier: &batchFakeQuerier{
				orgs:       []db.Organization{},
				metricsMap: map[uuid.UUID]db.GetOrgMonthlyMetricsRow{},
			},
			expected: expected{
				statusCode:          http.StatusOK,
				orgsProcessed:       0,
				promptsProcessed:    0,
				executionsProcessed: 0,
			},
		},
		// 異常系 - database error listing orgs
		{
			testName: "returns 500 when listing orgs fails",
			querier: &batchFakeQuerier{
				orgErr: errors.New("db connection lost"),
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
			},
		},
		// 異常系 - database error on metrics
		{
			testName: "returns 500 when metrics query fails",
			querier: &batchFakeQuerier{
				orgs:       []db.Organization{{ID: orgA, Name: "Test Org", Slug: "test-org"}},
				metricsErr: errors.New("timeout"),
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
			},
		},
		// Null/Nil - nil orgs
		{
			testName: "returns 200 when orgs list is nil",
			querier: &batchFakeQuerier{
				orgs:       nil,
				metricsMap: map[uuid.UUID]db.GetOrgMonthlyMetricsRow{},
			},
			expected: expected{
				statusCode:          http.StatusOK,
				orgsProcessed:       0,
				promptsProcessed:    0,
				executionsProcessed: 0,
			},
		},
		// 境界値 - org with zero executions
		{
			testName: "returns 200 with zero orgs processed when all have zero executions",
			querier: &batchFakeQuerier{
				orgs: []db.Organization{{ID: orgA, Name: "Empty Org", Slug: "empty-org"}},
				metricsMap: map[uuid.UUID]db.GetOrgMonthlyMetricsRow{
					orgA: {ExecutionCount: 0, AvgLatencyMs: 0, TotalTokens: 0, AvgScore: "0.00", ActivePrompts: 0},
				},
			},
			expected: expected{
				statusCode:          http.StatusOK,
				orgsProcessed:       0,
				promptsProcessed:    0,
				executionsProcessed: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			svc := batchservice.NewBatchService(tt.querier)
			handler := NewAdminHandler(svc).PostAggregate()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/batch/aggregate", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if diff := cmp.Diff(tt.expected.statusCode, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			if tt.expected.statusCode == http.StatusOK {
				var resp aggregationResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if diff := cmp.Diff(tt.expected.orgsProcessed, resp.OrgsProcessed); diff != "" {
					t.Errorf("OrgsProcessed mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.promptsProcessed, resp.PromptsProcessed); diff != "" {
					t.Errorf("PromptsProcessed mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.executionsProcessed, resp.ExecutionsProcessed); diff != "" {
					t.Errorf("ExecutionsProcessed mismatch (-want +got):\n%s", diff)
				}
				if resp.Period == "" {
					t.Error("expected non-empty period")
				}
				if resp.DurationMs < 0 {
					t.Error("expected non-negative duration")
				}
			}
		})
	}
}

// --- NewAdminHandler tests ---

func TestNewAdminHandler(t *testing.T) {
	tests := []struct {
		testName string
		batch    *batchservice.BatchService
	}{
		// 正常系
		{
			testName: "creates handler with valid batch service",
			batch:    batchservice.NewBatchService(&batchFakeQuerier{}),
		},
		// Null/Nil
		{
			testName: "creates handler with nil batch service",
			batch:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			h := NewAdminHandler(tt.batch)
			if h == nil {
				t.Fatal("expected non-nil handler")
			}
		})
	}
}

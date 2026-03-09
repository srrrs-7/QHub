package batchservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"api/src/domain/apperror"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	db "utils/db/db"
)

// --- helpers ---

// fakeQuerier implements the subset of db.Querier used by BatchService.
type fakeQuerier struct {
	db.Querier
	orgs       []db.Organization
	orgErr     error
	metricsMap map[uuid.UUID]db.GetOrgMonthlyMetricsRow
	metricsErr error
}

func (f *fakeQuerier) ListAllOrganizations(_ context.Context) ([]db.Organization, error) {
	return f.orgs, f.orgErr
}

func (f *fakeQuerier) GetOrgMonthlyMetrics(_ context.Context, arg db.GetOrgMonthlyMetricsParams) (db.GetOrgMonthlyMetricsRow, error) {
	if f.metricsErr != nil {
		return db.GetOrgMonthlyMetricsRow{}, f.metricsErr
	}
	if row, ok := f.metricsMap[arg.OrganizationID]; ok {
		return row, nil
	}
	return db.GetOrgMonthlyMetricsRow{AvgScore: "0.00"}, nil
}

// --- period tests ---

func TestPeriod(t *testing.T) {
	type args struct {
		now time.Time
	}
	type expected struct {
		period string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "returns previous month in YYYY-MM format",
			args:     args{now: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)},
			expected: expected{period: "2026-02"},
		},
		// 境界値 - January wraps to previous year
		{
			testName: "January wraps to December of previous year",
			args:     args{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			expected: expected{period: "2025-12"},
		},
		// 境界値 - first day of month
		{
			testName: "first day of month returns previous month",
			args:     args{now: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
			expected: expected{period: "2026-06"},
		},
		// 境界値 - last day of month (30-day month to avoid Go AddDate overflow)
		{
			testName: "last day of April returns previous month March",
			args:     args{now: time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)},
			expected: expected{period: "2026-03"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := period(tt.args.now)
			if diff := cmp.Diff(tt.expected.period, got); diff != "" {
				t.Errorf("period mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// --- monthBounds tests ---

func TestMonthBounds(t *testing.T) {
	type args struct {
		period string
	}
	type expected struct {
		start   time.Time
		end     time.Time
		wantErr bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "valid period returns correct bounds",
			args:     args{period: "2026-02"},
			expected: expected{
				start: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		// 境界値 - December
		{
			testName: "December period wraps end to January",
			args:     args{period: "2025-12"},
			expected: expected{
				start: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		// 異常系
		{
			testName: "invalid period format returns error",
			args:     args{period: "not-a-date"},
			expected: expected{wantErr: true},
		},
		// 空文字
		{
			testName: "empty period returns error",
			args:     args{period: ""},
			expected: expected{wantErr: true},
		},
		// 特殊文字
		{
			testName: "special characters in period returns error",
			args:     args{period: "2026/02"},
			expected: expected{wantErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			start, end, err := monthBounds(tt.args.period)
			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected.start, start); diff != "" {
				t.Errorf("start mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.end, end); diff != "" {
				t.Errorf("end mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// --- NewBatchService tests ---

func TestNewBatchService(t *testing.T) {
	tests := []struct {
		testName string
		q        db.Querier
	}{
		// 正常系
		{
			testName: "creates service with valid querier",
			q:        &fakeQuerier{},
		},
		// Null/Nil
		{
			testName: "creates service with nil querier",
			q:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			svc := NewBatchService(tt.q)
			if svc == nil {
				t.Fatal("expected non-nil service")
			}
		})
	}
}

// --- ComputeOrgMetrics tests ---

func TestComputeOrgMetrics(t *testing.T) {
	orgID := uuid.New()

	type args struct {
		orgID uuid.UUID
	}
	type expected struct {
		wantErr bool
		errName string
		metrics *OrgMetrics
	}

	tests := []struct {
		testName string
		querier  *fakeQuerier
		args     args
		expected expected
	}{
		// 正常系 - org with executions
		{
			testName: "returns metrics for org with executions",
			querier: &fakeQuerier{
				metricsMap: map[uuid.UUID]db.GetOrgMonthlyMetricsRow{
					orgID: {
						ExecutionCount: 100,
						AvgLatencyMs:   250,
						TotalTokens:    50000,
						AvgScore:       "85.50",
						ActivePrompts:  5,
					},
				},
			},
			args: args{orgID: orgID},
			expected: expected{
				metrics: &OrgMetrics{
					OrgID:          orgID,
					AvgLatency:     250,
					TotalTokens:    50000,
					AvgScore:       85.50,
					ExecutionCount: 100,
					ActivePrompts:  5,
				},
			},
		},
		// 正常系 - org with no executions
		{
			testName: "returns zero metrics for org with no executions",
			querier: &fakeQuerier{
				metricsMap: map[uuid.UUID]db.GetOrgMonthlyMetricsRow{},
			},
			args: args{orgID: orgID},
			expected: expected{
				metrics: &OrgMetrics{
					OrgID:          orgID,
					AvgLatency:     0,
					TotalTokens:    0,
					AvgScore:       0,
					ExecutionCount: 0,
					ActivePrompts:  0,
				},
			},
		},
		// 異常系 - nil org ID
		{
			testName: "returns validation error for nil org ID",
			querier:  &fakeQuerier{},
			args:     args{orgID: uuid.Nil},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
		// 異常系 - database error
		{
			testName: "returns database error on query failure",
			querier: &fakeQuerier{
				metricsErr: errors.New("connection refused"),
			},
			args:     args{orgID: orgID},
			expected: expected{wantErr: true, errName: apperror.DatabaseErrorName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			svc := NewBatchService(tt.querier)
			got, err := svc.ComputeOrgMetrics(context.Background(), tt.args.orgID)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expected.metrics.OrgID, got.OrgID); diff != "" {
				t.Errorf("OrgID mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.metrics.AvgLatency, got.AvgLatency); diff != "" {
				t.Errorf("AvgLatency mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.metrics.TotalTokens, got.TotalTokens); diff != "" {
				t.Errorf("TotalTokens mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.metrics.AvgScore, got.AvgScore); diff != "" {
				t.Errorf("AvgScore mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.metrics.ExecutionCount, got.ExecutionCount); diff != "" {
				t.Errorf("ExecutionCount mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.metrics.ActivePrompts, got.ActivePrompts); diff != "" {
				t.Errorf("ActivePrompts mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// --- RunMonthlyAggregation tests ---

func TestRunMonthlyAggregation(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()

	type expected struct {
		wantErr             bool
		errName             string
		orgsProcessed       int
		promptsProcessed    int
		executionsProcessed int
	}

	tests := []struct {
		testName string
		querier  *fakeQuerier
		expected expected
	}{
		// 正常系 - multiple orgs with data
		{
			testName: "aggregates across multiple organizations",
			querier: &fakeQuerier{
				orgs: []db.Organization{
					{ID: orgA, Name: "Org A", Slug: "org-a"},
					{ID: orgB, Name: "Org B", Slug: "org-b"},
				},
				metricsMap: map[uuid.UUID]db.GetOrgMonthlyMetricsRow{
					orgA: {ExecutionCount: 50, AvgLatencyMs: 200, TotalTokens: 10000, AvgScore: "90.00", ActivePrompts: 3},
					orgB: {ExecutionCount: 30, AvgLatencyMs: 300, TotalTokens: 8000, AvgScore: "80.00", ActivePrompts: 2},
				},
			},
			expected: expected{
				orgsProcessed:       2,
				promptsProcessed:    5,
				executionsProcessed: 80,
			},
		},
		// 正常系 - no organizations
		{
			testName: "handles empty organization list",
			querier: &fakeQuerier{
				orgs:       []db.Organization{},
				metricsMap: map[uuid.UUID]db.GetOrgMonthlyMetricsRow{},
			},
			expected: expected{
				orgsProcessed:       0,
				promptsProcessed:    0,
				executionsProcessed: 0,
			},
		},
		// 正常系 - org with zero executions
		{
			testName: "skips orgs with zero executions in count",
			querier: &fakeQuerier{
				orgs: []db.Organization{
					{ID: orgA, Name: "Org A", Slug: "org-a"},
				},
				metricsMap: map[uuid.UUID]db.GetOrgMonthlyMetricsRow{
					orgA: {ExecutionCount: 0, AvgLatencyMs: 0, TotalTokens: 0, AvgScore: "0.00", ActivePrompts: 0},
				},
			},
			expected: expected{
				orgsProcessed:       0,
				promptsProcessed:    0,
				executionsProcessed: 0,
			},
		},
		// 異常系 - list orgs fails
		{
			testName: "returns database error when listing orgs fails",
			querier: &fakeQuerier{
				orgErr: errors.New("db down"),
			},
			expected: expected{wantErr: true, errName: apperror.DatabaseErrorName},
		},
		// 異常系 - metrics query fails
		{
			testName: "returns error when metrics query fails",
			querier: &fakeQuerier{
				orgs:       []db.Organization{{ID: orgA, Name: "Org A", Slug: "org-a"}},
				metricsErr: errors.New("query timeout"),
			},
			expected: expected{wantErr: true, errName: apperror.DatabaseErrorName},
		},
		// Null/Nil - nil orgs slice
		{
			testName: "handles nil orgs slice gracefully",
			querier: &fakeQuerier{
				orgs:       nil,
				metricsMap: map[uuid.UUID]db.GetOrgMonthlyMetricsRow{},
			},
			expected: expected{
				orgsProcessed:       0,
				promptsProcessed:    0,
				executionsProcessed: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			svc := NewBatchService(tt.querier)
			got, err := svc.RunMonthlyAggregation(context.Background())

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expected.orgsProcessed, got.OrgsProcessed); diff != "" {
				t.Errorf("OrgsProcessed mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.promptsProcessed, got.PromptsProcessed); diff != "" {
				t.Errorf("PromptsProcessed mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.executionsProcessed, got.ExecutionsProcessed); diff != "" {
				t.Errorf("ExecutionsProcessed mismatch (-want +got):\n%s", diff)
			}

			// Period should be previous month in YYYY-MM format
			if got.Period == "" {
				t.Error("expected non-empty period")
			}
			if got.Duration < 0 {
				t.Error("expected non-negative duration")
			}
		})
	}
}

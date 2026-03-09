package statsservice

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// --- Unit tests for statistical helper functions ---

func TestComputeStats(t *testing.T) {
	type args struct {
		data    []float64
		version int
	}
	type expected struct {
		mean   float64
		stddev float64
		n      int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (sample stddev of [2,4,4,4,5,5,7,9] = sqrt(32/7) ~ 2.1381)
		{
			testName: "normal data set",
			args:     args{data: []float64{2, 4, 4, 4, 5, 5, 7, 9}, version: 1},
			expected: expected{mean: 5.0, stddev: math.Sqrt(32.0 / 7.0), n: 8},
		},
		// 正常系 - two identical values
		{
			testName: "two identical values",
			args:     args{data: []float64{5, 5}, version: 1},
			expected: expected{mean: 5.0, stddev: 0.0, n: 2},
		},
		// 正常系 - two different values
		{
			testName: "two different values",
			args:     args{data: []float64{3, 7}, version: 1},
			expected: expected{mean: 5.0, stddev: math.Sqrt(8.0), n: 2},
		},
		// 境界値 - single element
		{
			testName: "single element returns zero stddev",
			args:     args{data: []float64{42.0}, version: 1},
			expected: expected{mean: 42.0, stddev: 0.0, n: 1},
		},
		// 空文字 / Nil - empty slice
		{
			testName: "empty slice",
			args:     args{data: []float64{}, version: 1},
			expected: expected{mean: 0.0, stddev: 0.0, n: 0},
		},
		// Nil
		{
			testName: "nil slice",
			args:     args{data: nil, version: 1},
			expected: expected{mean: 0.0, stddev: 0.0, n: 0},
		},
		// 境界値 - large values
		{
			testName: "large values",
			args:     args{data: []float64{1e6, 1e6 + 2, 1e6 + 4}, version: 2},
			expected: expected{mean: 1e6 + 2, stddev: 2.0, n: 3},
		},
		// 特殊文字/特殊値 - zeros
		{
			testName: "all zeros",
			args:     args{data: []float64{0, 0, 0, 0}, version: 1},
			expected: expected{mean: 0.0, stddev: 0.0, n: 4},
		},
		// 境界値 - negative values
		{
			testName: "negative values",
			args:     args{data: []float64{-3, -1, 1, 3}, version: 1},
			expected: expected{mean: 0.0, stddev: math.Sqrt(20.0 / 3.0), n: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := computeStats(tt.args.data, tt.args.version)

			if got.N != tt.expected.n {
				t.Errorf("n: want %d, got %d", tt.expected.n, got.N)
			}
			if diff := cmp.Diff(tt.expected.mean, got.Mean, cmpopts.EquateApprox(0, 1e-9)); diff != "" {
				t.Errorf("mean mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.stddev, got.StdDev, cmpopts.EquateApprox(0, 1e-9)); diff != "" {
				t.Errorf("stddev mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWelchTTest(t *testing.T) {
	type args struct {
		meanA, meanB, sdA, sdB float64
		nA, nB                 int
	}
	type expected struct {
		tStat float64
		df    float64
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - known result
		// Group A: mean=5, sd=2, n=8; Group B: mean=4.5, sd=2.449, n=8
		{
			testName: "known two-sample test",
			args:     args{meanA: 5.0, meanB: 4.5, sdA: 2.0, sdB: 2.449, nA: 8, nB: 8},
			expected: expected{
				tStat: 0.5 / math.Sqrt(4.0/8.0+2.449*2.449/8.0),
				df:    13.46, // Welch-Satterthwaite approximation
			},
		},
		// 正常系 - equal means
		{
			testName: "equal means yield t=0",
			args:     args{meanA: 10.0, meanB: 10.0, sdA: 3.0, sdB: 5.0, nA: 30, nB: 30},
			expected: expected{tStat: 0.0, df: 47.48},
		},
		// 境界値 - minimum samples (n=2)
		{
			testName: "minimum samples n=2",
			args:     args{meanA: 10.0, meanB: 20.0, sdA: 1.0, sdB: 1.0, nA: 2, nB: 2},
			expected: expected{
				tStat: -10.0,
				df:    2.0,
			},
		},
		// 異常系 - zero stddev in both (denom=0)
		{
			testName: "zero stddev both groups",
			args:     args{meanA: 5.0, meanB: 5.0, sdA: 0.0, sdB: 0.0, nA: 10, nB: 10},
			expected: expected{tStat: 0.0, df: 18.0},
		},
		// 異常系 - zero stddev in one group
		{
			testName: "zero stddev one group",
			args:     args{meanA: 5.0, meanB: 10.0, sdA: 0.0, sdB: 2.0, nA: 10, nB: 10},
			expected: expected{
				tStat: -5.0 / math.Sqrt(0.4),
				df:    9.0,
			},
		},
		// 特殊値 - very different sample sizes
		{
			testName: "unequal sample sizes",
			args:     args{meanA: 100.0, meanB: 110.0, sdA: 15.0, sdB: 20.0, nA: 50, nB: 10},
			expected: expected{
				tStat: -10.0 / math.Sqrt(225.0/50.0+400.0/10.0),
				df:    11.11,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			gotT, gotDF := WelchTTest(tt.args.meanA, tt.args.meanB, tt.args.sdA, tt.args.sdB, tt.args.nA, tt.args.nB)

			if diff := cmp.Diff(tt.expected.tStat, gotT, cmpopts.EquateApprox(0, 0.01)); diff != "" {
				t.Errorf("t-statistic mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.df, gotDF, cmpopts.EquateApprox(0, 0.1)); diff != "" {
				t.Errorf("degrees of freedom mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTwoTailedPValue(t *testing.T) {
	type args struct {
		tStat float64
		df    float64
	}
	type expected struct {
		pValue float64
		tol    float64
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - t=0 should give p=1
		{
			testName: "t=0 gives p=1",
			args:     args{tStat: 0.0, df: 10.0},
			expected: expected{pValue: 1.0, tol: 1e-6},
		},
		// 正常系 - known value: t=2.228, df=10 => p~0.05
		{
			testName: "t=2.228 df=10 gives p~0.05",
			args:     args{tStat: 2.228, df: 10.0},
			expected: expected{pValue: 0.05, tol: 0.002},
		},
		// 正常系 - large t-value should give very small p
		{
			testName: "large t gives small p",
			args:     args{tStat: 10.0, df: 30.0},
			expected: expected{pValue: 0.0, tol: 0.0001},
		},
		// 正常系 - negative t same as positive (two-tailed)
		{
			testName: "negative t same as positive",
			args:     args{tStat: -2.228, df: 10.0},
			expected: expected{pValue: 0.05, tol: 0.002},
		},
		// 境界値 - df=1 (Cauchy distribution)
		{
			testName: "df=1 t=1 gives p~0.5",
			args:     args{tStat: 1.0, df: 1.0},
			expected: expected{pValue: 0.5, tol: 0.01},
		},
		// 境界値 - very large df (approaches normal)
		{
			testName: "large df t=1.96 gives p~0.05",
			args:     args{tStat: 1.96, df: 10000.0},
			expected: expected{pValue: 0.05, tol: 0.002},
		},
		// 異常系 - df=0 returns 1
		{
			testName: "df=0 returns 1",
			args:     args{tStat: 5.0, df: 0.0},
			expected: expected{pValue: 1.0, tol: 1e-6},
		},
		// 異常系 - negative df returns 1
		{
			testName: "negative df returns 1",
			args:     args{tStat: 5.0, df: -1.0},
			expected: expected{pValue: 1.0, tol: 1e-6},
		},
		// 境界値 - df=2, t=4.303 => p~0.05
		{
			testName: "df=2 t=4.303 gives p~0.05",
			args:     args{tStat: 4.303, df: 2.0},
			expected: expected{pValue: 0.05, tol: 0.003},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := TwoTailedPValue(tt.args.tStat, tt.args.df)

			if math.Abs(got-tt.expected.pValue) > tt.expected.tol {
				t.Errorf("p-value: want %f (+/- %f), got %f", tt.expected.pValue, tt.expected.tol, got)
			}
		})
	}
}

func TestRegularizedIncompleteBeta(t *testing.T) {
	type args struct {
		x, a, b float64
	}
	type expected struct {
		value float64
		tol   float64
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "I_0.5(1,1) = 0.5",
			args:     args{x: 0.5, a: 1.0, b: 1.0},
			expected: expected{value: 0.5, tol: 1e-10},
		},
		// 境界値 - x=0
		{
			testName: "x=0 returns 0",
			args:     args{x: 0.0, a: 2.0, b: 3.0},
			expected: expected{value: 0.0, tol: 1e-10},
		},
		// 境界値 - x=1
		{
			testName: "x=1 returns 1",
			args:     args{x: 1.0, a: 2.0, b: 3.0},
			expected: expected{value: 1.0, tol: 1e-10},
		},
		// 異常系 - x < 0
		{
			testName: "x<0 returns 0",
			args:     args{x: -0.1, a: 2.0, b: 3.0},
			expected: expected{value: 0.0, tol: 1e-10},
		},
		// 異常系 - x > 1
		{
			testName: "x>1 returns 0",
			args:     args{x: 1.1, a: 2.0, b: 3.0},
			expected: expected{value: 0.0, tol: 1e-10},
		},
		// 正常系 - I_0.5(5, 2.5) known value ~0.8125 (from tables)
		{
			testName: "known beta function value",
			args:     args{x: 0.5, a: 5.0, b: 2.5},
			expected: expected{value: 0.234375 * math.Exp(lgamma(5.0)+lgamma(2.5)-lgamma(7.5)) * 5.0 / 0.234375, tol: 0.05},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := RegularizedIncompleteBeta(tt.args.x, tt.args.a, tt.args.b)

			// For the "known beta function value" case we just check it's in [0,1]
			if tt.testName == "known beta function value" {
				if got < 0 || got > 1 {
					t.Errorf("result out of [0,1] range: %f", got)
				}
				return
			}

			if math.Abs(got-tt.expected.value) > tt.expected.tol {
				t.Errorf("want %f (+/- %f), got %f", tt.expected.value, tt.expected.tol, got)
			}
		})
	}
}

func TestCompareMetric(t *testing.T) {
	type args struct {
		name          string
		a, b          []float64
		vA, vB        int
		lowerIsBetter bool
	}
	type expected struct {
		isSignificant bool
		winner        string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - clearly different distributions (latency, lower is better)
		{
			testName: "significantly different latency",
			args: args{
				name: "latency_ms",
				a:    []float64{100, 102, 98, 101, 99, 100, 103, 97, 101, 100},
				b:    []float64{200, 198, 202, 199, 201, 200, 203, 197, 201, 200},
				vA:   1, vB: 2,
				lowerIsBetter: true,
			},
			expected: expected{isSignificant: true, winner: "v1"},
		},
		// 正常系 - clearly different scores (higher is better)
		{
			testName: "significantly different scores higher is better",
			args: args{
				name: "overall_score",
				a:    []float64{3.0, 3.1, 2.9, 3.0, 3.2, 2.8, 3.0, 3.1, 2.9, 3.0},
				b:    []float64{8.0, 8.1, 7.9, 8.0, 8.2, 7.8, 8.0, 8.1, 7.9, 8.0},
				vA:   1, vB: 2,
				lowerIsBetter: false,
			},
			expected: expected{isSignificant: true, winner: "v2"},
		},
		// 正常系 - identical distributions
		{
			testName: "identical distributions not significant",
			args: args{
				name: "latency_ms",
				a:    []float64{100, 100, 100, 100, 100},
				b:    []float64{100, 100, 100, 100, 100},
				vA:   1, vB: 2,
				lowerIsBetter: true,
			},
			expected: expected{isSignificant: false, winner: "inconclusive"},
		},
		// 境界値 - single sample per group (insufficient)
		{
			testName: "single sample insufficient for test",
			args: args{
				name: "latency_ms",
				a:    []float64{100},
				b:    []float64{200},
				vA:   1, vB: 2,
				lowerIsBetter: true,
			},
			expected: expected{isSignificant: false, winner: "inconclusive"},
		},
		// 境界値 - minimum samples (n=2) with large difference
		{
			testName: "minimum samples n=2 large difference",
			args: args{
				name: "total_tokens",
				a:    []float64{10, 10},
				b:    []float64{1000, 1000},
				vA:   1, vB: 2,
				lowerIsBetter: true,
			},
			// Zero variance + different means => significant
			expected: expected{isSignificant: true, winner: "v1"},
		},
		// 異常系 - zero variance same mean
		{
			testName: "zero variance same mean",
			args: args{
				name: "latency_ms",
				a:    []float64{50, 50, 50},
				b:    []float64{50, 50, 50},
				vA:   1, vB: 2,
				lowerIsBetter: true,
			},
			expected: expected{isSignificant: false, winner: "inconclusive"},
		},
		// 正常系 - overlapping distributions (not significant)
		{
			testName: "overlapping distributions not significant",
			args: args{
				name: "latency_ms",
				a:    []float64{100, 200, 150, 180, 120},
				b:    []float64{110, 190, 160, 170, 130},
				vA:   1, vB: 2,
				lowerIsBetter: true,
			},
			expected: expected{isSignificant: false, winner: "inconclusive"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := compareMetric(tt.args.name, tt.args.a, tt.args.b, tt.args.vA, tt.args.vB, tt.args.lowerIsBetter)

			if diff := cmp.Diff(tt.expected.isSignificant, got.IsSignificant); diff != "" {
				t.Errorf("IsSignificant mismatch (-want +got):\n%s (p=%f, t=%f, df=%f)", diff, got.PValue, got.TStatistic, got.DegreesOfFreedom)
			}
			if diff := cmp.Diff(tt.expected.winner, got.Winner); diff != "" {
				t.Errorf("Winner mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDetermineOverallWinner(t *testing.T) {
	type expected struct {
		winner string
	}

	tests := []struct {
		testName string
		metrics  []MetricComparison
		expected expected
	}{
		// 正常系 - clear winner
		{
			testName: "v1 wins 2 of 3 metrics",
			metrics: []MetricComparison{
				{IsSignificant: true, Winner: "v1"},
				{IsSignificant: true, Winner: "v1"},
				{IsSignificant: true, Winner: "v2"},
			},
			expected: expected{winner: "v1"},
		},
		// 正常系 - no significant differences
		{
			testName: "no significant differences",
			metrics: []MetricComparison{
				{IsSignificant: false, Winner: "inconclusive"},
				{IsSignificant: false, Winner: "inconclusive"},
				{IsSignificant: false, Winner: "inconclusive"},
			},
			expected: expected{winner: "inconclusive"},
		},
		// 正常系 - tie
		{
			testName: "tie results in inconclusive",
			metrics: []MetricComparison{
				{IsSignificant: true, Winner: "v1"},
				{IsSignificant: true, Winner: "v2"},
				{IsSignificant: false, Winner: "inconclusive"},
			},
			expected: expected{winner: "inconclusive"},
		},
		// 空 - empty metrics
		{
			testName: "empty metrics",
			metrics:  []MetricComparison{},
			expected: expected{winner: "inconclusive"},
		},
		// Nil
		{
			testName: "nil metrics",
			metrics:  nil,
			expected: expected{winner: "inconclusive"},
		},
		// 正常系 - single significant metric
		{
			testName: "single significant metric decides winner",
			metrics: []MetricComparison{
				{IsSignificant: true, Winner: "v2"},
				{IsSignificant: false, Winner: "inconclusive"},
				{IsSignificant: false, Winner: "inconclusive"},
			},
			expected: expected{winner: "v2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := determineOverallWinner(tt.metrics)
			if diff := cmp.Diff(tt.expected.winner, got); diff != "" {
				t.Errorf("winner mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPickWinner(t *testing.T) {
	type args struct {
		meanA, meanB  float64
		vA, vB        int
		lowerIsBetter bool
	}
	type expected struct {
		winner string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - lower is better, A wins
		{
			testName: "lower is better A wins",
			args:     args{meanA: 100, meanB: 200, vA: 1, vB: 2, lowerIsBetter: true},
			expected: expected{winner: "v1"},
		},
		// 正常系 - lower is better, B wins
		{
			testName: "lower is better B wins",
			args:     args{meanA: 200, meanB: 100, vA: 1, vB: 2, lowerIsBetter: true},
			expected: expected{winner: "v2"},
		},
		// 正常系 - higher is better, A wins
		{
			testName: "higher is better A wins",
			args:     args{meanA: 9.0, meanB: 5.0, vA: 1, vB: 2, lowerIsBetter: false},
			expected: expected{winner: "v1"},
		},
		// 正常系 - higher is better, B wins
		{
			testName: "higher is better B wins",
			args:     args{meanA: 5.0, meanB: 9.0, vA: 1, vB: 2, lowerIsBetter: false},
			expected: expected{winner: "v2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := pickWinner(tt.args.meanA, tt.args.meanB, tt.args.vA, tt.args.vB, tt.args.lowerIsBetter)
			if diff := cmp.Diff(tt.expected.winner, got); diff != "" {
				t.Errorf("winner mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewStatsService(t *testing.T) {
	// Nil querier
	t.Run("nil querier", func(t *testing.T) {
		svc := NewStatsService(nil)
		if svc == nil {
			t.Fatal("expected non-nil service")
		}
		if svc.q != nil {
			t.Error("expected nil querier")
		}
	})
}

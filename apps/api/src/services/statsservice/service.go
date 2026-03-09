// Package statsservice performs statistical significance testing for A/B
// comparison of prompt version metrics using Welch's t-test.
//
// It fetches per-execution metrics (latency, total tokens, overall score) for
// two prompt versions, computes descriptive statistics, and performs a two-sample
// Welch's t-test to determine whether the observed differences are statistically
// significant at the p < 0.05 level.
package statsservice

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"api/src/domain/apperror"

	"github.com/google/uuid"

	db "utils/db/db"
)

// StatsService compares metrics between prompt versions using Welch's t-test.
type StatsService struct {
	q db.Querier
}

// NewStatsService creates a new StatsService with the given querier.
func NewStatsService(q db.Querier) *StatsService {
	return &StatsService{q: q}
}

// VersionStats holds descriptive statistics for a single version's metric.
type VersionStats struct {
	VersionNumber int     `json:"version_number"`
	Mean          float64 `json:"mean"`
	StdDev        float64 `json:"std_dev"`
	N             int     `json:"n"`
}

// MetricComparison holds the t-test results for a single metric.
type MetricComparison struct {
	MetricName       string       `json:"metric_name"`
	VersionA         VersionStats `json:"version_a"`
	VersionB         VersionStats `json:"version_b"`
	TStatistic       float64      `json:"t_statistic"`
	DegreesOfFreedom float64      `json:"degrees_of_freedom"`
	PValue           float64      `json:"p_value"`
	IsSignificant    bool         `json:"is_significant"`
	Winner           string       `json:"winner"`
}

// ComparisonResult holds the full A/B comparison between two versions.
type ComparisonResult struct {
	PromptID       string             `json:"prompt_id"`
	VersionANumber int                `json:"version_a"`
	VersionBNumber int                `json:"version_b"`
	Metrics        []MetricComparison `json:"metrics"`
	OverallWinner  string             `json:"overall_winner"`
}

// CompareVersions fetches metrics for two prompt versions and performs Welch's
// t-test on latency_ms, total_tokens, and overall_score.
func (s *StatsService) CompareVersions(ctx context.Context, promptID uuid.UUID, versionA, versionB int) (*ComparisonResult, error) {
	if versionA <= 0 || versionB <= 0 {
		return nil, apperror.NewValidationError(
			fmt.Errorf("version numbers must be positive: got %d and %d", versionA, versionB),
			"StatsService",
		)
	}
	if versionA == versionB {
		return nil, apperror.NewValidationError(
			fmt.Errorf("version numbers must differ: both are %d", versionA),
			"StatsService",
		)
	}

	metricsA, err := s.q.GetVersionMetrics(ctx, db.GetVersionMetricsParams{
		PromptID:      promptID,
		VersionNumber: int32(versionA),
	})
	if err != nil {
		return nil, apperror.NewDatabaseError(
			fmt.Errorf("failed to get metrics for version %d: %w", versionA, err),
			"StatsService",
		)
	}
	if len(metricsA) == 0 {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("no execution data for version %d", versionA),
			"StatsService",
		)
	}

	metricsB, err := s.q.GetVersionMetrics(ctx, db.GetVersionMetricsParams{
		PromptID:      promptID,
		VersionNumber: int32(versionB),
	})
	if err != nil {
		return nil, apperror.NewDatabaseError(
			fmt.Errorf("failed to get metrics for version %d: %w", versionB, err),
			"StatsService",
		)
	}
	if len(metricsB) == 0 {
		return nil, apperror.NewNotFoundError(
			fmt.Errorf("no execution data for version %d", versionB),
			"StatsService",
		)
	}

	// Extract metric arrays
	latA, tokA, scrA := extractMetrics(metricsA)
	latB, tokB, scrB := extractMetrics(metricsB)

	comparisons := []MetricComparison{
		compareMetric("latency_ms", latA, latB, versionA, versionB, true),
		compareMetric("total_tokens", tokA, tokB, versionA, versionB, true),
		compareMetric("overall_score", scrA, scrB, versionA, versionB, false),
	}

	overallWinner := determineOverallWinner(comparisons)

	return &ComparisonResult{
		PromptID:       promptID.String(),
		VersionANumber: versionA,
		VersionBNumber: versionB,
		Metrics:        comparisons,
		OverallWinner:  overallWinner,
	}, nil
}

// extractMetrics converts DB rows into separate float64 slices per metric.
func extractMetrics(rows []db.GetVersionMetricsRow) (latency, tokens, scores []float64) {
	latency = make([]float64, len(rows))
	tokens = make([]float64, len(rows))
	scores = make([]float64, len(rows))
	for i, r := range rows {
		latency[i] = float64(r.LatencyMs)
		tokens[i] = float64(r.TotalTokens)
		s, _ := strconv.ParseFloat(r.OverallScore, 64)
		scores[i] = s
	}
	return
}

// compareMetric runs a Welch's t-test for a single metric between two samples.
// lowerIsBetter indicates whether lower values are preferred (e.g., latency).
func compareMetric(name string, a, b []float64, vA, vB int, lowerIsBetter bool) MetricComparison {
	statsA := computeStats(a, vA)
	statsB := computeStats(b, vB)

	mc := MetricComparison{
		MetricName: name,
		VersionA:   statsA,
		VersionB:   statsB,
		Winner:     "inconclusive",
	}

	// Need at least 2 samples per group for variance estimation.
	if statsA.N < 2 || statsB.N < 2 {
		mc.TStatistic = 0
		mc.DegreesOfFreedom = 0
		mc.PValue = 1.0
		return mc
	}

	// Zero variance in both groups means identical samples.
	if statsA.StdDev == 0 && statsB.StdDev == 0 {
		mc.TStatistic = 0
		mc.DegreesOfFreedom = float64(statsA.N + statsB.N - 2)
		if statsA.Mean == statsB.Mean {
			mc.PValue = 1.0
		} else {
			mc.PValue = 0.0
			mc.IsSignificant = true
			mc.Winner = pickWinner(statsA.Mean, statsB.Mean, vA, vB, lowerIsBetter)
		}
		return mc
	}

	t, df := WelchTTest(statsA.Mean, statsB.Mean, statsA.StdDev, statsB.StdDev, statsA.N, statsB.N)
	p := TwoTailedPValue(t, df)

	mc.TStatistic = t
	mc.DegreesOfFreedom = df
	mc.PValue = p
	mc.IsSignificant = p < 0.05

	if mc.IsSignificant {
		mc.Winner = pickWinner(statsA.Mean, statsB.Mean, vA, vB, lowerIsBetter)
	}

	return mc
}

// pickWinner returns the version label with the "better" mean.
func pickWinner(meanA, meanB float64, vA, vB int, lowerIsBetter bool) string {
	if lowerIsBetter {
		if meanA < meanB {
			return fmt.Sprintf("v%d", vA)
		}
		return fmt.Sprintf("v%d", vB)
	}
	// higher is better (e.g., scores)
	if meanA > meanB {
		return fmt.Sprintf("v%d", vA)
	}
	return fmt.Sprintf("v%d", vB)
}

// determineOverallWinner picks the version that wins the most metrics.
func determineOverallWinner(metrics []MetricComparison) string {
	wins := make(map[string]int)
	for _, m := range metrics {
		if m.IsSignificant && m.Winner != "inconclusive" {
			wins[m.Winner]++
		}
	}

	if len(wins) == 0 {
		return "inconclusive"
	}

	best := ""
	bestCount := 0
	for v, count := range wins {
		if count > bestCount {
			best = v
			bestCount = count
		}
	}

	// Tie check: if two versions have equal wins, inconclusive.
	for v, count := range wins {
		if v != best && count == bestCount {
			return "inconclusive"
		}
	}

	return best
}

// computeStats computes mean and sample standard deviation for a float64 slice.
func computeStats(data []float64, version int) VersionStats {
	n := len(data)
	if n == 0 {
		return VersionStats{VersionNumber: version}
	}

	sum := 0.0
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(n)

	if n < 2 {
		return VersionStats{VersionNumber: version, Mean: mean, N: n}
	}

	variance := 0.0
	for _, v := range data {
		d := v - mean
		variance += d * d
	}
	variance /= float64(n - 1) // sample variance (Bessel's correction)

	return VersionStats{
		VersionNumber: version,
		Mean:          mean,
		StdDev:        math.Sqrt(variance),
		N:             n,
	}
}

// WelchTTest computes the t-statistic and Welch-Satterthwaite degrees of
// freedom for a two-sample test assuming unequal variances.
func WelchTTest(meanA, meanB, sdA, sdB float64, nA, nB int) (tStat, df float64) {
	varA := sdA * sdA / float64(nA)
	varB := sdB * sdB / float64(nB)

	denom := math.Sqrt(varA + varB)
	if denom == 0 {
		return 0, float64(nA + nB - 2)
	}

	tStat = (meanA - meanB) / denom

	// Welch-Satterthwaite degrees of freedom
	num := (varA + varB) * (varA + varB)
	denomDF := (varA*varA)/float64(nA-1) + (varB*varB)/float64(nB-1)
	if denomDF == 0 {
		df = float64(nA + nB - 2)
	} else {
		df = num / denomDF
	}

	return tStat, df
}

// TwoTailedPValue computes the two-tailed p-value from a t-statistic and
// degrees of freedom using the regularised incomplete beta function.
func TwoTailedPValue(t, df float64) float64 {
	if df <= 0 {
		return 1.0
	}
	x := df / (df + t*t)
	p := RegularizedIncompleteBeta(x, df/2.0, 0.5)
	return p
}

// RegularizedIncompleteBeta computes I_x(a, b) = B(x; a, b) / B(a, b) using
// a continued-fraction expansion (Lentz's algorithm).
func RegularizedIncompleteBeta(x, a, b float64) float64 {
	if x < 0 || x > 1 {
		return 0
	}
	if x == 0 {
		return 0
	}
	if x == 1 {
		return 1
	}

	// Use symmetry relation when x > (a+1)/(a+b+2) for better convergence.
	if x > (a+1)/(a+b+2) {
		return 1.0 - RegularizedIncompleteBeta(1-x, b, a)
	}

	lnBeta := lgamma(a) + lgamma(b) - lgamma(a+b)
	front := math.Exp(math.Log(x)*a+math.Log(1-x)*b-lnBeta) / a

	// Lentz's continued fraction
	return front * betaCF(x, a, b)
}

// betaCF evaluates the continued fraction for the incomplete beta function.
func betaCF(x, a, b float64) float64 {
	const maxIter = 200
	const epsilon = 1e-14
	const tiny = 1e-30

	// Modified Lentz's algorithm
	c := 1.0
	d := 1.0 - (a+b)*x/(a+1)
	if math.Abs(d) < tiny {
		d = tiny
	}
	d = 1.0 / d
	f := d

	for m := 1; m <= maxIter; m++ {
		mf := float64(m)

		// Even step: d_{2m}
		num := mf * (b - mf) * x / ((a + 2*mf - 1) * (a + 2*mf))
		d = 1.0 + num*d
		if math.Abs(d) < tiny {
			d = tiny
		}
		c = 1.0 + num/c
		if math.Abs(c) < tiny {
			c = tiny
		}
		d = 1.0 / d
		f *= d * c

		// Odd step: d_{2m+1}
		num = -(a + mf) * (a + b + mf) * x / ((a + 2*mf) * (a + 2*mf + 1))
		d = 1.0 + num*d
		if math.Abs(d) < tiny {
			d = tiny
		}
		c = 1.0 + num/c
		if math.Abs(c) < tiny {
			c = tiny
		}
		d = 1.0 / d
		delta := d * c
		f *= delta

		if math.Abs(delta-1.0) < epsilon {
			break
		}
	}

	return f
}

// lgamma wraps math.Lgamma, discarding the sign.
func lgamma(x float64) float64 {
	v, _ := math.Lgamma(x)
	return v
}

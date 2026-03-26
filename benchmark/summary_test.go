package benchmark

import (
	"math"
	"testing"
)

func TestPercentile_Empty(t *testing.T) {
	if got := percentile(nil, 0.50); got != 0 {
		t.Errorf("expected 0 for empty slice, got %f", got)
	}
}

func TestPercentile_SingleValue(t *testing.T) {
	if got := percentile([]float64{42.0}, 0.95); got != 42.0 {
		t.Errorf("expected 42.0, got %f", got)
	}
}

func TestPercentile_OddCount(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}
	p50 := percentile(values, 0.50)
	if p50 != 30.0 {
		t.Errorf("P50 expected 30.0, got %f", p50)
	}
	p95 := percentile(values, 0.95)
	expected := 46.0 // 0.95 * 4 = 3.8 → 40*0.2 + 50*0.8 = 48, actually: 40*(4-3.8) + 50*(3.8-3) = 40*0.2 + 50*0.8 = 48
	// Recalculate: k = 0.95 * 4 = 3.8, f=3, c=4, sorted[3]*(4-3.8) + sorted[4]*(3.8-3) = 40*0.2 + 50*0.8 = 48
	expected = 48.0
	if math.Abs(p95-expected) > 0.001 {
		t.Errorf("P95 expected %f, got %f", expected, p95)
	}
}

func TestPercentile_EvenCount(t *testing.T) {
	values := []float64{10, 20, 30, 40}
	p50 := percentile(values, 0.50)
	// k = 0.50 * 3 = 1.5, f=1, c=2, 20*0.5 + 30*0.5 = 25
	if p50 != 25.0 {
		t.Errorf("P50 expected 25.0, got %f", p50)
	}
}

func TestPercentile_DoesNotMutateInput(t *testing.T) {
	values := []float64{50, 10, 30}
	percentile(values, 0.50)
	if values[0] != 50 || values[1] != 10 || values[2] != 30 {
		t.Errorf("percentile mutated input slice: %v", values)
	}
}

func TestComputeSummary_BasicAggregates(t *testing.T) {
	results := []Result{
		{Model: "model-a", Status: ResultStatusSuccess, Metrics: ResultMetrics{
			TTFB: 100, TotalLatency: 200, TokensPerSecond: 50, EstimatedCost: 0.01, CompletionTokens: 100,
		}},
		{Model: "model-a", Status: ResultStatusSuccess, Metrics: ResultMetrics{
			TTFB: 200, TotalLatency: 400, TokensPerSecond: 30, EstimatedCost: 0.02, CompletionTokens: 120,
		}},
		{Model: "model-b", Status: ResultStatusSuccess, Metrics: ResultMetrics{
			TTFB: 150, TotalLatency: 300, TokensPerSecond: 40, EstimatedCost: 0.015, CompletionTokens: 110,
		}},
	}

	summary := computeSummary(results, nil)
	if summary == nil {
		t.Fatal("summary is nil")
	}
	if len(summary.Models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(summary.Models))
	}

	// Models are sorted alphabetically
	a := summary.Models[0]
	if a.Model != "model-a" {
		t.Errorf("expected model-a, got %s", a.Model)
	}
	if a.AvgTTFB != 150.0 {
		t.Errorf("AvgTTFB expected 150.0, got %f", a.AvgTTFB)
	}
	if a.AvgLatency != 300.0 {
		t.Errorf("AvgLatency expected 300.0, got %f", a.AvgLatency)
	}
	if a.AvgTokensPerSecond != 40.0 {
		t.Errorf("AvgTokensPerSecond expected 40.0, got %f", a.AvgTokensPerSecond)
	}
	if a.TotalCost != 0.03 {
		t.Errorf("TotalCost expected 0.03, got %f", a.TotalCost)
	}
	if a.SuccessRate != 1.0 {
		t.Errorf("SuccessRate expected 1.0, got %f", a.SuccessRate)
	}

	b := summary.Models[1]
	if b.Model != "model-b" {
		t.Errorf("expected model-b, got %s", b.Model)
	}
	if b.SuccessRate != 1.0 {
		t.Errorf("SuccessRate expected 1.0, got %f", b.SuccessRate)
	}
}

func TestComputeSummary_WithFailures(t *testing.T) {
	results := []Result{
		{Model: "model-a", Status: ResultStatusSuccess, Metrics: ResultMetrics{
			TTFB: 100, TotalLatency: 200, TokensPerSecond: 50,
		}},
		{Model: "model-a", Status: ResultStatusError},
		{Model: "model-a", Status: ResultStatusTimeout},
	}

	summary := computeSummary(results, nil)
	a := summary.Models[0]
	if math.Abs(a.SuccessRate-1.0/3.0) > 0.001 {
		t.Errorf("SuccessRate expected ~0.333, got %f", a.SuccessRate)
	}
	if a.AvgLatency != 200.0 {
		t.Errorf("AvgLatency should only consider successful results, got %f", a.AvgLatency)
	}
}

func TestComputeSummary_AllFailures(t *testing.T) {
	results := []Result{
		{Model: "model-a", Status: ResultStatusError},
		{Model: "model-a", Status: ResultStatusTimeout},
	}

	summary := computeSummary(results, nil)
	if len(summary.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(summary.Models))
	}
	a := summary.Models[0]
	if a.SuccessRate != 0 {
		t.Errorf("SuccessRate expected 0, got %f", a.SuccessRate)
	}
	if a.AvgLatency != 0 {
		t.Errorf("AvgLatency expected 0 for all failures, got %f", a.AvgLatency)
	}
}

func TestComputeSummary_WithJudgeScores(t *testing.T) {
	results := []Result{
		{
			Model: "model-a", Status: ResultStatusSuccess,
			Metrics: ResultMetrics{TTFB: 100, TotalLatency: 200},
			JudgeScores: []JudgeScore{
				{Criterion: "accuracy", Score: 8.0, Explanation: "good"},
				{Criterion: "clarity", Score: 6.0, Explanation: "ok"},
			},
		},
		{
			Model: "model-a", Status: ResultStatusSuccess,
			Metrics: ResultMetrics{TTFB: 150, TotalLatency: 250},
			JudgeScores: []JudgeScore{
				{Criterion: "accuracy", Score: 10.0, Explanation: "perfect"},
				{Criterion: "clarity", Score: 8.0, Explanation: "great"},
			},
		},
	}

	summary := computeSummary(results, []string{"accuracy", "clarity"})
	a := summary.Models[0]
	if a.AvgJudgeScores == nil {
		t.Fatal("AvgJudgeScores is nil")
	}
	if a.AvgJudgeScores["accuracy"] != 9.0 {
		t.Errorf("accuracy expected 9.0, got %f", a.AvgJudgeScores["accuracy"])
	}
	if a.AvgJudgeScores["clarity"] != 7.0 {
		t.Errorf("clarity expected 7.0, got %f", a.AvgJudgeScores["clarity"])
	}
}

func TestComputeSummary_NoJudgeScoresWhenNoCriteria(t *testing.T) {
	results := []Result{
		{Model: "model-a", Status: ResultStatusSuccess, Metrics: ResultMetrics{TTFB: 100, TotalLatency: 200}},
	}
	summary := computeSummary(results, nil)
	a := summary.Models[0]
	if a.AvgJudgeScores != nil {
		t.Errorf("AvgJudgeScores should be nil when no criteria, got %v", a.AvgJudgeScores)
	}
}

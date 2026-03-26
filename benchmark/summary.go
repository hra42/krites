package benchmark

import (
	"math"
	"sort"
)

// computeSummary aggregates results per model into a Summary with percentiles and judge scores.
func computeSummary(results []Result, criteria []string) *Summary {
	grouped := make(map[string][]Result)
	for _, r := range results {
		grouped[r.Model] = append(grouped[r.Model], r)
	}

	var modelSummaries []ModelSummary
	for model, allResults := range grouped {
		var successful []Result
		for _, r := range allResults {
			if r.Status == ResultStatusSuccess {
				successful = append(successful, r)
			}
		}

		ms := ModelSummary{Model: model}

		if len(successful) > 0 {
			ttfbs := make([]float64, len(successful))
			latencies := make([]float64, len(successful))
			toksPerSec := make([]float64, len(successful))
			costs := make([]float64, len(successful))

			for i, r := range successful {
				ttfbs[i] = r.Metrics.TTFB
				latencies[i] = r.Metrics.TotalLatency
				toksPerSec[i] = r.Metrics.TokensPerSecond
				costs[i] = r.Metrics.EstimatedCost
			}

			ms.AvgTTFB = avg(ttfbs)
			ms.P50TTFB = percentile(ttfbs, 0.50)
			ms.P95TTFB = percentile(ttfbs, 0.95)
			ms.AvgLatency = avg(latencies)
			ms.P50Latency = percentile(latencies, 0.50)
			ms.P95Latency = percentile(latencies, 0.95)
			ms.AvgTokensPerSecond = avg(toksPerSec)
			ms.AvgCost = avg(costs)
			ms.TotalCost = sum(costs)
		}

		ms.SuccessRate = float64(len(successful)) / float64(len(allResults))

		// Judge score aggregation
		if len(criteria) > 0 {
			judgeAgg := make(map[string][]float64)
			for _, r := range successful {
				for _, js := range r.JudgeScores {
					judgeAgg[js.Criterion] = append(judgeAgg[js.Criterion], js.Score)
				}
			}
			if len(judgeAgg) > 0 {
				ms.AvgJudgeScores = make(map[string]float64)
				for criterion, scores := range judgeAgg {
					ms.AvgJudgeScores[criterion] = avg(scores)
				}
			}
		}

		modelSummaries = append(modelSummaries, ms)
	}

	// Sort by model name for deterministic output
	sort.Slice(modelSummaries, func(i, j int) bool {
		return modelSummaries[i].Model < modelSummaries[j].Model
	})

	return &Summary{Models: modelSummaries}
}

// percentile calculates the p-th percentile using linear interpolation.
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	if len(sorted) == 1 {
		return sorted[0]
	}

	k := p * float64(len(sorted)-1)
	f := math.Floor(k)
	c := math.Ceil(k)

	if f == c {
		return sorted[int(f)]
	}

	return sorted[int(f)]*(c-k) + sorted[int(c)]*(k-f)
}

func avg(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return sum(values) / float64(len(values))
}

func sum(values []float64) float64 {
	var total float64
	for _, v := range values {
		total += v
	}
	return total
}

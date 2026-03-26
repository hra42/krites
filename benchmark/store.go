package benchmark

import "time"

// Store defines the interface for benchmark data persistence.
type Store interface {
	CreateSuite(suite *Suite) error
	GetSuite(id string) (*Suite, error)
	ListSuites() ([]*Suite, error)
	UpdateSuite(suite *Suite) error
	DeleteSuite(id string) error

	CreateRun(run *Run) error
	GetRun(id string) (*Run, error)
	ListRuns() ([]*Run, error)
	UpdateRun(run *Run) error
}

// AnalyticsStore extends Store with analytical query methods.
type AnalyticsStore interface {
	Store
	GetOverview() (*AnalyticsOverview, error)
	GetModelTrends(modelID string, limit int) ([]ModelTrendPoint, error)
	GetCrossRunComparison(suiteID string) ([]CrossRunModelStats, error)
}

// AnalyticsOverview holds high-level platform statistics.
type AnalyticsOverview struct {
	TotalRuns      int     `json:"total_runs"`
	CompletedRuns  int     `json:"completed_runs"`
	TotalResults   int     `json:"total_results"`
	DistinctModels int     `json:"distinct_models"`
	TotalCost      float64 `json:"total_cost"`
	AvgLatency     float64 `json:"avg_latency_ms"`
}

// ModelTrendPoint represents a model's performance in a single run.
type ModelTrendPoint struct {
	RunID      string    `json:"run_id"`
	SuiteName  string    `json:"suite_name"`
	CreatedAt  time.Time `json:"created_at"`
	AvgLatency float64   `json:"avg_latency_ms"`
	AvgTTFB    float64   `json:"avg_ttfb_ms"`
	AvgTPS     float64   `json:"avg_tokens_per_second"`
	AvgCost    float64   `json:"avg_cost"`
	SuccessRate float64  `json:"success_rate"`
}

// CrossRunModelStats holds aggregated stats for a model across multiple runs.
type CrossRunModelStats struct {
	Model          string  `json:"model"`
	RunCount       int     `json:"run_count"`
	AvgLatency     float64 `json:"avg_latency_ms"`
	AvgTTFB        float64 `json:"avg_ttfb_ms"`
	AvgTPS         float64 `json:"avg_tokens_per_second"`
	TotalCost      float64 `json:"total_cost"`
	AvgSuccessRate float64 `json:"avg_success_rate"`
}

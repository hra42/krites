package benchmark

import (
	"errors"
	"time"
)

var (
	ErrSuiteNotFound = errors.New("suite not found")
	ErrRunNotFound   = errors.New("run not found")
)

// Prompt represents a single test prompt within a suite.
type Prompt struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	SystemMessage  string `json:"system_message"`
	UserMessage    string `json:"user_message"`
	ExpectedOutput string `json:"expected_output,omitempty"`
	Category       string `json:"category,omitempty"`
}

// RunConfig holds LLM and execution parameters for a benchmark run.
type RunConfig struct {
	Temperature    float64  `json:"temperature"`
	MaxTokens      int      `json:"max_tokens"`
	TopP           float64  `json:"top_p"`
	Iterations     int      `json:"iterations"`
	Concurrency    int      `json:"concurrency"`
	JudgeEnabled   bool     `json:"judge_enabled"`
	JudgeModel     string   `json:"judge_model,omitempty"`
	JudgeCriteria  []string `json:"judge_criteria,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds"`
}

// DefaultRunConfig returns sensible defaults for a RunConfig.
func DefaultRunConfig() RunConfig {
	return RunConfig{
		Temperature:    0.7,
		MaxTokens:      1024,
		TopP:           1.0,
		Iterations:     1,
		Concurrency:    3,
		TimeoutSeconds: 30,
	}
}

// Suite is a reusable test collection of prompts, models, and configuration.
type Suite struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Prompts     []Prompt  `json:"prompts"`
	Models      []string  `json:"models"`
	Config      RunConfig `json:"config"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RunStatus represents the lifecycle state of a benchmark run.
type RunStatus string

const (
	RunStatusPending  RunStatus = "pending"
	RunStatusRunning  RunStatus = "running"
	RunStatusComplete RunStatus = "complete"
	RunStatusFailed   RunStatus = "failed"
	RunStatusCanceled RunStatus = "canceled"
)

// ResultMetrics holds timing and token measurements for a single API call.
type ResultMetrics struct {
	TTFB             float64 `json:"ttfb_ms"`
	TotalLatency     float64 `json:"total_latency_ms"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TokensPerSecond  float64 `json:"tokens_per_second"`
	EstimatedCost    float64 `json:"estimated_cost"`
}

// JudgeScore holds a single criterion evaluation from the LLM judge.
type JudgeScore struct {
	Criterion   string  `json:"criterion"`
	Score       float64 `json:"score"`
	Explanation string  `json:"explanation"`
}

// ResultStatus represents the outcome of a single API call.
type ResultStatus string

const (
	ResultStatusSuccess ResultStatus = "success"
	ResultStatusError   ResultStatus = "error"
	ResultStatusTimeout ResultStatus = "timeout"
)

// Result is a single model response to a prompt with metrics.
type Result struct {
	ID          string        `json:"id"`
	RunID       string        `json:"run_id"`
	PromptID    string        `json:"prompt_id"`
	PromptName  string        `json:"prompt_name"`
	Model       string        `json:"model"`
	Iteration   int           `json:"iteration"`
	Response    string        `json:"response"`
	Metrics     ResultMetrics `json:"metrics"`
	JudgeScores []JudgeScore  `json:"judge_scores,omitempty"`
	Status      ResultStatus  `json:"status"`
	Error       string        `json:"error,omitempty"`
}

// ModelSummary holds aggregated statistics for one model across a run.
type ModelSummary struct {
	Model              string             `json:"model"`
	AvgTTFB            float64            `json:"avg_ttfb_ms"`
	P50TTFB            float64            `json:"p50_ttfb_ms"`
	P95TTFB            float64            `json:"p95_ttfb_ms"`
	AvgLatency         float64            `json:"avg_latency_ms"`
	P50Latency         float64            `json:"p50_latency_ms"`
	P95Latency         float64            `json:"p95_latency_ms"`
	AvgTokensPerSecond float64            `json:"avg_tokens_per_second"`
	AvgCost            float64            `json:"avg_cost"`
	TotalCost          float64            `json:"total_cost"`
	SuccessRate        float64            `json:"success_rate"`
	AvgJudgeScores     map[string]float64 `json:"avg_judge_scores,omitempty"`
}

// Summary holds aggregated statistics for an entire benchmark run.
type Summary struct {
	Models []ModelSummary `json:"models"`
}

// Run is a single execution of a suite.
type Run struct {
	ID        string     `json:"id"`
	SuiteID   string     `json:"suite_id"`
	SuiteName string     `json:"suite_name"`
	Status    RunStatus  `json:"status"`
	Results   []Result   `json:"results,omitempty"`
	Summary   *Summary   `json:"summary,omitempty"`
	Config    RunConfig  `json:"config"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

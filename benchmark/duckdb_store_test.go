package benchmark

import (
	"testing"
	"time"

	"github.com/hra42/krites/database"
)

func newDuckDBTestStore(t *testing.T) *DuckDBStore {
	t.Helper()
	db, err := database.OpenDB("", 1) // in-memory DuckDB
	if err != nil {
		t.Fatalf("opening test duckdb: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	store, err := NewDuckDBStore(db)
	if err != nil {
		t.Fatalf("creating duckdb store: %v", err)
	}
	return store
}

func newTestSuiteDuckDB(id, name string) *Suite {
	now := time.Now().Truncate(time.Microsecond)
	return &Suite{
		ID:          id,
		Name:        name,
		Description: "Test suite",
		Prompts: []Prompt{
			{ID: "p1", Name: "greeting", SystemMessage: "Be helpful", UserMessage: "Hello"},
		},
		Models: []string{"model-a", "model-b"},
		Config: RunConfig{
			Temperature:    0.7,
			MaxTokens:      100,
			TopP:           1.0,
			Iterations:     2,
			Concurrency:    3,
			TimeoutSeconds: 30,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestDuckDBStore_SuiteCRUD(t *testing.T) {
	store := newDuckDBTestStore(t)

	suite := newTestSuiteDuckDB("s1", "My Suite")

	// Create
	if err := store.CreateSuite(suite); err != nil {
		t.Fatalf("CreateSuite: %v", err)
	}

	// Get
	got, err := store.GetSuite("s1")
	if err != nil {
		t.Fatalf("GetSuite: %v", err)
	}
	if got.Name != "My Suite" {
		t.Errorf("expected name 'My Suite', got %q", got.Name)
	}
	if len(got.Prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(got.Prompts))
	}
	if got.Prompts[0].UserMessage != "Hello" {
		t.Errorf("expected user message 'Hello', got %q", got.Prompts[0].UserMessage)
	}
	if len(got.Models) != 2 {
		t.Errorf("expected 2 models, got %d", len(got.Models))
	}
	if got.Config.Iterations != 2 {
		t.Errorf("expected 2 iterations, got %d", got.Config.Iterations)
	}

	// Update
	suite.Name = "Updated Suite"
	suite.UpdatedAt = time.Now().Truncate(time.Microsecond)
	if err := store.UpdateSuite(suite); err != nil {
		t.Fatalf("UpdateSuite: %v", err)
	}
	got, _ = store.GetSuite("s1")
	if got.Name != "Updated Suite" {
		t.Errorf("expected 'Updated Suite', got %q", got.Name)
	}

	// List
	store.CreateSuite(newTestSuiteDuckDB("s2", "Second Suite"))
	suites, err := store.ListSuites()
	if err != nil {
		t.Fatalf("ListSuites: %v", err)
	}
	if len(suites) != 2 {
		t.Errorf("expected 2 suites, got %d", len(suites))
	}

	// Delete
	if err := store.DeleteSuite("s1"); err != nil {
		t.Fatalf("DeleteSuite: %v", err)
	}
	_, err = store.GetSuite("s1")
	if err != ErrSuiteNotFound {
		t.Errorf("expected ErrSuiteNotFound, got %v", err)
	}
}

func TestDuckDBStore_SuiteNotFound(t *testing.T) {
	store := newDuckDBTestStore(t)

	_, err := store.GetSuite("nonexistent")
	if err != ErrSuiteNotFound {
		t.Errorf("expected ErrSuiteNotFound, got %v", err)
	}

	err = store.UpdateSuite(&Suite{ID: "nonexistent"})
	if err != ErrSuiteNotFound {
		t.Errorf("expected ErrSuiteNotFound on update, got %v", err)
	}

	err = store.DeleteSuite("nonexistent")
	if err != ErrSuiteNotFound {
		t.Errorf("expected ErrSuiteNotFound on delete, got %v", err)
	}
}

func TestDuckDBStore_RunCRUD(t *testing.T) {
	store := newDuckDBTestStore(t)

	suite := newTestSuiteDuckDB("s1", "Suite")
	store.CreateSuite(suite)

	now := time.Now().Truncate(time.Microsecond)
	run := &Run{
		ID:        "r1",
		SuiteID:   "s1",
		SuiteName: "Suite",
		Status:    RunStatusPending,
		Config:    suite.Config,
		CreatedAt: now,
	}

	// Create
	if err := store.CreateRun(run); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	// Get (empty run)
	got, err := store.GetRun("r1")
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if got.Status != RunStatusPending {
		t.Errorf("expected pending, got %s", got.Status)
	}
	if len(got.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(got.Results))
	}

	// Update with results
	startTime := now
	run.Status = RunStatusRunning
	run.StartedAt = &startTime
	run.Results = []Result{
		{
			ID: "res1", RunID: "r1", PromptID: "p1", PromptName: "greeting",
			Model: "model-a", Iteration: 1, Response: "Hi!",
			Status: ResultStatusSuccess,
			Metrics: ResultMetrics{
				TTFB: 100, TotalLatency: 150,
				PromptTokens: 10, CompletionTokens: 5,
				TokensPerSecond: 33.3, EstimatedCost: 0.001,
			},
			JudgeScores: []JudgeScore{
				{Criterion: "accuracy", Score: 8.5, Explanation: "Good"},
			},
		},
	}
	if err := store.UpdateRun(run); err != nil {
		t.Fatalf("UpdateRun: %v", err)
	}

	// Verify results persisted
	got, err = store.GetRun("r1")
	if err != nil {
		t.Fatalf("GetRun after update: %v", err)
	}
	if got.Status != RunStatusRunning {
		t.Errorf("expected running, got %s", got.Status)
	}
	if len(got.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got.Results))
	}
	if got.Results[0].Model != "model-a" {
		t.Errorf("expected model-a, got %s", got.Results[0].Model)
	}
	if got.Results[0].Metrics.EstimatedCost != 0.001 {
		t.Errorf("expected cost 0.001, got %f", got.Results[0].Metrics.EstimatedCost)
	}
	if len(got.Results[0].JudgeScores) != 1 {
		t.Fatalf("expected 1 judge score, got %d", len(got.Results[0].JudgeScores))
	}
	if got.Results[0].JudgeScores[0].Score != 8.5 {
		t.Errorf("expected score 8.5, got %f", got.Results[0].JudgeScores[0].Score)
	}

	// Complete run and verify summary
	endTime := time.Now().Truncate(time.Microsecond)
	run.Status = RunStatusComplete
	run.EndedAt = &endTime
	if err := store.UpdateRun(run); err != nil {
		t.Fatalf("UpdateRun complete: %v", err)
	}

	got, _ = store.GetRun("r1")
	if got.Summary == nil {
		t.Fatal("expected summary to be computed")
	}
	if len(got.Summary.Models) != 1 {
		t.Errorf("expected 1 model summary, got %d", len(got.Summary.Models))
	}

	// List
	runs, err := store.ListRuns()
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}
}

func TestDuckDBStore_RunNotFound(t *testing.T) {
	store := newDuckDBTestStore(t)

	_, err := store.GetRun("nonexistent")
	if err != ErrRunNotFound {
		t.Errorf("expected ErrRunNotFound, got %v", err)
	}
}

func TestDuckDBStore_UpdateRunIdempotent(t *testing.T) {
	store := newDuckDBTestStore(t)

	now := time.Now().Truncate(time.Microsecond)
	run := &Run{
		ID: "r1", SuiteID: "s1", SuiteName: "Suite",
		Status: RunStatusRunning, Config: DefaultRunConfig(), CreatedAt: now,
		Results: []Result{
			{ID: "res1", RunID: "r1", Model: "m1", Status: ResultStatusSuccess,
				Metrics: ResultMetrics{TotalLatency: 100}},
		},
	}
	store.CreateRun(run)
	store.UpdateRun(run)

	// Add another result and update again (res1 should not be duplicated)
	run.Results = append(run.Results, Result{
		ID: "res2", RunID: "r1", Model: "m1", Status: ResultStatusSuccess,
		Metrics: ResultMetrics{TotalLatency: 200},
	})
	if err := store.UpdateRun(run); err != nil {
		t.Fatalf("second UpdateRun: %v", err)
	}

	got, _ := store.GetRun("r1")
	if len(got.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(got.Results))
	}
}

func TestDuckDBStore_Analytics(t *testing.T) {
	store := newDuckDBTestStore(t)

	suite := newTestSuiteDuckDB("s1", "Suite")
	store.CreateSuite(suite)

	now := time.Now().Truncate(time.Microsecond)
	run := &Run{
		ID: "r1", SuiteID: "s1", SuiteName: "Suite",
		Status: RunStatusComplete, Config: suite.Config, CreatedAt: now,
		Results: []Result{
			{ID: "res1", RunID: "r1", Model: "model-a", Status: ResultStatusSuccess,
				Metrics: ResultMetrics{TotalLatency: 100, TTFB: 50, TokensPerSecond: 30, EstimatedCost: 0.01}},
			{ID: "res2", RunID: "r1", Model: "model-b", Status: ResultStatusSuccess,
				Metrics: ResultMetrics{TotalLatency: 200, TTFB: 100, TokensPerSecond: 20, EstimatedCost: 0.02}},
		},
	}
	store.CreateRun(run)
	store.UpdateRun(run)

	// Overview
	overview, err := store.GetOverview()
	if err != nil {
		t.Fatalf("GetOverview: %v", err)
	}
	if overview.TotalRuns != 1 {
		t.Errorf("expected 1 total run, got %d", overview.TotalRuns)
	}
	if overview.CompletedRuns != 1 {
		t.Errorf("expected 1 completed run, got %d", overview.CompletedRuns)
	}
	if overview.TotalResults != 2 {
		t.Errorf("expected 2 total results, got %d", overview.TotalResults)
	}
	if overview.DistinctModels != 2 {
		t.Errorf("expected 2 distinct models, got %d", overview.DistinctModels)
	}

	// Model trends
	trends, err := store.GetModelTrends("model-a", 10)
	if err != nil {
		t.Fatalf("GetModelTrends: %v", err)
	}
	if len(trends) != 1 {
		t.Errorf("expected 1 trend point, got %d", len(trends))
	}

	// Cross-run comparison
	stats, err := store.GetCrossRunComparison("")
	if err != nil {
		t.Fatalf("GetCrossRunComparison: %v", err)
	}
	if len(stats) != 2 {
		t.Errorf("expected 2 model stats, got %d", len(stats))
	}

	// With suite filter
	stats, err = store.GetCrossRunComparison("s1")
	if err != nil {
		t.Fatalf("GetCrossRunComparison with suite: %v", err)
	}
	if len(stats) != 2 {
		t.Errorf("expected 2 model stats, got %d", len(stats))
	}
}

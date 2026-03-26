package benchmark

import (
	"fmt"
	"testing"
	"time"

	"github.com/hra42/krites/models"
)

func newTestSuiteForRunner() *Suite {
	return &Suite{
		ID:          "suite-1",
		Name:        "Runner Test Suite",
		Description: "Test",
		Prompts: []Prompt{
			{ID: "p1", Name: "greeting", UserMessage: "Hello", SystemMessage: "Be helpful"},
		},
		Models: []string{"test-model-a"},
		Config: RunConfig{
			Temperature:    0.7,
			MaxTokens:      100,
			TopP:           1.0,
			Iterations:     1,
			Concurrency:    3,
			TimeoutSeconds: 5,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestRunner_StartRun_CreatesRunWithPendingStatus(t *testing.T) {
	store := NewMemoryStore()
	mock := &mockChatCompleter{response: defaultMockResponse(), delay: 100 * time.Millisecond}
	broadcaster := NewBroadcaster()
	runner := NewRunner(store, mock, broadcaster, nil)

	suite := newTestSuiteForRunner()
	store.CreateSuite(suite)

	run, err := runner.StartRun(suite)
	if err != nil {
		t.Fatalf("StartRun: %v", err)
	}

	if run.Status != RunStatusPending {
		t.Errorf("status = %v, want pending", run.Status)
	}
	if run.SuiteID != suite.ID {
		t.Errorf("suite_id = %v, want %v", run.SuiteID, suite.ID)
	}
	if run.SuiteName != suite.Name {
		t.Errorf("suite_name = %v, want %v", run.SuiteName, suite.Name)
	}

	// Verify it's in the store
	stored, err := store.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if stored.ID != run.ID {
		t.Errorf("stored ID = %v, want %v", stored.ID, run.ID)
	}
}

func TestRunner_ExecuteRun_CompletesSuccessfully(t *testing.T) {
	store := NewMemoryStore()
	mock := &mockChatCompleter{response: defaultMockResponse()}
	broadcaster := NewBroadcaster()
	runner := NewRunner(store, mock, broadcaster, nil)

	suite := newTestSuiteForRunner()
	suite.Models = []string{"model-a", "model-b"}
	suite.Config.Iterations = 2
	store.CreateSuite(suite)

	run, _ := runner.StartRun(suite)

	// Wait for the runner goroutine to finish
	time.Sleep(500 * time.Millisecond)

	run, _ = store.GetRun(run.ID)

	if run.Status != RunStatusComplete {
		t.Errorf("status = %v, want complete", run.Status)
	}
	// 2 models × 1 prompt × 2 iterations = 4 results
	if len(run.Results) != 4 {
		t.Errorf("result count = %d, want 4", len(run.Results))
	}
	if run.StartedAt == nil {
		t.Error("expected StartedAt to be set")
	}
	if run.EndedAt == nil {
		t.Error("expected EndedAt to be set")
	}

	for _, res := range run.Results {
		if res.Status != ResultStatusSuccess {
			t.Errorf("result %v status = %v, want success", res.ID, res.Status)
		}
		if res.Response != "Hello!" {
			t.Errorf("response = %q, want 'Hello!'", res.Response)
		}
		if res.Metrics.TotalLatency < 0 {
			t.Error("expected non-negative total latency")
		}
		if res.Metrics.PromptTokens != 10 {
			t.Errorf("prompt_tokens = %d, want 10", res.Metrics.PromptTokens)
		}
		if res.Metrics.CompletionTokens != 5 {
			t.Errorf("completion_tokens = %d, want 5", res.Metrics.CompletionTokens)
		}
		if res.Metrics.TotalLatency > 0 && res.Metrics.TokensPerSecond <= 0 {
			t.Error("expected positive tokens_per_second when latency > 0")
		}
	}
}

func TestRunner_ExecuteRun_HandlesAPIError(t *testing.T) {
	store := NewMemoryStore()
	mock := &mockChatCompleter{err: fmt.Errorf("api error")}
	broadcaster := NewBroadcaster()
	runner := NewRunner(store, mock, broadcaster, nil)

	suite := newTestSuiteForRunner()
	store.CreateSuite(suite)

	run, _ := runner.StartRun(suite)
	time.Sleep(500 * time.Millisecond)

	run, _ = store.GetRun(run.ID)

	// All results failed → run should be failed
	if run.Status != RunStatusFailed {
		t.Errorf("status = %v, want failed", run.Status)
	}
	if len(run.Results) != 1 {
		t.Fatalf("result count = %d, want 1", len(run.Results))
	}
	if run.Results[0].Status != ResultStatusError {
		t.Errorf("result status = %v, want error", run.Results[0].Status)
	}
	if run.Results[0].Error == "" {
		t.Error("expected error message in result")
	}
}

func TestRunner_ExecuteRun_HandlesTimeout(t *testing.T) {
	store := NewMemoryStore()
	mock := &mockChatCompleter{response: defaultMockResponse(), delay: 3 * time.Second}
	broadcaster := NewBroadcaster()
	runner := NewRunner(store, mock, broadcaster, nil)

	suite := newTestSuiteForRunner()
	suite.Config.TimeoutSeconds = 1
	store.CreateSuite(suite)

	run, _ := runner.StartRun(suite)
	time.Sleep(2 * time.Second)

	run, _ = store.GetRun(run.ID)

	if len(run.Results) != 1 {
		t.Fatalf("result count = %d, want 1", len(run.Results))
	}
	if run.Results[0].Status != ResultStatusTimeout {
		t.Errorf("result status = %v, want timeout", run.Results[0].Status)
	}
}

func TestRunner_SSEEvents_Published(t *testing.T) {
	store := NewMemoryStore()
	mock := &mockChatCompleter{response: defaultMockResponse()}
	broadcaster := NewBroadcaster()
	runner := NewRunner(store, mock, broadcaster, nil)

	suite := newTestSuiteForRunner()
	store.CreateSuite(suite)

	// We need to subscribe BEFORE calling StartRun, but we don't know the run ID yet.
	// So we call StartRun, get the ID, then subscribe quickly.
	// Alternative: subscribe after creating the run in the store but before the goroutine publishes.
	// Since StartRun creates the run and launches the goroutine, there's a race.
	// We use a small delay mock to give us time to subscribe.
	mock.delay = 200 * time.Millisecond

	run, _ := runner.StartRun(suite)
	ch, unsub := broadcaster.Subscribe(run.ID)
	defer unsub()

	var events []SSEEvent
	timeout := time.After(3 * time.Second)
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				goto done
			}
			events = append(events, event)
		case <-timeout:
			t.Fatal("timed out waiting for events")
		}
	}
done:

	if len(events) < 2 {
		t.Fatalf("expected at least 2 events (result_completed + run_completed), got %d", len(events))
	}

	// Last event should be run_completed or run_error
	lastEvent := events[len(events)-1]
	if lastEvent.Type != EventRunCompleted && lastEvent.Type != EventRunError {
		t.Errorf("last event type = %v, want run_completed or run_error", lastEvent.Type)
	}

	// Should have at least one result_completed
	hasResult := false
	for _, e := range events {
		if e.Type == EventResultCompleted {
			hasResult = true
			break
		}
	}
	if !hasResult {
		t.Error("expected at least one result_completed event")
	}
}

func TestRunner_ExecuteRun_WithJudge(t *testing.T) {
	store := NewMemoryStore()
	mock := &mockChatCompleter{
		onRequest: func(req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
			if req.Model == "judge-model" {
				return &models.ChatCompletionResponse{
					ID:    "judge-resp",
					Model: "judge-model",
					Choices: []models.Choice{
						{Index: 0, Message: models.ChatMessage{Role: "assistant", Content: `{"score": 8.5, "explanation": "Good answer"}`}, FinishReason: "stop"},
					},
					Usage: &models.Usage{PromptTokens: 50, CompletionTokens: 20, TotalTokens: 70},
				}, nil
			}
			return defaultMockResponse(), nil
		},
	}
	broadcaster := NewBroadcaster()
	runner := NewRunner(store, mock, broadcaster, nil)

	suite := newTestSuiteForRunner()
	suite.Config.JudgeEnabled = true
	suite.Config.JudgeModel = "judge-model"
	suite.Config.JudgeCriteria = []string{"accuracy", "clarity"}
	store.CreateSuite(suite)

	run, _ := runner.StartRun(suite)

	// Subscribe to capture judge_scored events
	ch, unsub := broadcaster.Subscribe(run.ID)
	defer unsub()

	var judgeEvents []SSEEvent
	timeout := time.After(5 * time.Second)
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				goto done
			}
			if event.Type == EventJudgeScored {
				judgeEvents = append(judgeEvents, event)
			}
		case <-timeout:
			t.Fatal("timed out waiting for events")
		}
	}
done:

	run, _ = store.GetRun(run.ID)

	if run.Status != RunStatusComplete {
		t.Errorf("status = %v, want complete", run.Status)
	}
	if len(run.Results) != 1 {
		t.Fatalf("result count = %d, want 1", len(run.Results))
	}

	result := run.Results[0]
	if len(result.JudgeScores) != 2 {
		t.Fatalf("judge score count = %d, want 2", len(result.JudgeScores))
	}
	for _, js := range result.JudgeScores {
		if js.Score != 8.5 {
			t.Errorf("score = %f, want 8.5", js.Score)
		}
		if js.Explanation != "Good answer" {
			t.Errorf("explanation = %q, want 'Good answer'", js.Explanation)
		}
	}

	if len(judgeEvents) != 1 {
		t.Errorf("expected 1 judge_scored event, got %d", len(judgeEvents))
	}
}

func TestRunner_ExecuteRun_JudgeParseError(t *testing.T) {
	store := NewMemoryStore()
	mock := &mockChatCompleter{
		onRequest: func(req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
			if req.Model == "judge-model" {
				return &models.ChatCompletionResponse{
					ID:    "judge-resp",
					Model: "judge-model",
					Choices: []models.Choice{
						{Index: 0, Message: models.ChatMessage{Role: "assistant", Content: "This is not valid JSON"}, FinishReason: "stop"},
					},
					Usage: &models.Usage{PromptTokens: 50, CompletionTokens: 20, TotalTokens: 70},
				}, nil
			}
			return defaultMockResponse(), nil
		},
	}
	broadcaster := NewBroadcaster()
	runner := NewRunner(store, mock, broadcaster, nil)

	suite := newTestSuiteForRunner()
	suite.Config.JudgeEnabled = true
	suite.Config.JudgeModel = "judge-model"
	suite.Config.JudgeCriteria = []string{"accuracy"}
	store.CreateSuite(suite)

	run, _ := runner.StartRun(suite)
	time.Sleep(1 * time.Second)

	run, _ = store.GetRun(run.ID)

	if run.Status != RunStatusComplete {
		t.Errorf("status = %v, want complete (judge errors should not fail the run)", run.Status)
	}
	if len(run.Results[0].JudgeScores) != 0 {
		t.Errorf("expected 0 judge scores after parse error, got %d", len(run.Results[0].JudgeScores))
	}
}

func TestRunner_ExecuteRun_WithSummary(t *testing.T) {
	store := NewMemoryStore()
	mock := &mockChatCompleter{response: defaultMockResponse()}
	broadcaster := NewBroadcaster()
	runner := NewRunner(store, mock, broadcaster, nil)

	suite := newTestSuiteForRunner()
	suite.Models = []string{"model-a", "model-b"}
	suite.Config.Iterations = 2
	store.CreateSuite(suite)

	run, _ := runner.StartRun(suite)
	time.Sleep(1 * time.Second)

	run, _ = store.GetRun(run.ID)

	if run.Status != RunStatusComplete {
		t.Errorf("status = %v, want complete", run.Status)
	}
	if run.Summary == nil {
		t.Fatal("expected Summary to be set")
	}
	if len(run.Summary.Models) != 2 {
		t.Errorf("expected 2 model summaries, got %d", len(run.Summary.Models))
	}
	for _, ms := range run.Summary.Models {
		if ms.SuccessRate != 1.0 {
			t.Errorf("model %s success rate = %f, want 1.0", ms.Model, ms.SuccessRate)
		}
		if ms.AvgLatency < 0 {
			t.Errorf("model %s avg latency = %f, want >= 0", ms.Model, ms.AvgLatency)
		}
	}
}

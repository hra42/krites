package benchmark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	or "github.com/hra42/openrouter-go"

	"github.com/gofiber/fiber/v3"
	"github.com/hra42/krites/middleware"
	"github.com/hra42/krites/models"
)

// mockChatCompleter implements openrouter.ChatCompleter for tests.
type mockChatCompleter struct {
	response  *models.ChatCompletionResponse
	err       error
	delay     time.Duration
	onRequest func(req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error)
}

func (m *mockChatCompleter) ChatComplete(ctx context.Context, req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if m.onRequest != nil {
		return m.onRequest(req)
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *mockChatCompleter) ChatCompleteStream(_ context.Context, _ *models.ChatCompletionRequest) (*or.ChatStream, error) {
	return nil, fmt.Errorf("not implemented in test mock")
}

func defaultMockResponse() *models.ChatCompletionResponse {
	return &models.ChatCompletionResponse{
		ID:    "test-resp",
		Model: "test-model",
		Choices: []models.Choice{
			{Index: 0, Message: models.ChatMessage{Role: "assistant", Content: "Hello!"}, FinishReason: "stop"},
		},
		Usage: &models.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}
}

func setupBenchmarkApp(t *testing.T) (*fiber.App, *MemoryStore) {
	t.Helper()
	store := NewMemoryStore()
	mock := &mockChatCompleter{response: defaultMockResponse()}
	handler := NewHandler(store, mock, nil)
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	handler.RegisterRoutes(app)
	return app, store
}

func createTestSuiteViaAPI(t *testing.T, app *fiber.App) map[string]interface{} {
	t.Helper()
	reqBody := CreateSuiteRequest{
		Name:        "Test Suite",
		Description: "A test suite",
		Prompts: []Prompt{
			{Name: "greeting", UserMessage: "Hello, how are you?", SystemMessage: "Be helpful"},
		},
		Models: []string{"openai/gpt-4o", "anthropic/claude-3.5-sonnet"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/benchmarks/suites/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("create status = %d, want 201, body: %s", resp.StatusCode, respBody)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	return result
}

func TestHandleCreateSuite_Valid(t *testing.T) {
	app, _ := setupBenchmarkApp(t)
	result := createTestSuiteViaAPI(t, app)

	if result["id"] == "" {
		t.Error("expected non-empty ID")
	}
	if result["name"] != "Test Suite" {
		t.Errorf("name = %v, want 'Test Suite'", result["name"])
	}

	prompts := result["prompts"].([]interface{})
	if len(prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(prompts))
	}
	prompt := prompts[0].(map[string]interface{})
	if prompt["id"] == "" {
		t.Error("prompt should have auto-generated ID")
	}

	models := result["models"].([]interface{})
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models))
	}
}

func TestHandleCreateSuite_DefaultConfig(t *testing.T) {
	app, _ := setupBenchmarkApp(t)
	result := createTestSuiteViaAPI(t, app)

	cfg := result["config"].(map[string]interface{})
	if cfg["temperature"].(float64) != 0.7 {
		t.Errorf("default temperature = %v, want 0.7", cfg["temperature"])
	}
	if cfg["max_tokens"].(float64) != 1024 {
		t.Errorf("default max_tokens = %v, want 1024", cfg["max_tokens"])
	}
	if cfg["iterations"].(float64) != 1 {
		t.Errorf("default iterations = %v, want 1", cfg["iterations"])
	}
	if cfg["concurrency"].(float64) != 3 {
		t.Errorf("default concurrency = %v, want 3", cfg["concurrency"])
	}
}

func TestHandleCreateSuite_MissingName(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	reqBody := CreateSuiteRequest{
		Prompts: []Prompt{{UserMessage: "hi"}},
		Models:  []string{"openai/gpt-4o"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/benchmarks/suites/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestHandleCreateSuite_NoPrompts(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	reqBody := CreateSuiteRequest{
		Name:   "No Prompts",
		Models: []string{"openai/gpt-4o"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/benchmarks/suites/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestHandleCreateSuite_NoModels(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	reqBody := CreateSuiteRequest{
		Name:    "No Models",
		Prompts: []Prompt{{UserMessage: "hi"}},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/benchmarks/suites/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestHandleCreateSuite_EmptyUserMessage(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	reqBody := CreateSuiteRequest{
		Name:    "Bad Prompt",
		Prompts: []Prompt{{Name: "empty"}},
		Models:  []string{"openai/gpt-4o"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/benchmarks/suites/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestHandleListSuites_Empty(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	req := httptest.NewRequest(http.MethodGet, "/benchmarks/suites/", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var suites []SuiteListResponse
	json.Unmarshal(body, &suites)

	if len(suites) != 0 {
		t.Errorf("expected 0 suites, got %d", len(suites))
	}
}

func TestHandleListSuites_Multiple(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	createTestSuiteViaAPI(t, app)
	createTestSuiteViaAPI(t, app)

	req := httptest.NewRequest(http.MethodGet, "/benchmarks/suites/", nil)
	resp, _ := app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	var suites []SuiteListResponse
	json.Unmarshal(body, &suites)

	if len(suites) != 2 {
		t.Errorf("expected 2 suites, got %d", len(suites))
	}
	if suites[0].ModelCount != 2 {
		t.Errorf("model_count = %d, want 2", suites[0].ModelCount)
	}
	if suites[0].PromptCount != 1 {
		t.Errorf("prompt_count = %d, want 1", suites[0].PromptCount)
	}
}

func TestHandleGetSuite_Exists(t *testing.T) {
	app, _ := setupBenchmarkApp(t)
	created := createTestSuiteViaAPI(t, app)
	id := created["id"].(string)

	req := httptest.NewRequest(http.MethodGet, "/benchmarks/suites/"+id, nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var suite Suite
	json.Unmarshal(body, &suite)

	if suite.ID != id {
		t.Errorf("ID = %q, want %q", suite.ID, id)
	}
	if len(suite.Prompts) != 1 {
		t.Errorf("expected 1 prompt in detail, got %d", len(suite.Prompts))
	}
}

func TestHandleGetSuite_NotFound(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	req := httptest.NewRequest(http.MethodGet, "/benchmarks/suites/nonexistent", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestHandleUpdateSuite_Valid(t *testing.T) {
	app, _ := setupBenchmarkApp(t)
	created := createTestSuiteViaAPI(t, app)
	id := created["id"].(string)

	newName := "Updated Suite"
	reqBody := UpdateSuiteRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/benchmarks/suites/"+id, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200, body: %s", resp.StatusCode, respBody)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var suite Suite
	json.Unmarshal(respBody, &suite)

	if suite.Name != "Updated Suite" {
		t.Errorf("name = %q, want 'Updated Suite'", suite.Name)
	}
	if suite.Description != "A test suite" {
		t.Errorf("description should be unchanged, got %q", suite.Description)
	}
}

func TestHandleUpdateSuite_NotFound(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	newName := "Ghost"
	reqBody := UpdateSuiteRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/benchmarks/suites/nonexistent", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestHandleDeleteSuite_Exists(t *testing.T) {
	app, _ := setupBenchmarkApp(t)
	created := createTestSuiteViaAPI(t, app)
	id := created["id"].(string)

	req := httptest.NewRequest(http.MethodDelete, "/benchmarks/suites/"+id, nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/benchmarks/suites/"+id, nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("get after delete: status = %d, want 404", resp.StatusCode)
	}
}

func TestHandleDeleteSuite_NotFound(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	req := httptest.NewRequest(http.MethodDelete, "/benchmarks/suites/nonexistent", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestHandleStartRun_ValidSuite(t *testing.T) {
	app, _ := setupBenchmarkApp(t)
	created := createTestSuiteViaAPI(t, app)
	id := created["id"].(string)

	req := httptest.NewRequest(http.MethodPost, "/benchmarks/suites/"+id+"/run", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 202, body: %s", resp.StatusCode, body)
	}

	body, _ := io.ReadAll(resp.Body)
	var run map[string]interface{}
	json.Unmarshal(body, &run)

	if run["id"] == "" {
		t.Error("expected non-empty run ID")
	}
	if run["suite_id"] != id {
		t.Errorf("suite_id = %v, want %v", run["suite_id"], id)
	}
	status := run["status"].(string)
	if status != "pending" && status != "running" && status != "complete" {
		t.Errorf("status = %v, want pending/running/complete", status)
	}
}

func TestHandleStartRun_SuiteNotFound(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	req := httptest.NewRequest(http.MethodPost, "/benchmarks/suites/nonexistent/run", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestHandleListRuns_Empty(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	req := httptest.NewRequest(http.MethodGet, "/benchmarks/runs/", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var runs []RunListResponse
	json.Unmarshal(body, &runs)

	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestHandleGetRun_NotFound(t *testing.T) {
	app, _ := setupBenchmarkApp(t)

	req := httptest.NewRequest(http.MethodGet, "/benchmarks/runs/nonexistent", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestHandleGetRun_Exists(t *testing.T) {
	app, _ := setupBenchmarkApp(t)
	created := createTestSuiteViaAPI(t, app)
	id := created["id"].(string)

	// Start a run
	startReq := httptest.NewRequest(http.MethodPost, "/benchmarks/suites/"+id+"/run", nil)
	startResp, _ := app.Test(startReq)
	startBody, _ := io.ReadAll(startResp.Body)
	var runData map[string]interface{}
	json.Unmarshal(startBody, &runData)
	runID := runData["id"].(string)

	// Give runner goroutine a moment to finish
	time.Sleep(200 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/benchmarks/runs/"+runID, nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200, body: %s", resp.StatusCode, body)
	}
}

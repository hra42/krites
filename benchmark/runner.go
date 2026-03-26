package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hra42/krites/models"
	"github.com/hra42/krites/openrouter"
	"github.com/hra42/krites/pricing"
)

// workItem represents a single (model, prompt, iteration) tuple to execute.
type workItem struct {
	model     string
	prompt    Prompt
	iteration int
}

// Runner executes benchmark suites against the OpenRouter API.
type Runner struct {
	store        Store
	orClient     openrouter.ChatCompleter
	broadcaster  *Broadcaster
	pricingCache *pricing.PricingCache
}

// NewRunner creates a new Runner.
func NewRunner(store Store, orClient openrouter.ChatCompleter, broadcaster *Broadcaster, pricingCache *pricing.PricingCache) *Runner {
	return &Runner{
		store:        store,
		orClient:     orClient,
		broadcaster:  broadcaster,
		pricingCache: pricingCache,
	}
}

// StartRun creates a Run with status=pending, persists it, then launches
// a goroutine to execute all API calls. Returns the Run immediately.
func (r *Runner) StartRun(suite *Suite) (*Run, error) {
	now := time.Now()
	run := &Run{
		ID:        uuid.New().String(),
		SuiteID:   suite.ID,
		SuiteName: suite.Name,
		Status:    RunStatusPending,
		Config:    suite.Config,
		CreatedAt: now,
	}

	if err := r.store.CreateRun(run); err != nil {
		return nil, err
	}

	go r.executeRun(run, suite)

	return run, nil
}

// executeRun is the goroutine body that runs all benchmark calls.
func (r *Runner) executeRun(run *Run, suite *Suite) {
	now := time.Now()
	run.StartedAt = &now
	run.Status = RunStatusRunning
	_ = r.store.UpdateRun(run)

	r.broadcaster.Publish(run.ID, SSEEvent{
		Type: EventRunStarted,
		Data: run,
	})

	// Build work items
	var items []workItem
	for _, model := range suite.Models {
		for _, prompt := range suite.Prompts {
			for iter := 1; iter <= suite.Config.Iterations; iter++ {
				items = append(items, workItem{
					model:     model,
					prompt:    prompt,
					iteration: iter,
				})
			}
		}
	}

	// Semaphore for concurrency control
	sem := make(chan struct{}, suite.Config.Concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{} // acquire
		go func(wi workItem) {
			defer func() {
				<-sem // release
				wg.Done()
			}()

			result := r.executeSingleCall(wi.model, wi.prompt, wi.iteration, suite.Config)
			result.RunID = run.ID

			mu.Lock()
			run.Results = append(run.Results, result)
			mu.Unlock()

			_ = r.store.UpdateRun(run)

			r.broadcaster.Publish(run.ID, SSEEvent{
				Type: EventResultCompleted,
				Data: result,
			})
		}(item)
	}

	wg.Wait()

	// Phase 2: LLM-as-Judge scoring (if enabled)
	if run.Config.JudgeEnabled && run.Config.JudgeModel != "" && len(run.Config.JudgeCriteria) > 0 {
		r.runJudgePhase(run, suite)
	}

	// Phase 3: Compute summary
	run.Summary = computeSummary(run.Results, run.Config.JudgeCriteria)

	// Determine final status
	allFailed := true
	for _, res := range run.Results {
		if res.Status == ResultStatusSuccess {
			allFailed = false
			break
		}
	}

	endTime := time.Now()
	run.EndedAt = &endTime

	if len(run.Results) > 0 && allFailed {
		run.Status = RunStatusFailed
		_ = r.store.UpdateRun(run)
		r.broadcaster.Publish(run.ID, SSEEvent{
			Type: EventRunError,
			Data: run,
		})
	} else {
		run.Status = RunStatusComplete
		_ = r.store.UpdateRun(run)
		r.broadcaster.Publish(run.ID, SSEEvent{
			Type: EventRunCompleted,
			Data: run,
		})
	}

	r.broadcaster.CloseRun(run.ID)
}

// executeSingleCall makes one ChatComplete call, measures metrics, and returns a Result.
func (r *Runner) executeSingleCall(model string, prompt Prompt, iteration int, config RunConfig) Result {
	result := Result{
		ID:         uuid.New().String(),
		PromptID:   prompt.ID,
		PromptName: prompt.Name,
		Model:      model,
		Iteration:  iteration,
	}

	// Build messages
	var messages []models.ChatMessage
	if prompt.SystemMessage != "" {
		messages = append(messages, models.ChatMessage{Role: "system", Content: prompt.SystemMessage})
	}
	messages = append(messages, models.ChatMessage{Role: "user", Content: prompt.UserMessage})

	temp := config.Temperature
	maxTokens := config.MaxTokens
	topP := config.TopP
	req := &models.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		TopP:        &topP,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSeconds)*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := r.orClient.ChatComplete(ctx, req)
	totalLatency := time.Since(start)

	if err != nil {
		result.Status = ResultStatusError
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = ResultStatusTimeout
		}
		result.Error = err.Error()
		result.Metrics.TotalLatency = float64(totalLatency.Milliseconds())
		return result
	}

	result.Status = ResultStatusSuccess
	result.Metrics.TotalLatency = float64(totalLatency.Milliseconds())
	// TTFB equals TotalLatency for non-streaming calls
	result.Metrics.TTFB = result.Metrics.TotalLatency

	// Extract response text
	if len(resp.Choices) > 0 {
		result.Response = resp.Choices[0].Message.Content
	}

	// Extract token counts and calculate throughput
	if resp.Usage != nil {
		result.Metrics.PromptTokens = resp.Usage.PromptTokens
		result.Metrics.CompletionTokens = resp.Usage.CompletionTokens
		if result.Metrics.TotalLatency > 0 {
			result.Metrics.TokensPerSecond = float64(resp.Usage.CompletionTokens) / (result.Metrics.TotalLatency / 1000.0)
		}
		// Prefer actual cost from OpenRouter response over estimated cost
		if resp.Usage.Cost > 0 {
			result.Metrics.EstimatedCost = resp.Usage.Cost
		} else if r.pricingCache != nil {
			result.Metrics.EstimatedCost = r.pricingCache.EstimateCost(model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
		}
	}

	return result
}

// runJudgePhase scores all successful results using the configured judge model.
func (r *Runner) runJudgePhase(run *Run, suite *Suite) {
	promptMap := make(map[string]Prompt)
	for _, p := range suite.Prompts {
		promptMap[p.ID] = p
	}

	sem := make(chan struct{}, run.Config.Concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range run.Results {
		if run.Results[i].Status != ResultStatusSuccess {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer func() {
				<-sem
				wg.Done()
			}()

			result := &run.Results[idx]
			prompt := promptMap[result.PromptID]

			ctx, cancel := context.WithTimeout(context.Background(),
				time.Duration(run.Config.TimeoutSeconds)*time.Second)
			defer cancel()

			scores := r.judgeResult(ctx, result, prompt, run.Config)

			mu.Lock()
			result.JudgeScores = scores
			mu.Unlock()

			_ = r.store.UpdateRun(run)

			r.broadcaster.Publish(run.ID, SSEEvent{
				Type: EventJudgeScored,
				Data: *result,
			})
		}(i)
	}

	wg.Wait()
}

// judgeResponse is the expected JSON structure from the judge model.
type judgeResponse struct {
	Score       float64 `json:"score"`
	Explanation string  `json:"explanation"`
}

// judgeResult calls the judge model for each criterion and returns the scores.
func (r *Runner) judgeResult(ctx context.Context, result *Result, prompt Prompt, config RunConfig) []JudgeScore {
	var scores []JudgeScore

	for _, criterion := range config.JudgeCriteria {
		judgePrompt := buildJudgePrompt(criterion, prompt.UserMessage, result.Response, prompt.ExpectedOutput)

		temp := 0.1
		maxTokens := 200
		topP := 1.0
		req := &models.ChatCompletionRequest{
			Model: config.JudgeModel,
			Messages: []models.ChatMessage{
				{Role: "system", Content: "Du bist ein strenger aber fairer Bewerter von AI-Antworten. Antworte nur mit JSON."},
				{Role: "user", Content: judgePrompt},
			},
			Temperature: &temp,
			MaxTokens:   &maxTokens,
			TopP:        &topP,
		}

		resp, err := r.orClient.ChatComplete(ctx, req)
		if err != nil {
			log.Printf("judge error for result %s criterion %q: %v", result.ID, criterion, err)
			continue
		}

		if len(resp.Choices) == 0 {
			log.Printf("judge returned no choices for result %s criterion %q", result.ID, criterion)
			continue
		}

		responseText := resp.Choices[0].Message.Content
		var jr judgeResponse
		if err := json.Unmarshal([]byte(responseText), &jr); err != nil {
			log.Printf("judge JSON parse error for result %s criterion %q: %v (response: %s)", result.ID, criterion, err, responseText)
			continue
		}

		if jr.Score < 1 || jr.Score > 10 {
			log.Printf("judge score out of range for result %s criterion %q: %f", result.ID, criterion, jr.Score)
			continue
		}

		scores = append(scores, JudgeScore{
			Criterion:   criterion,
			Score:       jr.Score,
			Explanation: jr.Explanation,
		})
	}

	return scores
}

// buildJudgePrompt creates the evaluation prompt for the judge model.
func buildJudgePrompt(criterion, userMessage, response, expectedOutput string) string {
	prompt := fmt.Sprintf(`Bewerte die folgende AI-Antwort auf einer Skala von 1-10 fuer "%s".

Original-Prompt: %s
AI-Antwort: %s`, criterion, userMessage, response)

	if expectedOutput != "" {
		prompt += fmt.Sprintf("\nErwartete Antwort: %s", expectedOutput)
	}

	prompt += `

Antworte NUR mit einem JSON-Objekt: {"score": <zahl>, "explanation": "<kurze Begruendung>"}`

	return prompt
}

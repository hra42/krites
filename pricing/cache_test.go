package pricing

import (
	"context"
	"sync"
	"testing"
	"time"

	or "github.com/hra42/openrouter-go"
)

type mockModelLister struct {
	mu       sync.Mutex
	response *or.ModelsResponse
	err      error
	calls    int
}

func (m *mockModelLister) ListModels(_ context.Context) (*or.ModelsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	return m.response, m.err
}

func newTestModelsResponse() *or.ModelsResponse {
	return &or.ModelsResponse{
		Data: []or.Model{
			{
				ID:   "openai/gpt-4o",
				Name: "GPT-4o",
				Pricing: or.ModelPricing{
					Prompt:     "2.5",     // $2.50 per 1M tokens
					Completion: "10",      // $10.00 per 1M tokens
				},
			},
			{
				ID:   "anthropic/claude-3.5-sonnet",
				Name: "Claude 3.5 Sonnet",
				Pricing: or.ModelPricing{
					Prompt:     "3",       // $3.00 per 1M tokens
					Completion: "15",      // $15.00 per 1M tokens
				},
			},
			{
				ID:   "meta-llama/llama-3-8b",
				Name: "Llama 3 8B",
				Pricing: or.ModelPricing{
					Prompt:     "0.05",    // $0.05 per 1M tokens
					Completion: "0.08",    // $0.08 per 1M tokens
				},
			},
		},
	}
}

func TestPricingCache_Start(t *testing.T) {
	mock := &mockModelLister{response: newTestModelsResponse()}
	cache := NewPricingCache(mock, 1*time.Hour)

	if err := cache.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer cache.Stop()

	if cache.ModelCount() != 3 {
		t.Errorf("expected 3 models, got %d", cache.ModelCount())
	}
}

func TestPricingCache_GetPrice(t *testing.T) {
	mock := &mockModelLister{response: newTestModelsResponse()}
	cache := NewPricingCache(mock, 1*time.Hour)
	_ = cache.Start(context.Background())
	defer cache.Stop()

	price, ok := cache.GetPrice("openai/gpt-4o")
	if !ok {
		t.Fatal("expected to find openai/gpt-4o")
	}
	if price.PromptPrice != "2.5" {
		t.Errorf("expected prompt price '2.5', got %q", price.PromptPrice)
	}

	_, ok = cache.GetPrice("nonexistent/model")
	if ok {
		t.Error("expected model not to be found")
	}
}

func TestPricingCache_EstimateCost(t *testing.T) {
	mock := &mockModelLister{response: newTestModelsResponse()}
	cache := NewPricingCache(mock, 1*time.Hour)
	_ = cache.Start(context.Background())
	defer cache.Stop()

	// GPT-4o: prompt=$2.50/1M, completion=$10/1M
	// 1000 prompt tokens + 500 completion tokens
	// Expected: (1000 * 2.5 / 1_000_000) + (500 * 10 / 1_000_000) = 0.0025 + 0.005 = 0.0075
	cost := cache.EstimateCost("openai/gpt-4o", 1000, 500)
	if cost < 0.0074 || cost > 0.0076 {
		t.Errorf("expected cost ~0.0075, got %f", cost)
	}

	// Unknown model should return 0
	cost = cache.EstimateCost("unknown/model", 1000, 500)
	if cost != 0 {
		t.Errorf("expected 0 for unknown model, got %f", cost)
	}
}

func TestPricingCache_StartError(t *testing.T) {
	mock := &mockModelLister{
		err: context.DeadlineExceeded,
	}
	cache := NewPricingCache(mock, 1*time.Hour)

	err := cache.Start(context.Background())
	if err == nil {
		t.Fatal("expected error from Start")
	}
}

func TestPricingCache_ConcurrentAccess(t *testing.T) {
	mock := &mockModelLister{response: newTestModelsResponse()}
	cache := NewPricingCache(mock, 1*time.Hour)
	_ = cache.Start(context.Background())
	defer cache.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.EstimateCost("openai/gpt-4o", 1000, 500)
			cache.GetPrice("anthropic/claude-3.5-sonnet")
			cache.ModelCount()
		}()
	}
	wg.Wait()
}

package pricing

import (
	"context"
	"log/slog"
	"sync"
	"time"

	or "github.com/hra42/openrouter-go"
)

// ModelLister abstracts model listing for the pricing cache.
type ModelLister interface {
	ListModels(ctx context.Context) (*or.ModelsResponse, error)
}

// ModelPrice holds per-token pricing for a single model.
type ModelPrice struct {
	PromptPrice     string // raw pricing string from OpenRouter (cost per 1M tokens)
	CompletionPrice string // raw pricing string from OpenRouter (cost per 1M tokens)
}

// PricingCache caches model pricing data from OpenRouter and refreshes periodically.
type PricingCache struct {
	mu     sync.RWMutex
	prices map[string]ModelPrice
	client ModelLister
	ttl    time.Duration
	stopCh chan struct{}
}

// NewPricingCache creates a new pricing cache. Call Start to populate and begin background refresh.
func NewPricingCache(client ModelLister, ttl time.Duration) *PricingCache {
	return &PricingCache{
		prices: make(map[string]ModelPrice),
		client: client,
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
}

// Start loads initial pricing data and starts a background goroutine to refresh periodically.
func (pc *PricingCache) Start(ctx context.Context) error {
	if err := pc.refresh(ctx); err != nil {
		return err
	}
	go pc.backgroundRefresh()
	return nil
}

// Stop terminates the background refresh goroutine.
func (pc *PricingCache) Stop() {
	close(pc.stopCh)
}

// GetPrice returns the pricing for a model. The second return value is false if the model is not cached.
func (pc *PricingCache) GetPrice(modelID string) (ModelPrice, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	price, ok := pc.prices[modelID]
	return price, ok
}

// EstimateCost calculates the estimated cost for an API call using cached pricing.
func (pc *PricingCache) EstimateCost(modelID string, promptTokens, completionTokens int) float64 {
	price, ok := pc.GetPrice(modelID)
	if !ok {
		return 0
	}
	result := or.EstimateCostFromTokens(price.PromptPrice, price.CompletionPrice, promptTokens, completionTokens)
	return result.TotalCost
}

// ModelCount returns the number of cached model prices.
func (pc *PricingCache) ModelCount() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return len(pc.prices)
}

func (pc *PricingCache) refresh(ctx context.Context) error {
	resp, err := pc.client.ListModels(ctx)
	if err != nil {
		return err
	}

	newPrices := make(map[string]ModelPrice, len(resp.Data))
	for _, model := range resp.Data {
		newPrices[model.ID] = ModelPrice{
			PromptPrice:     model.Pricing.Prompt,
			CompletionPrice: model.Pricing.Completion,
		}
	}

	pc.mu.Lock()
	pc.prices = newPrices
	pc.mu.Unlock()

	slog.Debug("pricing cache refreshed", "models", len(newPrices))
	return nil
}

func (pc *PricingCache) backgroundRefresh() {
	ticker := time.NewTicker(pc.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-pc.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := pc.refresh(ctx); err != nil {
				slog.Warn("pricing cache refresh failed", "error", err)
			}
			cancel()
		}
	}
}

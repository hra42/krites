package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	or "github.com/hra42/openrouter-go"

	"github.com/hra42/krites/middleware"
	"github.com/hra42/krites/models"
)

// ChatCompleter abstracts chat completion operations for testability.
type ChatCompleter interface {
	ChatComplete(ctx context.Context, req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error)
	ChatCompleteStream(ctx context.Context, req *models.ChatCompletionRequest) (*or.ChatStream, error)
}

// ModelLister abstracts model listing for testability.
type ModelLister interface {
	ListModels(ctx context.Context) (*or.ModelsResponse, error)
}

// Client wraps the OpenRouter SDK client.
type Client struct {
	client *or.Client
	apiKey string
	debug  bool
}

// NewClient creates a new OpenRouter client with optional debug logging.
func NewClient(apiKey string, debug bool) *Client {
	return &Client{
		client: or.NewClient(or.WithAPIKey(apiKey)),
		apiKey: apiKey,
		debug:  debug,
	}
}

// ChatComplete sends a chat completion request to OpenRouter using a direct
// HTTP call so we can capture the byok_usage_inference field from the response.
func (c *Client) ChatComplete(ctx context.Context, req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
	if c.debug {
		log.Printf("DEBUG openrouter: ChatComplete model=%s messages=%d", req.Model, len(req.Messages))
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode >= 400 {
		return nil, fmt.Errorf("openrouter error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	var raw rawChatCompletionResponse
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if c.debug {
		log.Printf("DEBUG openrouter: ChatComplete response id=%s model=%s choices=%d cost=%f",
			raw.ID, raw.Model, len(raw.Choices), raw.Usage.Cost)
	}

	choices := make([]models.Choice, len(raw.Choices))
	for i, ch := range raw.Choices {
		choices[i] = models.Choice{
			Index: ch.Index,
			Message: models.ChatMessage{
				Role:    ch.Message.Role,
				Content: ch.Message.Content,
			},
			FinishReason: ch.FinishReason,
		}
	}

	result := &models.ChatCompletionResponse{
		ID:      raw.ID,
		Object:  raw.Object,
		Created: raw.Created,
		Model:   raw.Model,
		Choices: choices,
	}
	if raw.Usage.TotalTokens > 0 {
		result.Usage = &models.Usage{
			PromptTokens:     raw.Usage.PromptTokens,
			CompletionTokens: raw.Usage.CompletionTokens,
			TotalTokens:      raw.Usage.TotalTokens,
			Cost:             raw.Usage.Cost,
		}
	}

	return result, nil
}

// ChatCompleteStream sends a streaming chat completion request.
func (c *Client) ChatCompleteStream(ctx context.Context, req *models.ChatCompletionRequest) (*or.ChatStream, error) {
	messages := make([]or.Message, len(req.Messages))
	for i, msg := range req.Messages {
		switch msg.Role {
		case "system":
			messages[i] = or.CreateSystemMessage(msg.Content)
		case "assistant":
			messages[i] = or.CreateAssistantMessage(msg.Content)
		default:
			messages[i] = or.CreateUserMessage(msg.Content)
		}
	}

	opts := []or.ChatCompletionOption{
		or.WithModel(req.Model),
	}
	if req.Temperature != nil {
		opts = append(opts, or.WithTemperature(*req.Temperature))
	}
	if req.MaxTokens != nil {
		opts = append(opts, or.WithMaxTokens(*req.MaxTokens))
	}
	if req.TopP != nil {
		opts = append(opts, or.WithTopP(*req.TopP))
	}
	if len(req.Stop) > 0 {
		opts = append(opts, or.WithStop(req.Stop...))
	}

	if c.debug {
		log.Printf("DEBUG openrouter: ChatCompleteStream model=%s messages=%d", req.Model, len(req.Messages))
	}

	stream, err := c.client.ChatCompleteStream(ctx, messages, opts...)
	if err != nil {
		return nil, mapError(err)
	}
	return stream, nil
}

// ListModels retrieves available models from OpenRouter.
func (c *Client) ListModels(ctx context.Context) (*or.ModelsResponse, error) {
	resp, err := c.client.ListModels(ctx, nil)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// mapError translates SDK errors to AppError.
func mapError(err error) error {
	if reqErr, ok := or.IsRequestError(err); ok {
		if reqErr.IsRateLimitError() {
			return &middleware.AppError{
				Status:  429,
				Code:    middleware.CodeRateLimited,
				Message: fmt.Sprintf("rate limited: %s", reqErr.Message),
			}
		}
		if reqErr.IsAuthenticationError() {
			return &middleware.AppError{
				Status:  401,
				Code:    middleware.CodeOpenRouterError,
				Message: "invalid API key",
			}
		}
		return middleware.NewOpenRouterError(reqErr.Message, reqErr.StatusCode)
	}
	if valErr, ok := or.IsValidationError(err); ok {
		return middleware.NewInvalidRequestError(fmt.Sprintf("validation error: %s - %s", valErr.Field, valErr.Message))
	}
	return middleware.NewInternalError(fmt.Sprintf("openrouter error: %v", err))
}

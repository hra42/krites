package openrouter

// rawUsage mirrors OpenRouter's usage object including cost fields.
type rawUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	Cost             float64 `json:"cost,omitempty"`
	IsBYOK           bool    `json:"is_byok,omitempty"`
}

// rawChoice mirrors a single choice in the raw response.
type rawChoice struct {
	Index   int `json:"index"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

// rawChatCompletionResponse is the full OpenRouter response with BYOK fields.
type rawChatCompletionResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []rawChoice `json:"choices"`
	Usage   rawUsage    `json:"usage"`
}

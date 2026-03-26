package benchmark

import "time"

// CreateSuiteRequest is the JSON body for POST /benchmarks/suites.
type CreateSuiteRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Prompts     []Prompt  `json:"prompts"`
	Models      []string  `json:"models"`
	Config      *RunConfig `json:"config,omitempty"`
}

// UpdateSuiteRequest is the JSON body for PUT /benchmarks/suites/:id.
// All fields are optional for partial updates.
type UpdateSuiteRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	Prompts     []Prompt   `json:"prompts,omitempty"`
	Models      []string   `json:"models,omitempty"`
	Config      *RunConfig `json:"config,omitempty"`
}

// SuiteListResponse is the JSON response for listing suites (summary view).
type SuiteListResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ModelCount  int    `json:"model_count"`
	PromptCount int    `json:"prompt_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ToListResponse converts a Suite to a SuiteListResponse.
func (s *Suite) ToListResponse() SuiteListResponse {
	return SuiteListResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		ModelCount:  len(s.Models),
		PromptCount: len(s.Prompts),
		CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// RunListResponse is the JSON response for listing runs (summary view without results).
type RunListResponse struct {
	ID          string     `json:"id"`
	SuiteID     string     `json:"suite_id"`
	SuiteName   string     `json:"suite_name"`
	Status      RunStatus  `json:"status"`
	Config      RunConfig  `json:"config"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ResultCount int        `json:"result_count"`
}

// ToListResponse converts a Run to a RunListResponse.
func (r *Run) ToListResponse() RunListResponse {
	return RunListResponse{
		ID:          r.ID,
		SuiteID:     r.SuiteID,
		SuiteName:   r.SuiteName,
		Status:      r.Status,
		Config:      r.Config,
		StartedAt:   r.StartedAt,
		EndedAt:     r.EndedAt,
		CreatedAt:   r.CreatedAt,
		ResultCount: len(r.Results),
	}
}

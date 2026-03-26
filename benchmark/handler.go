package benchmark

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/hra42/krites/middleware"
	"github.com/hra42/krites/openrouter"
	"github.com/hra42/krites/pricing"
)

// Handler handles benchmark HTTP requests.
type Handler struct {
	store       Store
	runner      *Runner
	broadcaster *Broadcaster
}

// NewHandler creates a new benchmark Handler.
func NewHandler(store Store, orClient openrouter.ChatCompleter, pricingCache *pricing.PricingCache) *Handler {
	broadcaster := NewBroadcaster()
	runner := NewRunner(store, orClient, broadcaster, pricingCache)
	return &Handler{
		store:       store,
		runner:      runner,
		broadcaster: broadcaster,
	}
}

// HandleCreateSuite handles POST /benchmarks/suites.
func (h *Handler) HandleCreateSuite(c fiber.Ctx) error {
	var req CreateSuiteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return middleware.NewInvalidRequestError("invalid request body")
	}

	if req.Name == "" {
		return middleware.NewInvalidRequestError("name is required")
	}
	if len(req.Prompts) == 0 {
		return middleware.NewInvalidRequestError("at least one prompt is required")
	}
	if len(req.Models) == 0 {
		return middleware.NewInvalidRequestError("at least one model is required")
	}

	for i, p := range req.Prompts {
		if p.UserMessage == "" {
			return middleware.NewInvalidRequestError(
				fmt.Sprintf("prompt %d: user_message is required", i),
			)
		}
	}

	now := time.Now()
	suite := &Suite{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Models:      req.Models,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Assign IDs to prompts if missing
	suite.Prompts = make([]Prompt, len(req.Prompts))
	for i, p := range req.Prompts {
		if p.ID == "" {
			p.ID = uuid.New().String()
		}
		suite.Prompts[i] = p
	}

	// Apply config with defaults
	if req.Config != nil {
		suite.Config = *req.Config
	} else {
		suite.Config = DefaultRunConfig()
	}
	applyConfigDefaults(&suite.Config)

	if err := h.store.CreateSuite(suite); err != nil {
		return middleware.NewInternalError(err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(suite)
}

// HandleListSuites handles GET /benchmarks/suites.
func (h *Handler) HandleListSuites(c fiber.Ctx) error {
	suites, err := h.store.ListSuites()
	if err != nil {
		return middleware.NewInternalError(err.Error())
	}

	resp := make([]SuiteListResponse, len(suites))
	for i, s := range suites {
		resp[i] = s.ToListResponse()
	}

	return c.JSON(resp)
}

// HandleGetSuite handles GET /benchmarks/suites/:id.
func (h *Handler) HandleGetSuite(c fiber.Ctx) error {
	id := c.Params("id")

	suite, err := h.store.GetSuite(id)
	if err != nil {
		if errors.Is(err, ErrSuiteNotFound) {
			return &middleware.AppError{
				Status:  fiber.StatusNotFound,
				Code:    "SUITE_NOT_FOUND",
				Message: fmt.Sprintf("suite %q not found", id),
			}
		}
		return middleware.NewInternalError(err.Error())
	}

	return c.JSON(suite)
}

// HandleUpdateSuite handles PUT /benchmarks/suites/:id.
func (h *Handler) HandleUpdateSuite(c fiber.Ctx) error {
	id := c.Params("id")

	suite, err := h.store.GetSuite(id)
	if err != nil {
		if errors.Is(err, ErrSuiteNotFound) {
			return &middleware.AppError{
				Status:  fiber.StatusNotFound,
				Code:    "SUITE_NOT_FOUND",
				Message: fmt.Sprintf("suite %q not found", id),
			}
		}
		return middleware.NewInternalError(err.Error())
	}

	var req UpdateSuiteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return middleware.NewInvalidRequestError("invalid request body")
	}

	if req.Name != nil {
		suite.Name = *req.Name
	}
	if req.Description != nil {
		suite.Description = *req.Description
	}
	if req.Prompts != nil {
		for i := range req.Prompts {
			if req.Prompts[i].ID == "" {
				req.Prompts[i].ID = uuid.New().String()
			}
		}
		suite.Prompts = req.Prompts
	}
	if req.Models != nil {
		suite.Models = req.Models
	}
	if req.Config != nil {
		suite.Config = *req.Config
		applyConfigDefaults(&suite.Config)
	}

	suite.UpdatedAt = time.Now()

	if err := h.store.UpdateSuite(suite); err != nil {
		return middleware.NewInternalError(err.Error())
	}

	return c.JSON(suite)
}

// HandleDeleteSuite handles DELETE /benchmarks/suites/:id.
func (h *Handler) HandleDeleteSuite(c fiber.Ctx) error {
	id := c.Params("id")

	if err := h.store.DeleteSuite(id); err != nil {
		if errors.Is(err, ErrSuiteNotFound) {
			return &middleware.AppError{
				Status:  fiber.StatusNotFound,
				Code:    "SUITE_NOT_FOUND",
				Message: fmt.Sprintf("suite %q not found", id),
			}
		}
		return middleware.NewInternalError(err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleStartRun handles POST /benchmarks/suites/:id/run.
func (h *Handler) HandleStartRun(c fiber.Ctx) error {
	id := c.Params("id")

	suite, err := h.store.GetSuite(id)
	if err != nil {
		if errors.Is(err, ErrSuiteNotFound) {
			return &middleware.AppError{
				Status:  fiber.StatusNotFound,
				Code:    "SUITE_NOT_FOUND",
				Message: fmt.Sprintf("suite %q not found", id),
			}
		}
		return middleware.NewInternalError(err.Error())
	}

	run, err := h.runner.StartRun(suite)
	if err != nil {
		return middleware.NewInternalError(err.Error())
	}

	return c.Status(fiber.StatusAccepted).JSON(run)
}

// HandleListRuns handles GET /benchmarks/runs.
func (h *Handler) HandleListRuns(c fiber.Ctx) error {
	runs, err := h.store.ListRuns()
	if err != nil {
		return middleware.NewInternalError(err.Error())
	}

	resp := make([]RunListResponse, len(runs))
	for i, r := range runs {
		resp[i] = r.ToListResponse()
	}

	return c.JSON(resp)
}

// HandleGetRun handles GET /benchmarks/runs/:id.
func (h *Handler) HandleGetRun(c fiber.Ctx) error {
	id := c.Params("id")

	run, err := h.store.GetRun(id)
	if err != nil {
		if errors.Is(err, ErrRunNotFound) {
			return &middleware.AppError{
				Status:  fiber.StatusNotFound,
				Code:    "RUN_NOT_FOUND",
				Message: fmt.Sprintf("run %q not found", id),
			}
		}
		return middleware.NewInternalError(err.Error())
	}

	return c.JSON(run)
}

// HandleStreamRun handles GET /benchmarks/runs/:id/stream (SSE).
func (h *Handler) HandleStreamRun(c fiber.Ctx) error {
	id := c.Params("id")

	// Verify run exists
	_, err := h.store.GetRun(id)
	if err != nil {
		if errors.Is(err, ErrRunNotFound) {
			return &middleware.AppError{
				Status:  fiber.StatusNotFound,
				Code:    "RUN_NOT_FOUND",
				Message: fmt.Sprintf("run %q not found", id),
			}
		}
		return middleware.NewInternalError(err.Error())
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	events, unsubscribe := h.broadcaster.Subscribe(id)

	return c.SendStreamWriter(func(w *bufio.Writer) {
		defer unsubscribe()
		for event := range events {
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			_ = w.Flush()
		}
	})
}

// HandleExportRun handles GET /benchmarks/runs/:id/export.
func (h *Handler) HandleExportRun(c fiber.Ctx) error {
	id := c.Params("id")
	format := c.Query("format", "json")

	run, err := h.store.GetRun(id)
	if err != nil {
		if errors.Is(err, ErrRunNotFound) {
			return &middleware.AppError{
				Status:  fiber.StatusNotFound,
				Code:    "RUN_NOT_FOUND",
				Message: fmt.Sprintf("run %q not found", id),
			}
		}
		return middleware.NewInternalError(err.Error())
	}

	switch format {
	case "csv":
		return h.exportCSV(c, run)
	default:
		return h.exportJSON(c, run)
	}
}

func (h *Handler) exportJSON(c fiber.Ctx, run *Run) error {
	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="run-%s.json"`, run.ID))
	return c.JSON(run)
}

func (h *Handler) exportCSV(c fiber.Ctx, run *Run) error {
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="run-%s.csv"`, run.ID))

	// Collect all unique judge criteria
	criteriaSet := make(map[string]struct{})
	for _, r := range run.Results {
		for _, js := range r.JudgeScores {
			criteriaSet[js.Criterion] = struct{}{}
		}
	}
	var criteria []string
	for c := range criteriaSet {
		criteria = append(criteria, c)
	}

	// Build header
	header := []string{
		"result_id", "model", "prompt_id", "prompt_name", "iteration",
		"status", "error", "ttfb_ms", "total_latency_ms",
		"prompt_tokens", "completion_tokens", "tokens_per_second",
		"estimated_cost", "response",
	}
	for _, cr := range criteria {
		header = append(header, "judge_"+cr)
	}

	buf := c.Response().BodyWriter()
	w := csv.NewWriter(buf)
	if err := w.Write(header); err != nil {
		return err
	}

	for _, r := range run.Results {
		row := []string{
			r.ID, r.Model, r.PromptID, r.PromptName,
			strconv.Itoa(r.Iteration),
			string(r.Status), r.Error,
			strconv.FormatFloat(r.Metrics.TTFB, 'f', 2, 64),
			strconv.FormatFloat(r.Metrics.TotalLatency, 'f', 2, 64),
			strconv.Itoa(r.Metrics.PromptTokens),
			strconv.Itoa(r.Metrics.CompletionTokens),
			strconv.FormatFloat(r.Metrics.TokensPerSecond, 'f', 2, 64),
			strconv.FormatFloat(r.Metrics.EstimatedCost, 'f', 8, 64),
			r.Response,
		}

		// Build a map of criterion->score for this result
		scoreMap := make(map[string]float64)
		for _, js := range r.JudgeScores {
			scoreMap[js.Criterion] = js.Score
		}
		for _, cr := range criteria {
			if score, ok := scoreMap[cr]; ok {
				row = append(row, strconv.FormatFloat(score, 'f', 1, 64))
			} else {
				row = append(row, "")
			}
		}

		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}

// HandleAnalyticsOverview handles GET /benchmarks/analytics/overview.
func (h *Handler) HandleAnalyticsOverview(c fiber.Ctx) error {
	as, ok := h.store.(AnalyticsStore)
	if !ok {
		return &middleware.AppError{
			Status:  fiber.StatusNotImplemented,
			Code:    "NOT_IMPLEMENTED",
			Message: "analytics not available with current storage backend",
		}
	}

	overview, err := as.GetOverview()
	if err != nil {
		return middleware.NewInternalError(err.Error())
	}
	return c.JSON(overview)
}

// HandleAnalyticsModelComparison handles GET /benchmarks/analytics/models.
func (h *Handler) HandleAnalyticsModelComparison(c fiber.Ctx) error {
	as, ok := h.store.(AnalyticsStore)
	if !ok {
		return &middleware.AppError{
			Status:  fiber.StatusNotImplemented,
			Code:    "NOT_IMPLEMENTED",
			Message: "analytics not available with current storage backend",
		}
	}

	suiteID := c.Query("suite_id")
	stats, err := as.GetCrossRunComparison(suiteID)
	if err != nil {
		return middleware.NewInternalError(err.Error())
	}
	if stats == nil {
		stats = []CrossRunModelStats{}
	}
	return c.JSON(stats)
}

// HandleAnalyticsTrends handles GET /benchmarks/analytics/trends.
func (h *Handler) HandleAnalyticsTrends(c fiber.Ctx) error {
	as, ok := h.store.(AnalyticsStore)
	if !ok {
		return &middleware.AppError{
			Status:  fiber.StatusNotImplemented,
			Code:    "NOT_IMPLEMENTED",
			Message: "analytics not available with current storage backend",
		}
	}

	modelID := c.Query("model")
	if modelID == "" {
		return middleware.NewInvalidRequestError("model query parameter is required")
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	trends, err := as.GetModelTrends(modelID, limit)
	if err != nil {
		return middleware.NewInternalError(err.Error())
	}
	if trends == nil {
		trends = []ModelTrendPoint{}
	}
	return c.JSON(trends)
}

// RegisterRoutes registers all benchmark routes on a Fiber app.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	benchmarks := app.Group("/benchmarks")

	suites := benchmarks.Group("/suites")
	suites.Post("/", h.HandleCreateSuite)
	suites.Get("/", h.HandleListSuites)
	suites.Get("/:id", h.HandleGetSuite)
	suites.Put("/:id", h.HandleUpdateSuite)
	suites.Delete("/:id", h.HandleDeleteSuite)
	suites.Post("/:id/run", h.HandleStartRun)

	runs := benchmarks.Group("/runs")
	runs.Get("/", h.HandleListRuns)
	runs.Get("/:id", h.HandleGetRun)
	runs.Get("/:id/stream", h.HandleStreamRun)
	runs.Get("/:id/export", h.HandleExportRun)

	analytics := benchmarks.Group("/analytics")
	analytics.Get("/overview", h.HandleAnalyticsOverview)
	analytics.Get("/models", h.HandleAnalyticsModelComparison)
	analytics.Get("/trends", h.HandleAnalyticsTrends)
}

// applyConfigDefaults fills in zero-valued config fields with defaults.
func applyConfigDefaults(cfg *RunConfig) {
	defaults := DefaultRunConfig()
	if cfg.Temperature == 0 {
		cfg.Temperature = defaults.Temperature
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = defaults.MaxTokens
	}
	if cfg.TopP == 0 {
		cfg.TopP = defaults.TopP
	}
	if cfg.Iterations == 0 {
		cfg.Iterations = defaults.Iterations
	}
	if cfg.Concurrency == 0 {
		cfg.Concurrency = defaults.Concurrency
	}
	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = defaults.TimeoutSeconds
	}
}

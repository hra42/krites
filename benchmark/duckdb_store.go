package benchmark

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hra42/krites/database"
)

// DuckDBStore implements Store and AnalyticsStore backed by DuckDB.
type DuckDBStore struct {
	db *sql.DB
	mu sync.Mutex // serializes writes (DuckDB supports only one concurrent writer)
}

// NewDuckDBStore creates a new DuckDB-backed store and initializes the schema.
func NewDuckDBStore(db *sql.DB) (*DuckDBStore, error) {
	if err := database.InitializeBenchmarkSchema(db); err != nil {
		return nil, fmt.Errorf("initializing benchmark schema: %w", err)
	}
	return &DuckDBStore{db: db}, nil
}

// --- Suite CRUD ---

func (s *DuckDBStore) CreateSuite(suite *Suite) error {
	prompts, err := json.Marshal(suite.Prompts)
	if err != nil {
		return fmt.Errorf("marshaling prompts: %w", err)
	}
	models, err := json.Marshal(suite.Models)
	if err != nil {
		return fmt.Errorf("marshaling models: %w", err)
	}
	config, err := json.Marshal(suite.Config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = s.db.Exec(
		`INSERT INTO benchmark_suites (id, name, description, prompts, models, config, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		suite.ID, suite.Name, suite.Description,
		string(prompts), string(models), string(config),
		suite.CreatedAt, suite.UpdatedAt,
	)
	return err
}

func (s *DuckDBStore) GetSuite(id string) (*Suite, error) {
	row := s.db.QueryRow(
		`SELECT id, name, description, prompts, models, config, created_at, updated_at
		 FROM benchmark_suites WHERE id = ?`, id,
	)
	return s.scanSuite(row)
}

func (s *DuckDBStore) ListSuites() ([]*Suite, error) {
	rows, err := s.db.Query(
		`SELECT id, name, description, prompts, models, config, created_at, updated_at
		 FROM benchmark_suites ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suites []*Suite
	for rows.Next() {
		suite, err := s.scanSuiteRow(rows)
		if err != nil {
			return nil, err
		}
		suites = append(suites, suite)
	}
	return suites, rows.Err()
}

func (s *DuckDBStore) UpdateSuite(suite *Suite) error {
	// Check existence first
	if _, err := s.GetSuite(suite.ID); err != nil {
		return err
	}

	prompts, err := json.Marshal(suite.Prompts)
	if err != nil {
		return fmt.Errorf("marshaling prompts: %w", err)
	}
	models, err := json.Marshal(suite.Models)
	if err != nil {
		return fmt.Errorf("marshaling models: %w", err)
	}
	config, err := json.Marshal(suite.Config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = s.db.Exec(
		`UPDATE benchmark_suites SET name=?, description=?, prompts=?, models=?, config=?, updated_at=?
		 WHERE id=?`,
		suite.Name, suite.Description,
		string(prompts), string(models), string(config),
		suite.UpdatedAt, suite.ID,
	)
	return err
}

func (s *DuckDBStore) DeleteSuite(id string) error {
	if _, err := s.GetSuite(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`DELETE FROM benchmark_suites WHERE id = ?`, id)
	return err
}

// --- Run CRUD ---

func (s *DuckDBStore) CreateRun(run *Run) error {
	config, err := json.Marshal(run.Config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = s.db.Exec(
		`INSERT INTO benchmark_runs (id, suite_id, suite_name, status, config, started_at, ended_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.SuiteID, run.SuiteName, string(run.Status),
		string(config), run.StartedAt, run.EndedAt, run.CreatedAt,
	)
	return err
}

func (s *DuckDBStore) GetRun(id string) (*Run, error) {
	// Load run metadata
	var run Run
	var configJSON string
	var startedAt, endedAt sql.NullTime
	var status string

	err := s.db.QueryRow(
		`SELECT id, suite_id, suite_name, status, config, started_at, ended_at, created_at
		 FROM benchmark_runs WHERE id = ?`, id,
	).Scan(&run.ID, &run.SuiteID, &run.SuiteName, &status, &configJSON,
		&startedAt, &endedAt, &run.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRunNotFound
		}
		return nil, err
	}

	run.Status = RunStatus(status)
	if startedAt.Valid {
		run.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		run.EndedAt = &endedAt.Time
	}
	if err := json.Unmarshal([]byte(configJSON), &run.Config); err != nil {
		return nil, fmt.Errorf("unmarshaling run config: %w", err)
	}

	// Load results
	results, err := s.loadResults(id)
	if err != nil {
		return nil, err
	}
	run.Results = results

	// Compute summary from results
	if len(results) > 0 && (run.Status == RunStatusComplete || run.Status == RunStatusFailed) {
		run.Summary = computeSummary(results, run.Config.JudgeCriteria)
	}

	return &run, nil
}

func (s *DuckDBStore) ListRuns() ([]*Run, error) {
	rows, err := s.db.Query(
		`SELECT r.id, r.suite_id, r.suite_name, r.status, r.config,
		        r.started_at, r.ended_at, r.created_at,
		        (SELECT COUNT(*) FROM benchmark_results WHERE run_id = r.id) as result_count
		 FROM benchmark_runs r ORDER BY r.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*Run
	for rows.Next() {
		var run Run
		var configJSON string
		var startedAt, endedAt sql.NullTime
		var status string
		var resultCount int

		if err := rows.Scan(&run.ID, &run.SuiteID, &run.SuiteName, &status, &configJSON,
			&startedAt, &endedAt, &run.CreatedAt, &resultCount); err != nil {
			return nil, err
		}

		run.Status = RunStatus(status)
		if startedAt.Valid {
			run.StartedAt = &startedAt.Time
		}
		if endedAt.Valid {
			run.EndedAt = &endedAt.Time
		}
		if err := json.Unmarshal([]byte(configJSON), &run.Config); err != nil {
			return nil, fmt.Errorf("unmarshaling run config: %w", err)
		}

		runs = append(runs, &run)
	}
	return runs, rows.Err()
}

func (s *DuckDBStore) UpdateRun(run *Run) error {
	config, err := json.Marshal(run.Config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update run metadata
	_, err = tx.Exec(
		`UPDATE benchmark_runs SET status=?, config=?, started_at=?, ended_at=? WHERE id=?`,
		string(run.Status), string(config), run.StartedAt, run.EndedAt, run.ID,
	)
	if err != nil {
		return err
	}

	// Upsert results (INSERT OR IGNORE since IDs are UUIDs)
	for _, result := range run.Results {
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO benchmark_results
			 (id, run_id, prompt_id, prompt_name, model, iteration, response, status, error,
			  ttfb_ms, total_latency_ms, prompt_tokens, completion_tokens, tokens_per_second, estimated_cost)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			result.ID, result.RunID, result.PromptID, result.PromptName,
			result.Model, result.Iteration, result.Response, string(result.Status), result.Error,
			result.Metrics.TTFB, result.Metrics.TotalLatency,
			result.Metrics.PromptTokens, result.Metrics.CompletionTokens,
			result.Metrics.TokensPerSecond, result.Metrics.EstimatedCost,
		)
		if err != nil {
			return err
		}

		// Upsert judge scores
		for _, score := range result.JudgeScores {
			_, err = tx.Exec(
				`INSERT OR REPLACE INTO benchmark_judge_scores (result_id, criterion, score, explanation)
				 VALUES (?, ?, ?, ?)`,
				result.ID, score.Criterion, score.Score, score.Explanation,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// --- Analytics ---

func (s *DuckDBStore) GetOverview() (*AnalyticsOverview, error) {
	var overview AnalyticsOverview

	err := s.db.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM benchmark_runs),
			(SELECT COUNT(*) FROM benchmark_runs WHERE status = 'complete'),
			(SELECT COUNT(*) FROM benchmark_results),
			(SELECT COUNT(DISTINCT model) FROM benchmark_results),
			(SELECT COALESCE(SUM(estimated_cost), 0) FROM benchmark_results),
			(SELECT COALESCE(AVG(total_latency_ms), 0) FROM benchmark_results WHERE status = 'success')
	`).Scan(&overview.TotalRuns, &overview.CompletedRuns, &overview.TotalResults,
		&overview.DistinctModels, &overview.TotalCost, &overview.AvgLatency)
	if err != nil {
		return nil, err
	}
	return &overview, nil
}

func (s *DuckDBStore) GetModelTrends(modelID string, limit int) ([]ModelTrendPoint, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT
			r.id AS run_id,
			r.suite_name,
			r.created_at,
			AVG(res.total_latency_ms) AS avg_latency,
			AVG(res.ttfb_ms) AS avg_ttfb,
			AVG(res.tokens_per_second) AS avg_tps,
			AVG(res.estimated_cost) AS avg_cost,
			SUM(CASE WHEN res.status = 'success' THEN 1 ELSE 0 END)::DOUBLE / COUNT(*) AS success_rate
		FROM benchmark_results res
		JOIN benchmark_runs r ON res.run_id = r.id
		WHERE res.model = ? AND r.status = 'complete'
		GROUP BY r.id, r.suite_name, r.created_at
		ORDER BY r.created_at DESC
		LIMIT ?
	`, modelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []ModelTrendPoint
	for rows.Next() {
		var pt ModelTrendPoint
		if err := rows.Scan(&pt.RunID, &pt.SuiteName, &pt.CreatedAt,
			&pt.AvgLatency, &pt.AvgTTFB, &pt.AvgTPS, &pt.AvgCost, &pt.SuccessRate); err != nil {
			return nil, err
		}
		trends = append(trends, pt)
	}
	return trends, rows.Err()
}

func (s *DuckDBStore) GetCrossRunComparison(suiteID string) ([]CrossRunModelStats, error) {
	query := `
		SELECT
			res.model,
			COUNT(DISTINCT res.run_id) AS run_count,
			AVG(res.total_latency_ms) AS avg_latency,
			AVG(res.ttfb_ms) AS avg_ttfb,
			AVG(res.tokens_per_second) AS avg_tps,
			SUM(res.estimated_cost) AS total_cost,
			SUM(CASE WHEN res.status = 'success' THEN 1 ELSE 0 END)::DOUBLE / COUNT(*) AS avg_success_rate
		FROM benchmark_results res
		JOIN benchmark_runs r ON res.run_id = r.id
		WHERE r.status = 'complete'
	`
	var args []interface{}
	if suiteID != "" {
		query += ` AND r.suite_id = ?`
		args = append(args, suiteID)
	}
	query += ` GROUP BY res.model ORDER BY avg_latency ASC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []CrossRunModelStats
	for rows.Next() {
		var st CrossRunModelStats
		if err := rows.Scan(&st.Model, &st.RunCount, &st.AvgLatency, &st.AvgTTFB,
			&st.AvgTPS, &st.TotalCost, &st.AvgSuccessRate); err != nil {
			return nil, err
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

// --- internal helpers ---

func (s *DuckDBStore) loadResults(runID string) ([]Result, error) {
	rows, err := s.db.Query(
		`SELECT id, run_id, prompt_id, prompt_name, model, iteration, response, status, error,
		        ttfb_ms, total_latency_ms, prompt_tokens, completion_tokens, tokens_per_second, estimated_cost
		 FROM benchmark_results WHERE run_id = ? ORDER BY model, iteration`, runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var r Result
		var status string
		if err := rows.Scan(&r.ID, &r.RunID, &r.PromptID, &r.PromptName,
			&r.Model, &r.Iteration, &r.Response, &status, &r.Error,
			&r.Metrics.TTFB, &r.Metrics.TotalLatency,
			&r.Metrics.PromptTokens, &r.Metrics.CompletionTokens,
			&r.Metrics.TokensPerSecond, &r.Metrics.EstimatedCost); err != nil {
			return nil, err
		}
		r.Status = ResultStatus(status)
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load judge scores for all results
	resultIDs := make([]string, len(results))
	resultMap := make(map[string]*Result, len(results))
	for i := range results {
		resultIDs[i] = results[i].ID
		resultMap[results[i].ID] = &results[i]
	}

	if len(resultIDs) > 0 {
		scoreRows, err := s.db.Query(
			`SELECT result_id, criterion, score, explanation
			 FROM benchmark_judge_scores WHERE result_id IN (SELECT id FROM benchmark_results WHERE run_id = ?)`,
			runID,
		)
		if err != nil {
			return nil, err
		}
		defer scoreRows.Close()

		for scoreRows.Next() {
			var resultID, criterion, explanation string
			var score float64
			if err := scoreRows.Scan(&resultID, &criterion, &score, &explanation); err != nil {
				return nil, err
			}
			if r, ok := resultMap[resultID]; ok {
				r.JudgeScores = append(r.JudgeScores, JudgeScore{
					Criterion:   criterion,
					Score:       score,
					Explanation: explanation,
				})
			}
		}
		if err := scoreRows.Err(); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (s *DuckDBStore) scanSuite(row *sql.Row) (*Suite, error) {
	var suite Suite
	var promptsJSON, modelsJSON, configJSON string

	err := row.Scan(&suite.ID, &suite.Name, &suite.Description,
		&promptsJSON, &modelsJSON, &configJSON,
		&suite.CreatedAt, &suite.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSuiteNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(promptsJSON), &suite.Prompts); err != nil {
		return nil, fmt.Errorf("unmarshaling prompts: %w", err)
	}
	if err := json.Unmarshal([]byte(modelsJSON), &suite.Models); err != nil {
		return nil, fmt.Errorf("unmarshaling models: %w", err)
	}
	if err := json.Unmarshal([]byte(configJSON), &suite.Config); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &suite, nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (s *DuckDBStore) scanSuiteRow(row rowScanner) (*Suite, error) {
	var suite Suite
	var promptsJSON, modelsJSON, configJSON string

	err := row.Scan(&suite.ID, &suite.Name, &suite.Description,
		&promptsJSON, &modelsJSON, &configJSON,
		&suite.CreatedAt, &suite.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(promptsJSON), &suite.Prompts); err != nil {
		return nil, fmt.Errorf("unmarshaling prompts: %w", err)
	}
	if err := json.Unmarshal([]byte(modelsJSON), &suite.Models); err != nil {
		return nil, fmt.Errorf("unmarshaling models: %w", err)
	}
	if err := json.Unmarshal([]byte(configJSON), &suite.Config); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &suite, nil
}

// Ensure DuckDBStore implements both Store and AnalyticsStore.
var _ Store = (*DuckDBStore)(nil)
var _ AnalyticsStore = (*DuckDBStore)(nil)

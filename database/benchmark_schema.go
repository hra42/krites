package database

import "database/sql"

const BenchmarkSchema = `
CREATE TABLE IF NOT EXISTS benchmark_suites (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    prompts TEXT NOT NULL,
    models TEXT NOT NULL,
    config TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS benchmark_runs (
    id TEXT PRIMARY KEY,
    suite_id TEXT NOT NULL,
    suite_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    config TEXT NOT NULL,
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_runs_suite_id ON benchmark_runs(suite_id);
CREATE INDEX IF NOT EXISTS idx_runs_status ON benchmark_runs(status);
CREATE INDEX IF NOT EXISTS idx_runs_created_at ON benchmark_runs(created_at);

CREATE TABLE IF NOT EXISTS benchmark_results (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    prompt_id TEXT NOT NULL,
    prompt_name TEXT DEFAULT '',
    model TEXT NOT NULL,
    iteration INTEGER NOT NULL,
    response TEXT DEFAULT '',
    status TEXT NOT NULL,
    error TEXT DEFAULT '',
    ttfb_ms DOUBLE DEFAULT 0,
    total_latency_ms DOUBLE DEFAULT 0,
    prompt_tokens INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    tokens_per_second DOUBLE DEFAULT 0,
    estimated_cost DOUBLE DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_results_run_id ON benchmark_results(run_id);
CREATE INDEX IF NOT EXISTS idx_results_model ON benchmark_results(model);

CREATE TABLE IF NOT EXISTS benchmark_judge_scores (
    result_id TEXT NOT NULL,
    criterion TEXT NOT NULL,
    score DOUBLE NOT NULL,
    explanation TEXT DEFAULT '',
    PRIMARY KEY (result_id, criterion)
);
CREATE INDEX IF NOT EXISTS idx_judge_scores_result_id ON benchmark_judge_scores(result_id);
`

func InitializeBenchmarkSchema(db *sql.DB) error {
	_, err := db.Exec(BenchmarkSchema)
	return err
}

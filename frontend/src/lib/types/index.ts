export interface Prompt {
	id: string;
	name: string;
	system_message: string;
	user_message: string;
	expected_output?: string;
	category?: string;
}

export interface RunConfig {
	temperature: number;
	max_tokens: number;
	top_p: number;
	iterations: number;
	concurrency: number;
	judge_enabled: boolean;
	judge_model?: string;
	judge_criteria?: string[];
	timeout_seconds: number;
}

export interface Suite {
	id: string;
	name: string;
	description: string;
	prompts: Prompt[];
	models: string[];
	config: RunConfig;
	created_at: string;
	updated_at: string;
}

export interface SuiteSummary {
	id: string;
	name: string;
	description: string;
	model_count: number;
	prompt_count: number;
	created_at: string;
	updated_at: string;
}

export interface CreateSuiteRequest {
	name: string;
	description: string;
	prompts: Omit<Prompt, 'id'>[];
	models: string[];
	config?: Partial<RunConfig>;
}

export interface UpdateSuiteRequest {
	name?: string;
	description?: string;
	prompts?: Prompt[];
	models?: string[];
	config?: Partial<RunConfig>;
}

export type RunStatus = 'pending' | 'running' | 'complete' | 'failed' | 'canceled';
export type ResultStatus = 'success' | 'error' | 'timeout';

export interface ResultMetrics {
	ttfb_ms: number;
	total_latency_ms: number;
	prompt_tokens: number;
	completion_tokens: number;
	tokens_per_second: number;
	estimated_cost: number;
}

export interface JudgeScore {
	criterion: string;
	score: number;
	explanation: string;
}

export interface Result {
	id: string;
	run_id: string;
	prompt_id: string;
	prompt_name: string;
	model: string;
	iteration: number;
	response: string;
	metrics: ResultMetrics;
	judge_scores?: JudgeScore[];
	status: ResultStatus;
	error?: string;
}

export interface ModelSummary {
	model: string;
	avg_ttfb_ms: number;
	p50_ttfb_ms: number;
	p95_ttfb_ms: number;
	avg_latency_ms: number;
	p50_latency_ms: number;
	p95_latency_ms: number;
	avg_tokens_per_second: number;
	avg_cost: number;
	total_cost: number;
	success_rate: number;
	avg_judge_scores?: Record<string, number>;
}

export interface Summary {
	models: ModelSummary[];
}

export interface Run {
	id: string;
	suite_id: string;
	suite_name: string;
	status: RunStatus;
	results?: Result[];
	summary?: Summary;
	config: RunConfig;
	started_at?: string;
	ended_at?: string;
	created_at: string;
}

export interface RunSummary {
	id: string;
	suite_id: string;
	suite_name: string;
	status: RunStatus;
	config: RunConfig;
	started_at?: string;
	ended_at?: string;
	created_at: string;
	result_count: number;
}

export type SSEEventType = 'run_started' | 'result_completed' | 'judge_scored' | 'run_completed' | 'run_error';

export interface SSEEvent {
	type: SSEEventType;
	data: unknown;
}

export interface OpenRouterModel {
	id: string;
	name: string;
	description: string;
	context_length?: number | null;
	pricing: {
		prompt: string;
		completion: string;
	};
	architecture: {
		input_modalities: string[];
		output_modalities: string[];
		tokenizer: string;
	};
}

export interface OpenRouterModelsResponse {
	data: OpenRouterModel[];
}

export interface APIError {
	error: {
		message: string;
		code: string;
		status: number;
	};
}

export interface AnalyticsOverview {
	total_runs: number;
	completed_runs: number;
	total_results: number;
	distinct_models: number;
	total_cost: number;
	avg_latency_ms: number;
}

export interface ModelTrendPoint {
	run_id: string;
	suite_name: string;
	created_at: string;
	avg_latency_ms: number;
	avg_ttfb_ms: number;
	avg_tokens_per_second: number;
	avg_cost: number;
	success_rate: number;
}

export interface CrossRunModelStats {
	model: string;
	run_count: number;
	avg_latency_ms: number;
	avg_ttfb_ms: number;
	avg_tokens_per_second: number;
	total_cost: number;
	avg_success_rate: number;
}

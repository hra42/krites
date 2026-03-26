import type { Suite, SuiteSummary, CreateSuiteRequest, UpdateSuiteRequest, Run, RunSummary, AnalyticsOverview, CrossRunModelStats, ModelTrendPoint, OpenRouterModelsResponse } from '$lib/types';

class ApiError extends Error {
	constructor(
		public status: number,
		public code: string,
		message: string
	) {
		super(message);
	}
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
	const res = await fetch(path, {
		headers: { 'Content-Type': 'application/json' },
		...options
	});

	if (!res.ok) {
		const body = await res.json().catch(() => null);
		throw new ApiError(
			res.status,
			body?.error?.code ?? 'UNKNOWN',
			body?.error?.message ?? res.statusText
		);
	}

	if (res.status === 204) return undefined as T;
	return res.json();
}

export function listSuites(): Promise<SuiteSummary[]> {
	return request('/benchmarks/suites/');
}

export function getSuite(id: string): Promise<Suite> {
	return request(`/benchmarks/suites/${id}`);
}

export function createSuite(data: CreateSuiteRequest): Promise<Suite> {
	return request('/benchmarks/suites/', {
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export function updateSuite(id: string, data: UpdateSuiteRequest): Promise<Suite> {
	return request(`/benchmarks/suites/${id}`, {
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export function deleteSuite(id: string): Promise<void> {
	return request(`/benchmarks/suites/${id}`, { method: 'DELETE' });
}

export function listRuns(): Promise<RunSummary[]> {
	return request('/benchmarks/runs/');
}

export function getRun(id: string): Promise<Run> {
	return request(`/benchmarks/runs/${id}`);
}

export function startRun(suiteId: string): Promise<Run> {
	return request(`/benchmarks/suites/${suiteId}/run`, { method: 'POST' });
}

export function streamRun(
	runId: string,
	onEvent: (type: string, data: unknown) => void,
	onError: (error: Event) => void
): () => void {
	const url = `/benchmarks/runs/${runId}/stream`;
	const eventSource = new EventSource(url);

	const eventTypes = ['run_started', 'result_completed', 'judge_scored', 'run_completed', 'run_error'];
	for (const type of eventTypes) {
		eventSource.addEventListener(type, (e: MessageEvent) => {
			try {
				const parsed = JSON.parse(e.data);
				onEvent(type, parsed.data);
			} catch {
				onEvent(type, e.data);
			}
		});
	}

	eventSource.onerror = onError;

	return () => {
		eventSource.close();
	};
}

export function exportRunUrl(runId: string, format: 'csv' | 'json'): string {
	return `/benchmarks/runs/${runId}/export?format=${format}`;
}

export function getAnalyticsOverview(): Promise<AnalyticsOverview> {
	return request('/benchmarks/analytics/overview');
}

export function getModelComparison(suiteId?: string): Promise<CrossRunModelStats[]> {
	const params = suiteId ? `?suite_id=${suiteId}` : '';
	return request(`/benchmarks/analytics/models${params}`);
}

export function getModelTrends(modelId: string, limit = 20): Promise<ModelTrendPoint[]> {
	return request(`/benchmarks/analytics/trends?model=${encodeURIComponent(modelId)}&limit=${limit}`);
}

export function listModels(): Promise<OpenRouterModelsResponse> {
	return request('/v1/models');
}

export { ApiError };

import { writable, derived } from 'svelte/store';
import type { Run, Result } from '$lib/types';

export const currentRun = writable<Run | null>(null);
export const liveResults = writable<Result[]>([]);
export const isRunning = writable(false);
export const runError = writable<string | null>(null);
export const totalCalls = writable(0);
export const judgeProgress = writable<{ scored: number; total: number }>({ scored: 0, total: 0 });

export const resultsByModel = derived(liveResults, ($results) => {
	const grouped: Record<string, Result[]> = {};
	for (const result of $results) {
		if (!grouped[result.model]) {
			grouped[result.model] = [];
		}
		grouped[result.model].push(result);
	}
	return grouped;
});

export const runProgress = derived(
	[liveResults, totalCalls],
	([$results, $total]) => ({
		completed: $results.length,
		total: $total,
		percent: $total > 0 ? Math.round(($results.length / $total) * 100) : 0
	})
);

export const modelAggregates = derived(liveResults, ($results) => {
	const aggregates: Record<string, {
		avgLatency: number;
		avgTokensPerSecond: number;
		successCount: number;
		errorCount: number;
		totalCount: number;
	}> = {};

	for (const result of $results) {
		if (!aggregates[result.model]) {
			aggregates[result.model] = {
				avgLatency: 0,
				avgTokensPerSecond: 0,
				successCount: 0,
				errorCount: 0,
				totalCount: 0
			};
		}
		const agg = aggregates[result.model];
		agg.totalCount++;
		if (result.status === 'success') {
			agg.successCount++;
			agg.avgLatency =
				(agg.avgLatency * (agg.successCount - 1) + result.metrics.total_latency_ms) /
				agg.successCount;
			agg.avgTokensPerSecond =
				(agg.avgTokensPerSecond * (agg.successCount - 1) + result.metrics.tokens_per_second) /
				agg.successCount;
		} else {
			agg.errorCount++;
		}
	}

	return aggregates;
});

export const modelJudgeAggregates = derived(liveResults, ($results) => {
	const aggregates: Record<string, Record<string, { sum: number; count: number; avg: number }>> = {};
	for (const result of $results) {
		if (!result.judge_scores?.length) continue;
		if (!aggregates[result.model]) aggregates[result.model] = {};
		for (const score of result.judge_scores) {
			if (!aggregates[result.model][score.criterion]) {
				aggregates[result.model][score.criterion] = { sum: 0, count: 0, avg: 0 };
			}
			const agg = aggregates[result.model][score.criterion];
			agg.sum += score.score;
			agg.count++;
			agg.avg = agg.sum / agg.count;
		}
	}
	return aggregates;
});

export function handleStreamEvent(type: string, data: unknown) {
	switch (type) {
		case 'run_started':
			currentRun.set(data as Run);
			isRunning.set(true);
			break;
		case 'result_completed':
			liveResults.update((results) => [...results, data as Result]);
			break;
		case 'judge_scored': {
			const scoredResult = data as Result;
			liveResults.update((results) =>
				results.map((r) =>
					r.id === scoredResult.id ? { ...r, judge_scores: scoredResult.judge_scores } : r
				)
			);
			judgeProgress.update((p) => ({ ...p, scored: p.scored + 1 }));
			break;
		}
		case 'run_completed':
			currentRun.set(data as Run);
			isRunning.set(false);
			break;
		case 'run_error':
			currentRun.set(data as Run);
			isRunning.set(false);
			runError.set('Run failed');
			break;
	}
}

export function resetRunState() {
	currentRun.set(null);
	liveResults.set([]);
	isRunning.set(false);
	runError.set(null);
	totalCalls.set(0);
	judgeProgress.set({ scored: 0, total: 0 });
}

<script lang="ts">
	import { onMount } from 'svelte';
	import * as api from '$lib/api/client';
	import { exportRunUrl } from '$lib/api/client';
	import type { Run, Result, ModelSummary } from '$lib/types';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import ModelChip from '$lib/components/ModelChip.svelte';
	import JudgeScoreBadge from '$lib/components/JudgeScoreBadge.svelte';
	import LatencyChart from '$lib/components/charts/LatencyChart.svelte';
	import JudgeRadarChart from '$lib/components/charts/JudgeRadarChart.svelte';
	import CostChart from '$lib/components/charts/CostChart.svelte';
	import IterationLineChart from '$lib/components/charts/IterationLineChart.svelte';

	let { data } = $props();
	let run = $state<Run | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			run = await api.getRun(data.id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load run';
		} finally {
			loading = false;
		}
	});

	function groupByModel(results: Result[]): Record<string, Result[]> {
		const grouped: Record<string, Result[]> = {};
		for (const r of results) {
			if (!grouped[r.model]) grouped[r.model] = [];
			grouped[r.model].push(r);
		}
		return grouped;
	}

	function fallbackModelStats(results: Result[]) {
		const successful = results.filter((r) => r.status === 'success');
		if (successful.length === 0) {
			return { avgLatency: 0, avgToksPerSec: 0, successRate: 0, totalTokens: 0 };
		}
		const avgLatency =
			successful.reduce((sum, r) => sum + r.metrics.total_latency_ms, 0) / successful.length;
		const avgToksPerSec =
			successful.reduce((sum, r) => sum + r.metrics.tokens_per_second, 0) / successful.length;
		const totalTokens = successful.reduce((sum, r) => sum + r.metrics.completion_tokens, 0);
		return {
			avgLatency,
			avgToksPerSec,
			successRate: (successful.length / results.length) * 100,
			totalTokens
		};
	}

	function formatDate(dateStr?: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleString('en-US', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		});
	}

	function duration(start?: string, end?: string): string {
		if (!start || !end) return '-';
		const ms = new Date(end).getTime() - new Date(start).getTime();
		if (ms < 1000) return `${ms}ms`;
		return `${(ms / 1000).toFixed(1)}s`;
	}

	const statusDotColor: Record<string, string> = {
		success: 'bg-success',
		error: 'bg-error',
		timeout: 'bg-warning'
	};

	const hasSummary = $derived(run?.summary && run.summary.models.length > 0);
	const hasJudge = $derived(run?.config.judge_enabled && run?.config.judge_criteria?.length);
	const criteria = $derived(run?.config.judge_criteria ?? []);
	const models = $derived(run?.results ? [...new Set(run.results.map((r) => r.model))] : []);
</script>

<div>
	<a href="/runs" class="text-[15px] text-text-muted hover:text-accent mb-4 inline-block">&larr; Back to Runs</a>

	{#if loading}
		<p class="text-text-muted text-center py-10">Loading run...</p>
	{:else if error}
		<div class="bg-error/10 border border-error rounded-[--radius] px-4 py-3 text-error">{error}</div>
	{:else if run}
		<div class="flex justify-between items-start mb-6">
			<div>
				<h1 class="text-2xl mb-1">{run.suite_name}</h1>
				<p class="text-sm text-text-dim mono">{run.id}</p>
			</div>
			<div class="flex items-center gap-2">
				{#if run.status === 'complete' || run.status === 'failed'}
					<a href={exportRunUrl(run.id, 'csv')} download class="text-xs px-3 py-1 rounded-[6px] bg-bg-elevated border border-border text-text-muted no-underline transition-all duration-150 hover:border-accent hover:text-accent">CSV</a>
					<a href={exportRunUrl(run.id, 'json')} download class="text-xs px-3 py-1 rounded-[6px] bg-bg-elevated border border-border text-text-muted no-underline transition-all duration-150 hover:border-accent hover:text-accent">JSON</a>
				{/if}
				<StatusBadge status={run.status} />
			</div>
		</div>

		<div class="grid grid-cols-4 gap-3 mb-8">
			<div class="bg-bg-card border border-border rounded-[--radius] px-4 py-3">
				<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Started</span>
				<span class="text-lg font-semibold mono">{formatDate(run.started_at)}</span>
			</div>
			<div class="bg-bg-card border border-border rounded-[--radius] px-4 py-3">
				<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Ended</span>
				<span class="text-lg font-semibold mono">{formatDate(run.ended_at)}</span>
			</div>
			<div class="bg-bg-card border border-border rounded-[--radius] px-4 py-3">
				<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Duration</span>
				<span class="text-lg font-semibold mono">{duration(run.started_at, run.ended_at)}</span>
			</div>
			<div class="bg-bg-card border border-border rounded-[--radius] px-4 py-3">
				<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Results</span>
				<span class="text-lg font-semibold mono">{run.results?.length ?? 0}</span>
			</div>
		</div>

		{#if run.results && run.results.length > 0}
			<!-- Summary cards from backend -->
			{#if hasSummary}
				<section class="mb-8">
					<h2 class="text-lg text-text-muted uppercase tracking-wide mb-3">Model Overview</h2>
					<div class="grid grid-cols-[repeat(auto-fill,minmax(340px,1fr))] gap-3">
						{#each run.summary!.models as ms}
							<div class="card p-4">
								<div class="flex justify-between items-center mb-3">
									<ModelChip model={ms.model} />
									<span class="text-[15px] font-semibold text-success mono">{(ms.success_rate * 100).toFixed(0)}%</span>
								</div>
								<div class="grid grid-cols-2 gap-2 mb-2.5">
									<div class="flex flex-col gap-0.5">
										<span class="text-xs text-text-dim uppercase tracking-wide">Avg Latency</span>
										<span class="text-base font-semibold mono">{ms.avg_latency_ms.toFixed(0)}ms</span>
									</div>
									<div class="flex flex-col gap-0.5">
										<span class="text-xs text-text-dim uppercase tracking-wide">P95 Latency</span>
										<span class="text-base font-semibold mono">{ms.p95_latency_ms.toFixed(0)}ms</span>
									</div>
									<div class="flex flex-col gap-0.5">
										<span class="text-xs text-text-dim uppercase tracking-wide">Avg tok/s</span>
										<span class="text-base font-semibold mono">{ms.avg_tokens_per_second.toFixed(1)}</span>
									</div>
									<div class="flex flex-col gap-0.5">
										<span class="text-xs text-text-dim uppercase tracking-wide">Cost</span>
										<span class="text-base font-semibold mono">${ms.total_cost.toFixed(4)}</span>
									</div>
								</div>
								{#if ms.avg_judge_scores && Object.keys(ms.avg_judge_scores).length > 0}
									<div class="flex flex-wrap gap-1 pt-2 border-t border-border">
										{#each Object.entries(ms.avg_judge_scores) as [criterion, score]}
											<JudgeScoreBadge {criterion} {score} />
										{/each}
									</div>
								{/if}
							</div>
						{/each}
					</div>
				</section>

				<!-- Charts -->
				<section class="mb-8">
					<h2 class="text-lg text-text-muted uppercase tracking-wide mb-3">Visualizations</h2>
					<div class="grid grid-cols-2 gap-4">
						<div class="bg-bg-card border border-border rounded-[--radius-lg] p-5 h-[340px] flex flex-col">
							<h3 class="text-[15px] text-text-muted uppercase tracking-wide mb-3 shrink-0">Latency Comparison</h3>
							<div class="flex-1 min-h-0">
								<LatencyChart modelSummaries={run.summary!.models} />
							</div>
						</div>
						{#if hasJudge}
							<div class="bg-bg-card border border-border rounded-[--radius-lg] p-5 h-[340px] flex flex-col">
								<h3 class="text-[15px] text-text-muted uppercase tracking-wide mb-3 shrink-0">Judge-Scores</h3>
								<div class="flex-1 min-h-0">
									<JudgeRadarChart modelSummaries={run.summary!.models} {criteria} />
								</div>
							</div>
						{/if}
						<div class="bg-bg-card border border-border rounded-[--radius-lg] p-5 h-[340px] flex flex-col">
							<h3 class="text-[15px] text-text-muted uppercase tracking-wide mb-3 shrink-0">Cost Comparison</h3>
							<div class="flex-1 min-h-0">
								<CostChart modelSummaries={run.summary!.models} />
							</div>
						</div>
						{#if run.config.iterations > 1 && run.results}
							<div class="bg-bg-card border border-border rounded-[--radius-lg] p-5 h-[340px] flex flex-col">
								<h3 class="text-[15px] text-text-muted uppercase tracking-wide mb-3 shrink-0">Latency per Iteration</h3>
								<div class="flex-1 min-h-0">
									<IterationLineChart results={run.results} {models} />
								</div>
							</div>
						{/if}
					</div>
				</section>
			{/if}

			<!-- Detailed model comparison table -->
			<section class="mb-8">
				<h2 class="text-lg text-text-muted uppercase tracking-wide mb-3">Model Comparison</h2>
				<div class="overflow-x-auto">
					{#if hasSummary}
						<table class="w-full border-collapse">
							<thead>
								<tr>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Model</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Avg Latency</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">P50</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">P95</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Avg tok/s</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Success Rate</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Avg Cost</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Total</th>
									{#if hasJudge}
										{#each criteria as c}
											<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">{c}</th>
										{/each}
									{/if}
								</tr>
							</thead>
							<tbody>
								{#each run.summary!.models as ms}
									<tr class="hover:bg-bg-elevated">
										<td class="px-3 py-2.5 border-b border-border text-[15px]"><ModelChip model={ms.model} /></td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{ms.avg_latency_ms.toFixed(0)}ms</td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{ms.p50_latency_ms.toFixed(0)}ms</td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{ms.p95_latency_ms.toFixed(0)}ms</td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{ms.avg_tokens_per_second.toFixed(1)}</td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{(ms.success_rate * 100).toFixed(0)}%</td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">${ms.avg_cost.toFixed(4)}</td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">${ms.total_cost.toFixed(4)}</td>
										{#if hasJudge}
											{#each criteria as c}
												<td class="px-3 py-2.5 border-b border-border text-[15px]">
													{#if ms.avg_judge_scores?.[c]}
														<JudgeScoreBadge criterion={c} score={ms.avg_judge_scores[c]} />
													{:else}
														<span class="text-text-dim">-</span>
													{/if}
												</td>
											{/each}
										{/if}
									</tr>
								{/each}
							</tbody>
						</table>
					{:else}
						<!-- Fallback: client-side computed -->
						<table class="w-full border-collapse">
							<thead>
								<tr>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Model</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Avg Latency</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Avg tok/s</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Success Rate</th>
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Total Tokens</th>
								</tr>
							</thead>
							<tbody>
								{#each Object.entries(groupByModel(run.results)) as [model, results]}
									{@const stats = fallbackModelStats(results)}
									<tr class="hover:bg-bg-elevated">
										<td class="px-3 py-2.5 border-b border-border text-[15px]"><ModelChip {model} /></td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{stats.avgLatency.toFixed(0)}ms</td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{stats.avgToksPerSec.toFixed(1)}</td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{stats.successRate.toFixed(0)}%</td>
										<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{stats.totalTokens}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					{/if}
				</div>
			</section>

			<!-- All results -->
			<section class="mb-8">
				<h2 class="text-lg text-text-muted uppercase tracking-wide mb-3">All Results</h2>
				<div class="overflow-x-auto">
					<table class="w-full border-collapse">
						<thead>
							<tr>
								<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Model</th>
								<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Prompt</th>
								<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">#</th>
								<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Status</th>
								<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Latency</th>
								<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">tok/s</th>
								<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Tokens</th>
								{#if hasJudge}
									<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Judge</th>
								{/if}
							</tr>
						</thead>
						<tbody>
							{#each run.results as result}
								<tr class="hover:bg-bg-elevated {result.status !== 'success' ? 'opacity-60' : ''}">
									<td class="px-3 py-2.5 border-b border-border text-sm max-w-30 overflow-hidden text-ellipsis mono">{result.model.split('/').pop()}</td>
									<td class="px-3 py-2.5 border-b border-border text-[15px]">{result.prompt_name || '-'}</td>
									<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{result.iteration}</td>
									<td class="px-3 py-2.5 border-b border-border">
										<span class="inline-block w-2 h-2 rounded-full {statusDotColor[result.status] || 'bg-text-muted'}"></span>
									</td>
									<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{result.metrics.total_latency_ms.toFixed(0)}ms</td>
									<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{result.metrics.tokens_per_second.toFixed(1)}</td>
									<td class="px-3 py-2.5 border-b border-border text-[15px] mono">{result.metrics.completion_tokens}</td>
									{#if hasJudge}
										<td class="px-3 py-2.5 border-b border-border">
											{#if result.judge_scores?.length}
												<div class="flex flex-wrap gap-1">
													{#each result.judge_scores as js}
														<JudgeScoreBadge criterion={js.criterion} score={js.score} />
													{/each}
												</div>
											{:else}
												<span class="text-text-dim">-</span>
											{/if}
										</td>
									{/if}
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>
		{:else}
			<div class="card text-center py-10 text-text-muted">
				<p>No results available.</p>
			</div>
		{/if}
	{/if}
</div>

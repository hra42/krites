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

	const hasSummary = $derived(run?.summary && run.summary.models.length > 0);
	const hasJudge = $derived(run?.config.judge_enabled && run?.config.judge_criteria?.length);
	const criteria = $derived(run?.config.judge_criteria ?? []);
	const models = $derived(run?.results ? [...new Set(run.results.map((r) => r.model))] : []);
</script>

<div class="page">
	<a href="/runs" class="back">&larr; Back to Runs</a>

	{#if loading}
		<p class="loading">Loading run...</p>
	{:else if error}
		<div class="error-banner">{error}</div>
	{:else if run}
		<div class="header">
			<div>
				<h1>{run.suite_name}</h1>
				<p class="run-id mono">{run.id}</p>
			</div>
			<div class="header-actions">
				{#if run.status === 'complete' || run.status === 'failed'}
					<a href={exportRunUrl(run.id, 'csv')} download class="btn btn-sm">CSV</a>
					<a href={exportRunUrl(run.id, 'json')} download class="btn btn-sm">JSON</a>
				{/if}
				<StatusBadge status={run.status} />
			</div>
		</div>

		<div class="meta-grid">
			<div class="meta-item">
				<span class="meta-label">Started</span>
				<span class="meta-value mono">{formatDate(run.started_at)}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Ended</span>
				<span class="meta-value mono">{formatDate(run.ended_at)}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Duration</span>
				<span class="meta-value mono">{duration(run.started_at, run.ended_at)}</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Results</span>
				<span class="meta-value mono">{run.results?.length ?? 0}</span>
			</div>
		</div>

		{#if run.results && run.results.length > 0}
			<!-- Summary cards from backend -->
			{#if hasSummary}
				<section class="section">
					<h2>Model Overview</h2>
					<div class="summary-cards">
						{#each run.summary!.models as ms}
							<div class="summary-card card">
								<div class="summary-card-header">
									<ModelChip model={ms.model} />
									<span class="success-rate mono">{(ms.success_rate * 100).toFixed(0)}%</span>
								</div>
								<div class="summary-metrics">
									<div class="summary-metric">
										<span class="sm-label">Avg Latency</span>
										<span class="sm-value mono">{ms.avg_latency_ms.toFixed(0)}ms</span>
									</div>
									<div class="summary-metric">
										<span class="sm-label">P95 Latency</span>
										<span class="sm-value mono">{ms.p95_latency_ms.toFixed(0)}ms</span>
									</div>
									<div class="summary-metric">
										<span class="sm-label">Avg tok/s</span>
										<span class="sm-value mono">{ms.avg_tokens_per_second.toFixed(1)}</span>
									</div>
									<div class="summary-metric">
										<span class="sm-label">Cost</span>
										<span class="sm-value mono">${ms.total_cost.toFixed(4)}</span>
									</div>
								</div>
								{#if ms.avg_judge_scores && Object.keys(ms.avg_judge_scores).length > 0}
									<div class="summary-judges">
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
				<section class="section">
					<h2>Visualizations</h2>
					<div class="charts-grid">
						<div class="chart-container">
							<h3>Latency Comparison</h3>
							<LatencyChart modelSummaries={run.summary!.models} />
						</div>
						{#if hasJudge}
							<div class="chart-container">
								<h3>Judge-Scores</h3>
								<JudgeRadarChart modelSummaries={run.summary!.models} {criteria} />
							</div>
						{/if}
						<div class="chart-container">
							<h3>Cost Comparison</h3>
							<CostChart modelSummaries={run.summary!.models} />
						</div>
						{#if run.config.iterations > 1 && run.results}
							<div class="chart-container">
								<h3>Latency per Iteration</h3>
								<IterationLineChart results={run.results} {models} />
							</div>
						{/if}
					</div>
				</section>
			{/if}

			<!-- Detailed model comparison table -->
			<section class="section">
				<h2>Model Comparison</h2>
				<div class="table-wrapper">
					{#if hasSummary}
						<table>
							<thead>
								<tr>
									<th>Model</th>
									<th>Avg Latency</th>
									<th>P50</th>
									<th>P95</th>
									<th>Avg tok/s</th>
									<th>Success Rate</th>
									<th>Avg Cost</th>
									<th>Total</th>
									{#if hasJudge}
										{#each criteria as c}
											<th>{c}</th>
										{/each}
									{/if}
								</tr>
							</thead>
							<tbody>
								{#each run.summary!.models as ms}
									<tr>
										<td><ModelChip model={ms.model} /></td>
										<td class="mono">{ms.avg_latency_ms.toFixed(0)}ms</td>
										<td class="mono">{ms.p50_latency_ms.toFixed(0)}ms</td>
										<td class="mono">{ms.p95_latency_ms.toFixed(0)}ms</td>
										<td class="mono">{ms.avg_tokens_per_second.toFixed(1)}</td>
										<td class="mono">{(ms.success_rate * 100).toFixed(0)}%</td>
										<td class="mono">${ms.avg_cost.toFixed(4)}</td>
										<td class="mono">${ms.total_cost.toFixed(4)}</td>
										{#if hasJudge}
											{#each criteria as c}
												<td>
													{#if ms.avg_judge_scores?.[c]}
														<JudgeScoreBadge criterion={c} score={ms.avg_judge_scores[c]} />
													{:else}
														<span class="text-dim">-</span>
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
						<table>
							<thead>
								<tr>
									<th>Model</th>
									<th>Avg Latency</th>
									<th>Avg tok/s</th>
									<th>Success Rate</th>
									<th>Total Tokens</th>
								</tr>
							</thead>
							<tbody>
								{#each Object.entries(groupByModel(run.results)) as [model, results]}
									{@const stats = fallbackModelStats(results)}
									<tr>
										<td><ModelChip {model} /></td>
										<td class="mono">{stats.avgLatency.toFixed(0)}ms</td>
										<td class="mono">{stats.avgToksPerSec.toFixed(1)}</td>
										<td class="mono">{stats.successRate.toFixed(0)}%</td>
										<td class="mono">{stats.totalTokens}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					{/if}
				</div>
			</section>

			<!-- All results -->
			<section class="section">
				<h2>All Results</h2>
				<div class="table-wrapper">
					<table class="results-table">
						<thead>
							<tr>
								<th>Model</th>
								<th>Prompt</th>
								<th>#</th>
								<th>Status</th>
								<th>Latency</th>
								<th>tok/s</th>
								<th>Tokens</th>
								{#if hasJudge}
									<th>Judge</th>
								{/if}
							</tr>
						</thead>
						<tbody>
							{#each run.results as result}
								<tr class:error-row={result.status !== 'success'}>
									<td class="mono model-cell">{result.model.split('/').pop()}</td>
									<td>{result.prompt_name || '-'}</td>
									<td class="mono">{result.iteration}</td>
									<td>
										<span class="status-dot {result.status}"></span>
									</td>
									<td class="mono">{result.metrics.total_latency_ms.toFixed(0)}ms</td>
									<td class="mono">{result.metrics.tokens_per_second.toFixed(1)}</td>
									<td class="mono">{result.metrics.completion_tokens}</td>
									{#if hasJudge}
										<td>
											{#if result.judge_scores?.length}
												<div class="judge-badges-row">
													{#each result.judge_scores as js}
														<JudgeScoreBadge criterion={js.criterion} score={js.score} />
													{/each}
												</div>
											{:else}
												<span class="text-dim">-</span>
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
			<div class="empty card">
				<p>No results available.</p>
			</div>
		{/if}
	{/if}
</div>

<style>
	.back {
		font-size: 13px;
		color: var(--color-text-muted);
		margin-bottom: 16px;
		display: inline-block;
	}

	.back:hover {
		color: var(--color-accent);
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 24px;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.btn-sm {
		font-size: 12px;
		padding: 4px 12px;
		border-radius: 6px;
		background: var(--color-elevated);
		border: 1px solid var(--color-border);
		color: var(--color-text-muted);
		text-decoration: none;
		transition: all 0.15s ease;
	}

	.btn-sm:hover {
		border-color: var(--color-accent);
		color: var(--color-accent);
	}

	h1 {
		font-size: 22px;
		margin-bottom: 4px;
	}

	.run-id {
		font-size: 12px;
		color: var(--color-text-dim);
	}

	.meta-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 12px;
		margin-bottom: 32px;
	}

	.meta-item {
		background: var(--color-bg-card);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		padding: 12px 16px;
	}

	.meta-label {
		display: block;
		font-size: 11px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 4px;
	}

	.meta-value {
		font-size: 15px;
		font-weight: 600;
	}

	.section {
		margin-bottom: 32px;
	}

	.section h2 {
		font-size: 16px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 12px;
	}

	/* Summary cards */
	.summary-cards {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
		gap: 12px;
	}

	.summary-card {
		padding: 16px;
	}

	.summary-card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}

	.success-rate {
		font-size: 13px;
		font-weight: 600;
		color: var(--color-success);
	}

	.summary-metrics {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
		margin-bottom: 10px;
	}

	.summary-metric {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.sm-label {
		font-size: 10px;
		color: var(--color-text-dim);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.sm-value {
		font-size: 14px;
		font-weight: 600;
	}

	.summary-judges {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		padding-top: 8px;
		border-top: 1px solid var(--color-border);
	}

	/* Charts */
	.charts-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
	}

	.chart-container {
		background: var(--color-bg-card);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: 20px;
		height: 340px;
		display: flex;
		flex-direction: column;
	}

	.chart-container h3 {
		font-size: 13px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 12px;
		flex-shrink: 0;
	}

	.chart-container :global(.chart-wrapper) {
		flex: 1;
		min-height: 0;
	}

	/* Tables */
	.table-wrapper {
		overflow-x: auto;
	}

	table {
		width: 100%;
		border-collapse: collapse;
	}

	th {
		text-align: left;
		font-size: 11px;
		color: var(--color-text-dim);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 8px 12px;
		border-bottom: 1px solid var(--color-border);
	}

	td {
		padding: 10px 12px;
		border-bottom: 1px solid var(--color-border);
		font-size: 13px;
	}

	tr:hover {
		background: var(--color-bg-elevated);
	}

	.error-row {
		opacity: 0.6;
	}

	.model-cell {
		font-size: 12px;
		max-width: 120px;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.status-dot {
		display: inline-block;
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.status-dot.success {
		background: var(--color-success);
	}

	.status-dot.error {
		background: var(--color-error);
	}

	.status-dot.timeout {
		background: var(--color-warning);
	}

	.judge-badges-row {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}

	.text-dim {
		color: var(--color-text-dim);
	}

	.loading {
		color: var(--color-text-muted);
		text-align: center;
		padding: 40px;
	}

	.error-banner {
		background: rgba(248, 113, 113, 0.1);
		border: 1px solid var(--color-error);
		border-radius: var(--radius);
		padding: 12px 16px;
		color: var(--color-error);
	}

	.empty {
		text-align: center;
		padding: 40px;
		color: var(--color-text-muted);
	}
</style>

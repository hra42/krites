<script lang="ts">
	import { onMount } from 'svelte';
	import * as api from '$lib/api/client';
	import type { AnalyticsOverview, CrossRunModelStats, ModelTrendPoint } from '$lib/types';
	import ModelComparisonChart from '$lib/components/charts/ModelComparisonChart.svelte';
	import TrendLineChart from '$lib/components/charts/TrendLineChart.svelte';

	let overview = $state<AnalyticsOverview | null>(null);
	let modelStats = $state<CrossRunModelStats[]>([]);
	let trends = $state<ModelTrendPoint[]>([]);
	let selectedModel = $state('');
	let trendMetric = $state<'latency' | 'tps'>('latency');
	let loading = $state(true);
	let error = $state<string | null>(null);

	const distinctModels = $derived([...new Set(modelStats.map((s) => s.model))]);

	onMount(async () => {
		try {
			const [ov, stats] = await Promise.all([
				api.getAnalyticsOverview(),
				api.getModelComparison()
			]);
			overview = ov;
			modelStats = stats;
			if (stats.length > 0) {
				selectedModel = stats[0].model;
				trends = await api.getModelTrends(selectedModel);
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load analytics';
		} finally {
			loading = false;
		}
	});

	async function loadTrends() {
		if (!selectedModel) return;
		try {
			trends = await api.getModelTrends(selectedModel);
		} catch {
			trends = [];
		}
	}

	function formatCost(cost: number): string {
		return cost < 0.01 ? `$${cost.toFixed(6)}` : `$${cost.toFixed(4)}`;
	}

	let sortKey = $state<string>('avg_latency_ms');
	let sortAsc = $state(true);

	function sortStats(stats: CrossRunModelStats[]): CrossRunModelStats[] {
		return [...stats].sort((a, b) => {
			const va = (a as unknown as Record<string, number>)[sortKey];
			const vb = (b as unknown as Record<string, number>)[sortKey];
			return sortAsc ? va - vb : vb - va;
		});
	}

	function toggleSort(key: string) {
		if (sortKey === key) {
			sortAsc = !sortAsc;
		} else {
			sortKey = key;
			sortAsc = true;
		}
	}

	const sortedStats = $derived(sortStats(modelStats));
</script>

<div class="page">
	<h1>Analytics</h1>
	<p class="subtitle">Model comparisons across all runs</p>

	{#if loading}
		<p class="loading">Loading analytics...</p>
	{:else if error}
		<div class="error-banner">{error}</div>
	{:else if overview}
		<div class="stats">
			<div class="stat-card card">
				<span class="stat-value mono">{overview.total_runs}</span>
				<span class="stat-label">Total Runs</span>
			</div>
			<div class="stat-card card">
				<span class="stat-value mono">{overview.completed_runs}</span>
				<span class="stat-label">Completed</span>
			</div>
			<div class="stat-card card">
				<span class="stat-value mono">{overview.total_results}</span>
				<span class="stat-label">Results</span>
			</div>
			<div class="stat-card card">
				<span class="stat-value mono">{overview.distinct_models}</span>
				<span class="stat-label">Models</span>
			</div>
			<div class="stat-card card">
				<span class="stat-value mono">{formatCost(overview.total_cost)}</span>
				<span class="stat-label">Total Cost</span>
			</div>
			<div class="stat-card card">
				<span class="stat-value mono">{overview.avg_latency_ms.toFixed(0)} ms</span>
				<span class="stat-label">Avg Latency</span>
			</div>
		</div>

		{#if modelStats.length > 0}
			<section class="section">
				<h2>Model Comparison</h2>
				<div class="chart-container card">
					<ModelComparisonChart stats={modelStats} />
				</div>

				<div class="table-container card">
					<table>
						<thead>
							<tr>
								<th>Modell</th>
								<th class="sortable" onclick={() => toggleSort('run_count')}>
									Runs {sortKey === 'run_count' ? (sortAsc ? '▲' : '▼') : ''}
								</th>
								<th class="sortable" onclick={() => toggleSort('avg_latency_ms')}>
									Avg Latency {sortKey === 'avg_latency_ms' ? (sortAsc ? '▲' : '▼') : ''}
								</th>
								<th class="sortable" onclick={() => toggleSort('avg_tps')}>
									Avg tok/s {sortKey === 'avg_tps' ? (sortAsc ? '▲' : '▼') : ''}
								</th>
								<th class="sortable" onclick={() => toggleSort('total_cost')}>
									Cost {sortKey === 'total_cost' ? (sortAsc ? '▲' : '▼') : ''}
								</th>
								<th class="sortable" onclick={() => toggleSort('avg_success_rate')}>
									Success {sortKey === 'avg_success_rate' ? (sortAsc ? '▲' : '▼') : ''}
								</th>
							</tr>
						</thead>
						<tbody>
							{#each sortedStats as stat}
								<tr>
									<td class="mono">{stat.model}</td>
									<td class="mono">{stat.run_count}</td>
									<td class="mono">{stat.avg_latency_ms.toFixed(0)} ms</td>
									<td class="mono">{stat.avg_tokens_per_second.toFixed(1)}</td>
									<td class="mono">{formatCost(stat.total_cost)}</td>
									<td class="mono">{(stat.avg_success_rate * 100).toFixed(0)}%</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>

			<section class="section">
				<h2>Performance Trend</h2>
				<div class="trend-controls">
					<select bind:value={selectedModel} onchange={loadTrends}>
						{#each distinctModels as model}
							<option value={model}>{model}</option>
						{/each}
					</select>
					<div class="metric-toggle">
						<button class:active={trendMetric === 'latency'} onclick={() => (trendMetric = 'latency')}>
							Latency
						</button>
						<button class:active={trendMetric === 'tps'} onclick={() => (trendMetric = 'tps')}>
							tok/s
						</button>
					</div>
				</div>
				{#if trends.length > 0}
					<div class="chart-container card">
						{#key `${selectedModel}-${trendMetric}`}
							<TrendLineChart {trends} metric={trendMetric} />
						{/key}
					</div>
				{:else}
					<p class="empty">No trend data for this model</p>
				{/if}
			</section>
		{:else}
			<div class="empty-state card">
				<p>No benchmark data yet.</p>
				<a href="/suites" class="btn btn-primary">Create First Suite</a>
			</div>
		{/if}
	{/if}
</div>

<style>
	.page {
		max-width: 1400px;
	}

	h1 {
		font-size: 28px;
		margin-bottom: 4px;
	}

	.subtitle {
		color: var(--color-text-muted);
		margin-bottom: 32px;
	}

	.stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;
		margin-bottom: 40px;
	}

	.stat-card {
		display: flex;
		flex-direction: column;
		gap: 4px;
		text-align: center;
		padding: 24px;
	}

	.stat-value {
		font-size: 28px;
		font-weight: 700;
		color: var(--color-accent);
	}

	.stat-label {
		font-size: 12px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.section {
		margin-bottom: 40px;
	}

	.section h2 {
		font-size: 18px;
		margin-bottom: 16px;
	}

	.chart-container {
		padding: 20px;
		height: 300px;
		display: flex;
		flex-direction: column;
	}

	.chart-container :global(.chart-wrapper) {
		flex: 1;
		min-height: 0;
	}

	.table-container {
		margin-top: 16px;
		overflow-x: auto;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 13px;
	}

	thead {
		border-bottom: 1px solid var(--color-border);
	}

	th {
		text-align: left;
		padding: 10px 12px;
		color: var(--color-text-muted);
		font-weight: 500;
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	th.sortable {
		cursor: pointer;
		user-select: none;
	}

	th.sortable:hover {
		color: var(--color-accent);
	}

	td {
		padding: 10px 12px;
		border-top: 1px solid var(--color-border);
	}

	.trend-controls {
		display: flex;
		align-items: center;
		gap: 16px;
		margin-bottom: 16px;
	}

	select {
		background: var(--color-input);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		color: var(--color-text);
		padding: 8px 12px;
		font-size: 13px;
		font-family: 'JetBrains Mono', monospace;
	}

	.metric-toggle {
		display: flex;
		gap: 0;
		border: 1px solid var(--color-border);
		border-radius: 8px;
		overflow: hidden;
	}

	.metric-toggle button {
		background: var(--color-input);
		border: none;
		color: var(--color-text-muted);
		padding: 8px 16px;
		font-size: 12px;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.metric-toggle button.active {
		background: var(--color-accent);
		color: var(--color-bg);
		font-weight: 600;
	}

	.empty {
		color: var(--color-text-muted);
		font-size: 14px;
		padding: 20px 0;
	}

	.empty-state {
		text-align: center;
		padding: 48px;
	}

	.empty-state p {
		color: var(--color-text-muted);
		margin-bottom: 16px;
	}

	.loading {
		color: var(--color-text-muted);
	}

	.error-banner {
		background: rgba(248, 113, 113, 0.1);
		border: 1px solid var(--color-error);
		color: var(--color-error);
		padding: 12px 16px;
		border-radius: 8px;
	}
</style>

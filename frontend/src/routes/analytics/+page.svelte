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

<div>
	<h1 class="text-3xl mb-1">Analytics</h1>
	<p class="text-text-muted mb-8">Model comparisons across all runs</p>

	{#if loading}
		<p class="text-text-muted">Loading analytics...</p>
	{:else if error}
		<div class="bg-error/10 border border-error text-error px-4 py-3 rounded-[--radius]">{error}</div>
	{:else if overview}
		<div class="grid grid-cols-3 gap-4 mb-10">
			<div class="card flex flex-col gap-1 text-center p-6">
				<span class="text-3xl font-bold text-accent mono">{overview.total_runs}</span>
				<span class="text-sm text-text-muted uppercase tracking-wide">Total Runs</span>
			</div>
			<div class="card flex flex-col gap-1 text-center p-6">
				<span class="text-3xl font-bold text-accent mono">{overview.completed_runs}</span>
				<span class="text-sm text-text-muted uppercase tracking-wide">Completed</span>
			</div>
			<div class="card flex flex-col gap-1 text-center p-6">
				<span class="text-3xl font-bold text-accent mono">{overview.total_results}</span>
				<span class="text-sm text-text-muted uppercase tracking-wide">Results</span>
			</div>
			<div class="card flex flex-col gap-1 text-center p-6">
				<span class="text-3xl font-bold text-accent mono">{overview.distinct_models}</span>
				<span class="text-sm text-text-muted uppercase tracking-wide">Models</span>
			</div>
			<div class="card flex flex-col gap-1 text-center p-6">
				<span class="text-3xl font-bold text-accent mono">{formatCost(overview.total_cost)}</span>
				<span class="text-sm text-text-muted uppercase tracking-wide">Total Cost</span>
			</div>
			<div class="card flex flex-col gap-1 text-center p-6">
				<span class="text-3xl font-bold text-accent mono">{overview.avg_latency_ms.toFixed(0)} ms</span>
				<span class="text-sm text-text-muted uppercase tracking-wide">Avg Latency</span>
			</div>
		</div>

		{#if modelStats.length > 0}
			<section class="mb-10">
				<h2 class="text-xl mb-4">Model Comparison</h2>
				<div class="card p-5 h-[300px] flex flex-col">
					<div class="flex-1 min-h-0">
						<ModelComparisonChart stats={modelStats} />
					</div>
				</div>

				<div class="card mt-4 overflow-x-auto">
					<table class="w-full border-collapse text-[15px]">
						<thead>
							<tr class="border-b border-border">
								<th class="text-left px-3 py-2.5 text-text-muted font-medium text-sm uppercase tracking-wide">Modell</th>
								<th class="text-left px-3 py-2.5 text-text-muted font-medium text-sm uppercase tracking-wide cursor-pointer select-none hover:text-accent" onclick={() => toggleSort('run_count')}>
									Runs {sortKey === 'run_count' ? (sortAsc ? '▲' : '▼') : ''}
								</th>
								<th class="text-left px-3 py-2.5 text-text-muted font-medium text-sm uppercase tracking-wide cursor-pointer select-none hover:text-accent" onclick={() => toggleSort('avg_latency_ms')}>
									Avg Latency {sortKey === 'avg_latency_ms' ? (sortAsc ? '▲' : '▼') : ''}
								</th>
								<th class="text-left px-3 py-2.5 text-text-muted font-medium text-sm uppercase tracking-wide cursor-pointer select-none hover:text-accent" onclick={() => toggleSort('avg_tps')}>
									Avg tok/s {sortKey === 'avg_tps' ? (sortAsc ? '▲' : '▼') : ''}
								</th>
								<th class="text-left px-3 py-2.5 text-text-muted font-medium text-sm uppercase tracking-wide cursor-pointer select-none hover:text-accent" onclick={() => toggleSort('total_cost')}>
									Cost {sortKey === 'total_cost' ? (sortAsc ? '▲' : '▼') : ''}
								</th>
								<th class="text-left px-3 py-2.5 text-text-muted font-medium text-sm uppercase tracking-wide cursor-pointer select-none hover:text-accent" onclick={() => toggleSort('avg_success_rate')}>
									Success {sortKey === 'avg_success_rate' ? (sortAsc ? '▲' : '▼') : ''}
								</th>
							</tr>
						</thead>
						<tbody>
							{#each sortedStats as stat}
								<tr>
									<td class="px-3 py-2.5 border-t border-border mono">{stat.model}</td>
									<td class="px-3 py-2.5 border-t border-border mono">{stat.run_count}</td>
									<td class="px-3 py-2.5 border-t border-border mono">{stat.avg_latency_ms.toFixed(0)} ms</td>
									<td class="px-3 py-2.5 border-t border-border mono">{stat.avg_tokens_per_second.toFixed(1)}</td>
									<td class="px-3 py-2.5 border-t border-border mono">{formatCost(stat.total_cost)}</td>
									<td class="px-3 py-2.5 border-t border-border mono">{(stat.avg_success_rate * 100).toFixed(0)}%</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>

			<section class="mb-10">
				<h2 class="text-xl mb-4">Performance Trend</h2>
				<div class="flex items-center gap-4 mb-4">
					<select class="bg-bg-input border border-border rounded-[--radius] text-text px-3 py-2 text-[15px] font-mono" bind:value={selectedModel} onchange={loadTrends}>
						{#each distinctModels as model}
							<option value={model}>{model}</option>
						{/each}
					</select>
					<div class="flex border border-border rounded-[--radius] overflow-hidden">
						<button
							class="border-none px-4 py-2 text-sm cursor-pointer transition-all duration-150 {trendMetric === 'latency' ? 'bg-accent text-bg font-semibold' : 'bg-bg-input text-text-muted'}"
							onclick={() => (trendMetric = 'latency')}
						>
							Latency
						</button>
						<button
							class="border-none px-4 py-2 text-sm cursor-pointer transition-all duration-150 {trendMetric === 'tps' ? 'bg-accent text-bg font-semibold' : 'bg-bg-input text-text-muted'}"
							onclick={() => (trendMetric = 'tps')}
						>
							tok/s
						</button>
					</div>
				</div>
				{#if trends.length > 0}
					<div class="card p-5 h-[300px] flex flex-col">
						{#key `${selectedModel}-${trendMetric}`}
							<div class="flex-1 min-h-0">
								<TrendLineChart {trends} metric={trendMetric} />
							</div>
						{/key}
					</div>
				{:else}
					<p class="text-text-muted text-base py-5">No trend data for this model</p>
				{/if}
			</section>
		{:else}
			<div class="card text-center py-12">
				<p class="text-text-muted mb-4">No benchmark data yet.</p>
				<a href="/suites" class="btn btn-primary">Create First Suite</a>
			</div>
		{/if}
	{/if}
</div>

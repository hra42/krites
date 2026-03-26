<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import * as api from '$lib/api/client';
	import {
		currentRun,
		liveResults,
		isRunning,
		runError,
		resultsByModel,
		runProgress,
		modelAggregates,
		modelJudgeAggregates,
		judgeProgress,
		totalCalls,
		handleStreamEvent,
		resetRunState
	} from '$lib/stores/run';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import ModelChip from '$lib/components/ModelChip.svelte';
	import ProgressBar from '$lib/components/ProgressBar.svelte';
	import ResultCard from '$lib/components/ResultCard.svelte';
	import JudgeScoreBadge from '$lib/components/JudgeScoreBadge.svelte';
	import type { Suite } from '$lib/types';

	let { data } = $props();
	let suite = $state<Suite | null>(null);
	let loading = $state(true);
	let loadError = $state<string | null>(null);
	let started = $state(false);
	let cleanup: (() => void) | null = null;

	const successfulCount = $derived($liveResults.filter((r) => r.status === 'success').length);
	const benchmarkDone = $derived(
		$runProgress.completed > 0 && $runProgress.completed === $runProgress.total
	);
	const showJudgeProgress = $derived(
		suite?.config.judge_enabled && benchmarkDone && successfulCount > 0
	);

	onMount(async () => {
		resetRunState();
		try {
			suite = await api.getSuite(data.id);
			if (suite) {
				totalCalls.set(
					suite.models.length * suite.prompts.length * suite.config.iterations
				);
			}
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Suite konnte nicht geladen werden';
		} finally {
			loading = false;
		}
	});

	onDestroy(() => {
		if (cleanup) cleanup();
	});

	async function startBenchmark() {
		if (!suite) return;
		started = true;
		try {
			const run = await api.startRun(suite.id);
			currentRun.set(run);
			cleanup = api.streamRun(
				run.id,
				handleStreamEvent,
				() => {
					if ($isRunning) {
						runError.set('SSE-Verbindung verloren');
					}
				}
			);
		} catch (e) {
			runError.set(e instanceof Error ? e.message : 'Failed to start benchmark');
			started = false;
		}
	}
</script>

<div class="page">
	<a href="/suites/{data.id}" class="back">&larr; Back to Suite</a>

	{#if loading}
		<p class="loading">Loading suite...</p>
	{:else if loadError}
		<div class="error-banner">{loadError}</div>
	{:else if suite}
		<div class="header">
			<h1>Benchmark: {suite.name}</h1>
		</div>

		{#if !started}
			<!-- Pre-run summary -->
			<div class="pre-run card">
				<h2>Summary</h2>
				<div class="summary-grid">
					<div class="summary-item">
						<span class="summary-label">Models</span>
						<span class="summary-value mono">{suite.models.length}</span>
					</div>
					<div class="summary-item">
						<span class="summary-label">Prompts</span>
						<span class="summary-value mono">{suite.prompts.length}</span>
					</div>
					<div class="summary-item">
						<span class="summary-label">Iterations</span>
						<span class="summary-value mono">{suite.config.iterations}</span>
					</div>
					<div class="summary-item">
						<span class="summary-label">Total API Calls</span>
						<span class="summary-value mono highlight">{$totalCalls}</span>
					</div>
				</div>

				{#if suite.config.judge_enabled}
					<div class="judge-info">
						<span class="judge-label">Judge:</span>
						<span class="mono">{suite.config.judge_model}</span>
						<span class="judge-criteria">
							{suite.config.judge_criteria?.join(', ')}
						</span>
					</div>
				{/if}

				<div class="models-preview">
					{#each suite.models as model}
						<ModelChip {model} />
					{/each}
				</div>

				<button class="btn btn-primary start-btn" onclick={startBenchmark}>
					Start Benchmark
				</button>
			</div>
		{:else}
			<!-- Running / completed -->
			<div class="run-status">
				{#if $currentRun}
					<StatusBadge status={$currentRun.status} />
				{/if}
				{#if $runError}
					<span class="error-text">{$runError}</span>
				{/if}
			</div>

			<div class="progress-section">
				<div class="progress-label-row">
					<span class="progress-label">Benchmark</span>
				</div>
				<ProgressBar
					completed={$runProgress.completed}
					total={$runProgress.total}
					percent={$runProgress.percent}
				/>
			</div>

			{#if showJudgeProgress}
				<div class="progress-section">
					<div class="progress-label-row">
						<span class="progress-label">Judge Scoring</span>
					</div>
					<ProgressBar
						completed={$judgeProgress.scored}
						total={successfulCount}
						percent={successfulCount > 0 ? Math.round(($judgeProgress.scored / successfulCount) * 100) : 0}
					/>
				</div>
			{/if}

			<!-- Live aggregates per model -->
			{#if Object.keys($modelAggregates).length > 0}
				<div class="aggregates">
					{#each Object.entries($modelAggregates) as [model, agg]}
						<div class="aggregate-card card">
							<div class="aggregate-model">
								<ModelChip {model} />
							</div>
							<div class="aggregate-metrics mono">
								<div class="agg-metric">
									<span class="agg-label">Avg Latency</span>
									<span class="agg-value">{agg.avgLatency.toFixed(0)}ms</span>
								</div>
								<div class="agg-metric">
									<span class="agg-label">Avg tok/s</span>
									<span class="agg-value">{agg.avgTokensPerSecond.toFixed(1)}</span>
								</div>
								<div class="agg-metric">
									<span class="agg-label">Success</span>
									<span class="agg-value">{agg.successCount}/{agg.totalCount}</span>
								</div>
							</div>
							{#if $modelJudgeAggregates[model]}
								<div class="agg-judges">
									{#each Object.entries($modelJudgeAggregates[model]) as [criterion, jagg]}
										<JudgeScoreBadge {criterion} score={jagg.avg} />
									{/each}
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}

			<!-- Results grouped by model -->
			{#each Object.entries($resultsByModel) as [model, results]}
				<section class="model-results">
					<h3>
						<ModelChip {model} />
						<span class="result-count mono">{results.length} Results</span>
					</h3>
					<div class="results-list">
						{#each results as result (result.id)}
							<ResultCard {result} />
						{/each}
					</div>
				</section>
			{/each}

			<!-- Post-run actions -->
			{#if !$isRunning && $currentRun && ($currentRun.status === 'complete' || $currentRun.status === 'failed')}
				<div class="post-run">
					<button class="btn btn-primary" onclick={() => goto(`/runs/${$currentRun?.id}`)}>
						View Run Details
					</button>
				</div>
			{/if}
		{/if}
	{/if}
</div>

<style>
	.page {
		max-width: 1200px;
	}

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
		margin-bottom: 24px;
	}

	.header h1 {
		font-size: 22px;
	}

	.pre-run {
		padding: 24px;
	}

	.pre-run h2 {
		font-size: 16px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 16px;
	}

	.summary-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 12px;
		margin-bottom: 16px;
	}

	.summary-item {
		text-align: center;
		padding: 12px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
	}

	.summary-label {
		display: block;
		font-size: 11px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 4px;
	}

	.summary-value {
		font-size: 20px;
		font-weight: 700;
	}

	.highlight {
		color: var(--color-accent);
	}

	.judge-info {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 16px;
		padding: 10px 14px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: 12px;
	}

	.judge-label {
		color: var(--color-text-muted);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.judge-criteria {
		color: var(--color-text-dim);
		margin-left: auto;
	}

	.models-preview {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-bottom: 20px;
	}

	.start-btn {
		width: 100%;
		padding: 12px;
		font-size: 15px;
	}

	.run-status {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 16px;
	}

	.error-text {
		color: var(--color-error);
		font-size: 13px;
	}

	.progress-section {
		margin-bottom: 16px;
	}

	.progress-label-row {
		margin-bottom: 6px;
	}

	.progress-label {
		font-size: 11px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.aggregates {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: 12px;
		margin-bottom: 24px;
	}

	.aggregate-card {
		padding: 14px 16px;
	}

	.aggregate-model {
		margin-bottom: 10px;
	}

	.aggregate-metrics {
		display: flex;
		gap: 16px;
	}

	.agg-metric {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.agg-label {
		font-size: 10px;
		color: var(--color-text-dim);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.agg-value {
		font-size: 14px;
		font-weight: 600;
	}

	.agg-judges {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin-top: 10px;
		padding-top: 8px;
		border-top: 1px solid var(--color-border);
	}

	.model-results {
		margin-bottom: 24px;
	}

	.model-results h3 {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-bottom: 12px;
	}

	.result-count {
		font-size: 12px;
		color: var(--color-text-dim);
	}

	.results-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.post-run {
		margin-top: 24px;
		padding-top: 24px;
		border-top: 1px solid var(--color-border);
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
</style>

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

<div>
	<a href="/suites/{data.id}" class="text-[15px] text-text-muted hover:text-accent mb-4 inline-block">&larr; Back to Suite</a>

	{#if loading}
		<p class="text-text-muted text-center py-10">Loading suite...</p>
	{:else if loadError}
		<div class="bg-error/10 border border-error rounded-[--radius] px-4 py-3 text-error">{loadError}</div>
	{:else if suite}
		<div class="mb-6">
			<h1 class="text-2xl">Benchmark: {suite.name}</h1>
		</div>

		{#if !started}
			<!-- Pre-run summary -->
			<div class="card p-6">
				<h2 class="text-lg text-text-muted uppercase tracking-wide mb-4">Summary</h2>
				<div class="grid grid-cols-4 gap-3 mb-4">
					<div class="text-center p-3 bg-bg border border-border rounded-[--radius-sm]">
						<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Models</span>
						<span class="text-xl font-bold mono">{suite.models.length}</span>
					</div>
					<div class="text-center p-3 bg-bg border border-border rounded-[--radius-sm]">
						<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Prompts</span>
						<span class="text-xl font-bold mono">{suite.prompts.length}</span>
					</div>
					<div class="text-center p-3 bg-bg border border-border rounded-[--radius-sm]">
						<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Iterations</span>
						<span class="text-xl font-bold mono">{suite.config.iterations}</span>
					</div>
					<div class="text-center p-3 bg-bg border border-border rounded-[--radius-sm]">
						<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Total API Calls</span>
						<span class="text-xl font-bold text-accent mono">{$totalCalls}</span>
					</div>
				</div>

				{#if suite.config.judge_enabled}
					<div class="flex items-center gap-2 mb-4 px-3.5 py-2.5 bg-bg border border-border rounded-[--radius-sm] text-sm">
						<span class="text-text-muted font-semibold uppercase tracking-wide">Judge:</span>
						<span class="mono">{suite.config.judge_model}</span>
						<span class="text-text-dim ml-auto">
							{suite.config.judge_criteria?.join(', ')}
						</span>
					</div>
				{/if}

				<div class="flex flex-wrap gap-1.5 mb-5">
					{#each suite.models as model}
						<ModelChip {model} />
					{/each}
				</div>

				<button class="btn btn-primary w-full py-3 text-[15px]" onclick={startBenchmark}>
					Start Benchmark
				</button>
			</div>
		{:else}
			<!-- Running / completed -->
			<div class="flex items-center gap-3 mb-4">
				{#if $currentRun}
					<StatusBadge status={$currentRun.status} />
				{/if}
				{#if $runError}
					<span class="text-error text-[15px]">{$runError}</span>
				{/if}
			</div>

			<div class="mb-4">
				<div class="mb-1.5">
					<span class="text-sm text-text-muted uppercase tracking-wide">Benchmark</span>
				</div>
				<ProgressBar
					completed={$runProgress.completed}
					total={$runProgress.total}
					percent={$runProgress.percent}
				/>
			</div>

			{#if showJudgeProgress}
				<div class="mb-4">
					<div class="mb-1.5">
						<span class="text-sm text-text-muted uppercase tracking-wide">Judge Scoring</span>
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
				<div class="grid grid-cols-[repeat(auto-fill,minmax(240px,1fr))] gap-3 mb-6">
					{#each Object.entries($modelAggregates) as [model, agg]}
						<div class="card px-4 py-3.5">
							<div class="mb-2.5">
								<ModelChip {model} />
							</div>
							<div class="flex gap-4 mono">
								<div class="flex flex-col gap-0.5">
									<span class="text-xs text-text-dim uppercase tracking-wide">Avg Latency</span>
									<span class="text-base font-semibold">{agg.avgLatency.toFixed(0)}ms</span>
								</div>
								<div class="flex flex-col gap-0.5">
									<span class="text-xs text-text-dim uppercase tracking-wide">Avg tok/s</span>
									<span class="text-base font-semibold">{agg.avgTokensPerSecond.toFixed(1)}</span>
								</div>
								<div class="flex flex-col gap-0.5">
									<span class="text-xs text-text-dim uppercase tracking-wide">Success</span>
									<span class="text-base font-semibold">{agg.successCount}/{agg.totalCount}</span>
								</div>
							</div>
							{#if $modelJudgeAggregates[model]}
								<div class="flex flex-wrap gap-1 mt-2.5 pt-2 border-t border-border">
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
				<section class="mb-6">
					<h3 class="flex items-center gap-2.5 mb-3">
						<ModelChip {model} />
						<span class="text-sm text-text-dim mono">{results.length} Results</span>
					</h3>
					<div class="flex flex-col gap-2">
						{#each results as result (result.id)}
							<ResultCard {result} />
						{/each}
					</div>
				</section>
			{/each}

			<!-- Post-run actions -->
			{#if !$isRunning && $currentRun && ($currentRun.status === 'complete' || $currentRun.status === 'failed')}
				<div class="mt-6 pt-6 border-t border-border">
					<button class="btn btn-primary" onclick={() => goto(`/runs/${$currentRun?.id}`)}>
						View Run Details
					</button>
				</div>
			{/if}
		{/if}
	{/if}
</div>

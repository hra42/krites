<script lang="ts">
	import { onMount, mount, unmount, tick } from 'svelte';
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
	import RunPdfReport from '$lib/components/RunPdfReport.svelte';
	import { toastInfo, toastSuccess, toastError } from '$lib/stores/toast';
	import { renderMarkdown } from '$lib/utils/markdown';

	let { data } = $props();
	let run = $state<Run | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let exportingPdf = $state(false);

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

	function countKeyword(text: string, keyword: string): number {
		if (!keyword || !text) return 0;
		const needle = keyword.toLowerCase();
		const hay = text.toLowerCase();
		let count = 0;
		let idx = 0;
		while ((idx = hay.indexOf(needle, idx)) !== -1) {
			count++;
			idx += needle.length;
		}
		return count;
	}

	const keyword = $derived(run?.config.keyword?.trim() ?? '');
	const hasSummary = $derived(run?.summary && run.summary.models.length > 0);
	const hasJudge = $derived(run?.config.judge_enabled && run?.config.judge_criteria?.length);
	const criteria = $derived(run?.config.judge_criteria ?? []);
	const models = $derived(run?.results ? [...new Set(run.results.map((r) => r.model))] : []);

	type KeywordCell = { count: number; hasSuccess: boolean };
	type KeywordRow = { name: string; perModel: Map<string, KeywordCell> };

	const keywordMatrix = $derived.by<KeywordRow[] | null>(() => {
		if (!keyword || !run?.results) return null;
		const prompts = new Map<string, KeywordRow>();
		for (const r of run.results) {
			let row = prompts.get(r.prompt_id);
			if (!row) {
				row = { name: r.prompt_name || r.prompt_id, perModel: new Map() };
				prompts.set(r.prompt_id, row);
			}
			const cell = row.perModel.get(r.model) ?? { count: 0, hasSuccess: false };
			if (r.status === 'success' && r.response) {
				cell.count += countKeyword(r.response, keyword);
				cell.hasSuccess = true;
			}
			row.perModel.set(r.model, cell);
		}
		return Array.from(prompts.values());
	});

	async function exportPdf() {
		if (!run || exportingPdf) return;
		exportingPdf = true;
		toastInfo('Generating PDF…');

		const host = document.createElement('div');
		host.style.position = 'fixed';
		host.style.left = '-10000px';
		host.style.top = '0';
		host.style.width = '794px';
		host.style.background = '#ffffff';
		document.body.appendChild(host);

		let component: ReturnType<typeof mount> | null = null;
		try {
			component = mount(RunPdfReport, { target: host, props: { run } });
			// Allow charts (layerchart/d3) to render before rasterizing.
			await tick();
			await new Promise((r) => setTimeout(r, 600));

			const html2pdf = (await import('html2pdf.js')).default;
			const safeName = run.suite_name.replace(/[^a-z0-9-_]+/gi, '_');
			await html2pdf()
				.from(host.firstElementChild as HTMLElement)
				.set({
					filename: `${safeName}-${run.id}.pdf`,
					// [top, left, bottom, right] in mm — applied on every page by html2pdf.
					margin: [15, 12, 18, 12],
					image: { type: 'jpeg', quality: 0.98 },
					html2canvas: {
						scale: 2,
						useCORS: true,
						backgroundColor: '#ffffff',
						letterRendering: true,
						logging: false,
						removeContainer: true
					},
					jsPDF: { unit: 'mm', format: 'a4', orientation: 'portrait', compress: true },
					pagebreak: {
						mode: ['css', 'legacy'],
						// Targeted avoid list. avoid-all is unreliable past ~20 pages
						// (eKoopmans/html2pdf.js#227, #675), so we only protect atoms
						// that should never split mid-element.
						avoid: [
							'tr',
							'.chart-card',
							'.kpi',
							'.response-head',
							'.markdown-body p',
							'.markdown-body li',
							'.markdown-body h1',
							'.markdown-body h2',
							'.markdown-body h3',
							'.markdown-body h4',
							'.markdown-body pre',
							'.markdown-body blockquote',
							'.judge-row'
						]
					}
				})
				.save();
			toastSuccess('PDF downloaded');
		} catch (e) {
			console.error(e);
			toastError(e instanceof Error ? e.message : 'PDF export failed');
		} finally {
			if (component) unmount(component);
			host.remove();
			exportingPdf = false;
		}
	}
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
					<button
						type="button"
						onclick={exportPdf}
						disabled={exportingPdf}
						class="text-xs px-3 py-1 rounded-[6px] bg-bg-elevated border border-border text-text-muted transition-all duration-150 hover:border-accent hover:text-accent disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{exportingPdf ? 'Generating…' : 'PDF'}
					</button>
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

			{#if keyword && keywordMatrix && keywordMatrix.length > 0}
				<section class="mb-8">
					<h2 class="text-lg text-text-muted uppercase tracking-wide mb-3">
						Keyword summary — <span class="mono text-accent">"{keyword}"</span>
					</h2>
					<div class="card p-0 overflow-x-auto">
						<table class="w-full text-sm border-collapse">
							<thead>
								<tr class="text-left text-text-muted">
									<th class="px-4 py-3 border-b border-border font-normal">Prompt</th>
									{#each models as m}
										<th class="px-4 py-3 border-b border-border mono font-normal text-right">{m}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each keywordMatrix as row}
									<tr class="border-b border-border last:border-0">
										<td class="px-4 py-2.5">{row.name}</td>
										{#each models as m}
											{@const cell = row.perModel.get(m)}
											<td
												class="px-4 py-2.5 mono text-right {!cell || !cell.hasSuccess
													? 'text-text-dim'
													: cell.count > 0
														? 'text-accent font-semibold'
														: 'text-text-dim'}"
											>
												{!cell || !cell.hasSuccess ? '—' : cell.count}
											</td>
										{/each}
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</section>
			{/if}

			<!-- Responses -->
			<section class="mb-8">
				<h2 class="text-lg text-text-muted uppercase tracking-wide mb-3">Responses</h2>
				<div class="flex flex-col gap-4">
					{#each run.results as result}
						<article class="card p-5">
							<header class="flex flex-wrap items-center justify-between gap-3 mb-3 pb-3 border-b border-border">
								<div class="flex items-center gap-2 flex-wrap">
									<ModelChip model={result.model} />
									<span class="text-sm text-text-muted">·</span>
									<span class="text-[15px]">{result.prompt_name || result.prompt_id}</span>
									<span class="text-sm text-text-muted">·</span>
									<span class="text-sm text-text-muted mono">iter {result.iteration}</span>
								</div>
								<div class="flex items-center gap-2 text-sm text-text-muted mono flex-wrap justify-end">
									<span class="inline-block w-2 h-2 rounded-full {statusDotColor[result.status] || 'bg-text-muted'}"></span>
									<span class="uppercase tracking-wide text-xs">{result.status}</span>
									<span class="text-text-dim">·</span>
									<span title="time to first byte">TTFB {result.metrics.ttfb_ms.toFixed(0)}ms</span>
									<span class="text-text-dim">·</span>
									<span>{result.metrics.total_latency_ms.toFixed(0)}ms</span>
									<span class="text-text-dim">·</span>
									<span>{result.metrics.tokens_per_second.toFixed(1)} tok/s</span>
									<span class="text-text-dim">·</span>
									<span>{result.metrics.completion_tokens} tok</span>
									<span class="text-text-dim">·</span>
									<span>${result.metrics.estimated_cost.toFixed(4)}</span>
								</div>
							</header>
							{#if result.status !== 'success'}
								<pre class="bg-bg-elevated border border-error/40 border-l-4 border-l-error rounded-[--radius] p-3 text-sm text-error font-mono whitespace-pre-wrap break-words m-0">{result.error || '(no error message)'}</pre>
							{:else if result.response}
								{#if keyword}
									{@const kc = countKeyword(result.response, keyword)}
									<div class="mb-2 text-sm text-text-muted">
										Keyword <span class="mono text-accent">"{keyword}"</span>:
										<span class="mono font-semibold {kc > 0 ? 'text-accent' : 'text-text-dim'}">{kc}</span>
										<span class="text-text-dim">{kc === 1 ? 'occurrence' : 'occurrences'}</span>
									</div>
								{/if}
								<div class="markdown bg-bg-elevated border border-border border-l-4 border-l-accent rounded-[--radius] p-4">
									{@html renderMarkdown(result.response)}
								</div>
							{:else}
								<pre class="bg-bg-elevated border border-border border-l-4 border-l-accent rounded-[--radius] p-3 text-sm font-mono whitespace-pre-wrap break-words m-0 text-text-dim">(empty response)</pre>
							{/if}
							{#if result.judge_scores?.length}
								<div class="mt-3 pt-3 border-t border-border flex flex-col gap-2">
									{#each result.judge_scores as js}
										<div>
											<div class="flex items-baseline justify-between mb-0.5">
												<span class="text-sm font-semibold capitalize">{js.criterion}</span>
												<span class="mono text-base text-accent font-semibold">{js.score}<span class="text-text-dim font-normal">/10</span></span>
											</div>
											{#if js.explanation}
												<p class="text-sm text-text-muted m-0 leading-relaxed">{js.explanation}</p>
											{/if}
										</div>
									{/each}
								</div>
							{/if}
						</article>
					{/each}
				</div>
			</section>
		{:else}
			<div class="card text-center py-10 text-text-muted">
				<p>No results available.</p>
			</div>
		{/if}
	{/if}
</div>

<style>
	/* Rendered markdown inside response cards. Mirrors app's dark theme. */
	.markdown :global(> *:first-child) {
		margin-top: 0;
	}
	.markdown :global(> *:last-child) {
		margin-bottom: 0;
	}
	.markdown :global(h1),
	.markdown :global(h2),
	.markdown :global(h3),
	.markdown :global(h4),
	.markdown :global(h5),
	.markdown :global(h6) {
		font-weight: 600;
		line-height: 1.25;
		margin: 1em 0 0.5em 0;
		color: var(--color-text);
	}
	.markdown :global(h1) {
		font-size: 1.4em;
	}
	.markdown :global(h2) {
		font-size: 1.25em;
	}
	.markdown :global(h3) {
		font-size: 1.1em;
	}
	.markdown :global(h4),
	.markdown :global(h5),
	.markdown :global(h6) {
		font-size: 1em;
	}
	.markdown :global(p) {
		margin: 0.6em 0;
		line-height: 1.6;
	}
	.markdown :global(ul),
	.markdown :global(ol) {
		margin: 0.6em 0;
		padding-left: 1.5em;
	}
	.markdown :global(li) {
		margin: 0.25em 0;
		line-height: 1.55;
	}
	.markdown :global(li > ul),
	.markdown :global(li > ol) {
		margin: 0.25em 0;
	}
	.markdown :global(strong) {
		font-weight: 600;
		color: var(--color-text);
	}
	.markdown :global(em) {
		font-style: italic;
	}
	.markdown :global(a) {
		color: var(--color-accent);
		text-decoration: underline;
		text-underline-offset: 2px;
	}
	.markdown :global(code) {
		font-family: 'JetBrains Mono', ui-monospace, monospace;
		font-size: 0.9em;
		padding: 1px 5px;
		background: rgba(167, 139, 250, 0.12);
		border-radius: 4px;
		color: var(--color-accent);
	}
	.markdown :global(pre) {
		font-family: 'JetBrains Mono', ui-monospace, monospace;
		font-size: 0.9em;
		background: rgba(0, 0, 0, 0.25);
		border: 1px solid var(--color-border);
		border-radius: 6px;
		padding: 10px 12px;
		margin: 0.8em 0;
		overflow-x: auto;
		white-space: pre;
	}
	.markdown :global(pre code) {
		background: transparent;
		padding: 0;
		color: inherit;
		font-size: inherit;
	}
	.markdown :global(blockquote) {
		margin: 0.6em 0;
		padding: 0.2em 0 0.2em 1em;
		border-left: 3px solid var(--color-border);
		color: var(--color-text-muted);
	}
	.markdown :global(hr) {
		border: none;
		border-top: 1px solid var(--color-border);
		margin: 1em 0;
	}
	.markdown :global(table) {
		border-collapse: collapse;
		margin: 0.8em 0;
		font-size: 0.95em;
	}
	.markdown :global(th),
	.markdown :global(td) {
		border: 1px solid var(--color-border);
		padding: 6px 10px;
		text-align: left;
	}
	.markdown :global(th) {
		background: rgba(255, 255, 255, 0.04);
		font-weight: 600;
	}
	.markdown :global(img) {
		max-width: 100%;
		height: auto;
		border-radius: 4px;
	}
</style>

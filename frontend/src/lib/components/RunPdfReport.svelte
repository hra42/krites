<script lang="ts">
	import type { Run, Result } from '$lib/types';
	import LatencyChart from '$lib/components/charts/LatencyChart.svelte';
	import JudgeRadarChart from '$lib/components/charts/JudgeRadarChart.svelte';
	import CostChart from '$lib/components/charts/CostChart.svelte';
	import IterationLineChart from '$lib/components/charts/IterationLineChart.svelte';

	interface Props {
		run: Run;
	}

	let { run }: Props = $props();

	const hasSummary = $derived(!!run.summary && run.summary.models.length > 0);
	const hasJudge = $derived(!!run.config.judge_enabled && (run.config.judge_criteria?.length ?? 0) > 0);
	const criteria = $derived(run.config.judge_criteria ?? []);
	const models = $derived(run.results ? [...new Set(run.results.map((r) => r.model))] : []);

	const totals = $derived.by(() => {
		const results = run.results ?? [];
		const success = results.filter((r) => r.status === 'success').length;
		const totalCost = results.reduce((s, r) => s + (r.metrics?.estimated_cost ?? 0), 0);
		const totalTokens = results.reduce(
			(s, r) => s + (r.metrics?.prompt_tokens ?? 0) + (r.metrics?.completion_tokens ?? 0),
			0
		);
		const successRate = results.length ? (success / results.length) * 100 : 0;
		return { count: results.length, success, totalCost, totalTokens, successRate };
	});

	function formatDateTime(dateStr?: string): string {
		if (!dateStr) return '—';
		return new Date(dateStr).toLocaleString('en-GB', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatDate(dateStr?: string): string {
		if (!dateStr) return '—';
		return new Date(dateStr).toLocaleDateString('en-GB', {
			day: '2-digit',
			month: 'long',
			year: 'numeric'
		});
	}

	function duration(start?: string, end?: string): string {
		if (!start || !end) return '—';
		const ms = new Date(end).getTime() - new Date(start).getTime();
		if (ms < 1000) return `${ms} ms`;
		if (ms < 60000) return `${(ms / 1000).toFixed(1)} s`;
		const m = Math.floor(ms / 60000);
		const s = Math.round((ms % 60000) / 1000);
		return `${m}m ${s}s`;
	}

	function shortModel(model: string): string {
		return model.split('/').pop() ?? model;
	}

	function modelOrg(model: string): string {
		const parts = model.split('/');
		return parts.length > 1 ? parts[0] : '';
	}

	function fmtMs(ms: number): string {
		if (ms >= 1000) return `${(ms / 1000).toFixed(2)} s`;
		return `${ms.toFixed(0)} ms`;
	}

	function fmtCost(cost: number): string {
		if (cost === 0) return '$0';
		if (cost < 0.0001) return `$${cost.toExponential(2)}`;
		return `$${cost.toFixed(4)}`;
	}

	function fmtPct(v: number): string {
		return `${v.toFixed(0)}%`;
	}

	function resultLabel(r: Result): string {
		return `${shortModel(r.model)} · ${r.prompt_name || r.prompt_id} · iteration ${r.iteration}`;
	}

	const generatedAt = new Date();
</script>

<div class="pdf">
	<!-- Cover -->
	<section class="cover">
		<div class="brand">KRITES · BENCHMARK REPORT</div>
		<h1>{run.suite_name}</h1>
		<p class="subtitle">
			Generated {formatDate(generatedAt.toISOString())} · Run ID
			<span class="mono">{run.id}</span>
		</p>

		<div class="kpis">
			<div class="kpi">
				<div class="kpi-label">Duration</div>
				<div class="kpi-value">{duration(run.started_at, run.ended_at)}</div>
			</div>
			<div class="kpi">
				<div class="kpi-label">Results</div>
				<div class="kpi-value">{totals.count}</div>
				<div class="kpi-sub">{totals.success} successful · {fmtPct(totals.successRate)}</div>
			</div>
			<div class="kpi">
				<div class="kpi-label">Total cost</div>
				<div class="kpi-value">{fmtCost(totals.totalCost)}</div>
				<div class="kpi-sub">{totals.totalTokens.toLocaleString()} tokens</div>
			</div>
			<div class="kpi">
				<div class="kpi-label">Models</div>
				<div class="kpi-value">{models.length}</div>
				<div class="kpi-sub">{run.config.iterations} iter · concurrency {run.config.concurrency}</div>
			</div>
		</div>

		<div class="meta-grid">
			<div class="meta-block">
				<h4>Run</h4>
				<dl>
					<dt>Status</dt>
					<dd><span class="status status-{run.status}">{run.status}</span></dd>
					<dt>Started</dt>
					<dd class="mono">{formatDateTime(run.started_at)}</dd>
					<dt>Ended</dt>
					<dd class="mono">{formatDateTime(run.ended_at)}</dd>
				</dl>
			</div>
			<div class="meta-block">
				<h4>Configuration</h4>
				<dl>
					<dt>Temperature</dt>
					<dd class="mono">{run.config.temperature}</dd>
					<dt>Max tokens</dt>
					<dd class="mono">{run.config.max_tokens}</dd>
					<dt>Top-p</dt>
					<dd class="mono">{run.config.top_p}</dd>
					<dt>Timeout</dt>
					<dd class="mono">{run.config.timeout_seconds}s</dd>
				</dl>
			</div>
			<div class="meta-block">
				<h4>Judge</h4>
				<dl>
					<dt>Enabled</dt>
					<dd>{run.config.judge_enabled ? 'Yes' : 'No'}</dd>
					{#if run.config.judge_enabled}
						<dt>Model</dt>
						<dd class="mono">{run.config.judge_model ?? '—'}</dd>
						<dt>Criteria</dt>
						<dd>{criteria.join(', ') || '—'}</dd>
					{/if}
				</dl>
			</div>
		</div>

		<div class="models-list">
			<h4>Models tested</h4>
			<ul>
				{#each models as m}
					<li>
						<span class="mono">{shortModel(m)}</span>
						{#if modelOrg(m)}<span class="muted"> · {modelOrg(m)}</span>{/if}
					</li>
				{/each}
			</ul>
		</div>
	</section>

	{#if hasSummary}
		<!-- Model comparison -->
		<section class="page-break">
			<header class="section-head">
				<h2>Model comparison</h2>
				<p class="section-sub">Aggregated metrics across all prompts and iterations.</p>
			</header>
			<table class="data">
				<thead>
					<tr>
						<th class="left">Model</th>
						<th>Avg latency</th>
						<th>P50</th>
						<th>P95</th>
						<th>tok/s</th>
						<th>Success</th>
						<th>Avg cost</th>
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
							<td class="left">
								<div class="cell-model">
									<span class="model-name mono">{shortModel(ms.model)}</span>
									{#if modelOrg(ms.model)}<span class="muted"> · {modelOrg(ms.model)}</span>{/if}
								</div>
							</td>
							<td class="mono">{fmtMs(ms.avg_latency_ms)}</td>
							<td class="mono">{fmtMs(ms.p50_latency_ms)}</td>
							<td class="mono">{fmtMs(ms.p95_latency_ms)}</td>
							<td class="mono">{ms.avg_tokens_per_second.toFixed(1)}</td>
							<td class="mono">{fmtPct(ms.success_rate * 100)}</td>
							<td class="mono">{fmtCost(ms.avg_cost)}</td>
							<td class="mono strong">{fmtCost(ms.total_cost)}</td>
							{#if hasJudge}
								{#each criteria as c}
									<td class="mono">{ms.avg_judge_scores?.[c]?.toFixed(2) ?? '—'}</td>
								{/each}
							{/if}
						</tr>
					{/each}
				</tbody>
			</table>
		</section>

		<!-- Charts -->
		<section class="page-break">
			<header class="section-head">
				<h2>Visualisations</h2>
				<p class="section-sub">Visual breakdown of performance and cost.</p>
			</header>
			<div class="chart-stack">
				<figure class="chart-card">
					<figcaption>Latency comparison</figcaption>
					<div class="chart-box chart-box-tall">
						<LatencyChart modelSummaries={run.summary!.models} pdfMode />
					</div>
				</figure>
				<figure class="chart-card">
					<figcaption>Cost comparison</figcaption>
					<div
						class="chart-box"
						style:height="{Math.max(220, run.summary!.models.length * 36 + 80)}px"
					>
						<CostChart modelSummaries={run.summary!.models} pdfMode />
					</div>
				</figure>
				{#if hasJudge}
					<figure class="chart-card">
						<figcaption>Judge scores</figcaption>
						<div class="chart-box chart-box-tall radar-box">
							<JudgeRadarChart modelSummaries={run.summary!.models} {criteria} pdfMode />
						</div>
					</figure>
				{/if}
				{#if run.config.iterations > 1 && run.results}
					<figure class="chart-card">
						<figcaption>Latency per iteration</figcaption>
						<div class="chart-box chart-box-tall">
							<IterationLineChart results={run.results} {models} pdfMode />
						</div>
					</figure>
				{/if}
			</div>
		</section>
	{/if}

	{#if run.results && run.results.length > 0}
		<!-- Responses -->
		<section class="page-break">
			<header class="section-head">
				<h2>Responses</h2>
				<p class="section-sub">Full response text for every result.</p>
			</header>
			{#each run.results as r}
				<article class="response">
					<header class="response-head">
						<h3>{resultLabel(r)}</h3>
						<div class="response-meta">
							<span class="dot dot-{r.status}"></span>
							<span>{r.status}</span>
							<span class="sep">·</span>
							<span class="mono">TTFB {fmtMs(r.metrics.ttfb_ms)}</span>
							<span class="sep">·</span>
							<span class="mono">{fmtMs(r.metrics.total_latency_ms)}</span>
							<span class="sep">·</span>
							<span class="mono">{r.metrics.tokens_per_second.toFixed(1)} tok/s</span>
							<span class="sep">·</span>
							<span class="mono">{r.metrics.completion_tokens} tok</span>
							<span class="sep">·</span>
							<span class="mono">{fmtCost(r.metrics.estimated_cost)}</span>
						</div>
					</header>
					{#if r.status !== 'success'}
						<pre class="response-body error-body">{r.error || '(no error message)'}</pre>
					{:else}
						<pre class="response-body">{r.response || '(empty response)'}</pre>
					{/if}
					{#if r.judge_scores?.length}
						<div class="judge-block">
							{#each r.judge_scores as js}
								<div class="judge-row">
									<div class="judge-row-head">
										<span class="judge-criterion">{js.criterion}</span>
										<span class="judge-score mono">{js.score}<span class="judge-max">/10</span></span>
									</div>
									{#if js.explanation}
										<p class="judge-explanation">{js.explanation}</p>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</article>
			{/each}
		</section>
	{/if}

	<footer class="report-footer">
		Krites · {formatDateTime(generatedAt.toISOString())}
	</footer>
</div>

<style>
	/* ---------- shell ---------- */
	.pdf {
		/* A4 portrait at 96dpi minus 12mm L/R margins ≈ 794 - 90 ≈ 704 */
		width: 704px;
		background: #ffffff;
		color: #1a1a1a;
		font-family: 'Outfit', system-ui, -apple-system, 'Helvetica Neue', sans-serif;
		font-size: 11px;
		line-height: 1.5;
		padding: 0;
		box-sizing: border-box;
		-webkit-font-smoothing: antialiased;
	}

	.pdf :global(*) {
		box-sizing: border-box;
	}

	/* ---------- chart overrides ----------
	   html2canvas drops Tailwind v4 preflight (uses :where() + oklch) and
	   falls back to user-agent styles, which puts a 2px outset border on
	   <button> elements. Layerchart's legend uses <button>, so we have to
	   reset it explicitly here. We also keep axis text in the body font
	   (Outfit) so html2canvas rasterizes it cleanly — the previous mono
	   override was being rendered as a fallback bitmap font. */
	.pdf :global(.lc-legend-container) {
		background: transparent !important;
		color: #1a1a1a !important;
		font-family: 'Outfit', system-ui, sans-serif !important;
		font-size: 9px !important;
	}
	.pdf :global(.lc-legend-swatch-button) {
		all: unset !important;
		display: inline-flex !important;
		align-items: center !important;
		gap: 6px !important;
		padding: 0 8px 0 0 !important;
		font-size: 9px !important;
		color: #555 !important;
		background: transparent !important;
		border: none !important;
		cursor: default !important;
	}
	.pdf :global(.lc-legend-swatch) {
		width: 8px !important;
		height: 8px !important;
		border-radius: 50% !important;
		border: none !important;
		flex-shrink: 0 !important;
	}
	.pdf :global(.lc-legend-swatch-label) {
		color: #555 !important;
		font-family: 'Outfit', system-ui, sans-serif !important;
	}
	.pdf :global(svg) {
		color-scheme: light !important;
		color: #6a6a6a !important;
	}
	.pdf :global(svg text) {
		fill: #6a6a6a !important;
		stroke: none !important;
		font-family: 'Outfit', system-ui, sans-serif !important;
		font-size: 9px !important;
		font-weight: 400 !important;
		paint-order: normal !important;
	}
	/* layerchart specifically applies a 2px stroke matching --color-surface-100,
	   which falls back to black in dark color-scheme. Force it off. */
	.pdf :global(.lc-axis-tick-label),
	.pdf :global(.lc-axis-label) {
		fill: #6a6a6a !important;
		stroke: #ffffff !important;
		stroke-width: 0 !important;
		font-weight: 400 !important;
	}
	.pdf :global(.lc-axis-tick),
	.pdf :global(.lc-axis-rule),
	.pdf :global(.lc-axis-grid),
	.pdf :global(.tick line),
	.pdf :global(.domain) {
		stroke: #d8d8d8 !important;
		--stroke-color: #d8d8d8 !important;
	}

	.mono {
		font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace;
		font-feature-settings: 'tnum' 1;
	}

	.muted {
		color: #8a8a8a;
		font-weight: 400;
	}

	.strong {
		font-weight: 600;
	}

	section {
		margin-bottom: 32px;
	}

	/* Top breathing room on every section that starts a new page. */
	.page-break {
		page-break-before: always;
		break-before: page;
		padding-top: 8px;
	}

	.cover {
		padding-top: 4px;
	}

	/* ---------- cover ---------- */
	.brand {
		font-family: 'JetBrains Mono', ui-monospace, monospace;
		font-size: 9px;
		letter-spacing: 0.18em;
		color: #7c3aed;
		font-weight: 600;
		margin-bottom: 16px;
	}

	h1 {
		font-size: 32px;
		font-weight: 600;
		margin: 0 0 8px 0;
		letter-spacing: -0.02em;
		color: #0a0a0a;
		line-height: 1.15;
	}

	.subtitle {
		color: #6a6a6a;
		font-size: 11px;
		margin: 0 0 32px 0;
	}

	.subtitle .mono {
		font-size: 10px;
		color: #555;
	}

	/* ---------- KPIs ---------- */
	.kpis {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 12px;
		margin-bottom: 28px;
		padding-top: 20px;
		border-top: 2px solid #0a0a0a;
	}

	.kpi {
		padding: 14px 14px 12px;
		background: #fafafa;
		border: 1px solid #ececec;
		border-radius: 6px;
		page-break-inside: avoid;
	}

	.kpi-label {
		font-size: 9px;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: #888;
		margin-bottom: 6px;
	}

	.kpi-value {
		font-family: 'JetBrains Mono', ui-monospace, monospace;
		font-size: 22px;
		font-weight: 600;
		color: #0a0a0a;
		line-height: 1.1;
		font-feature-settings: 'tnum' 1;
	}

	.kpi-sub {
		margin-top: 4px;
		font-size: 10px;
		color: #888;
	}

	/* ---------- meta grid ---------- */
	.meta-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 24px;
		margin-bottom: 28px;
	}

	.meta-block h4,
	.models-list h4 {
		font-size: 10px;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: #888;
		margin: 0 0 8px 0;
		font-weight: 600;
	}

	dl {
		margin: 0;
		display: grid;
		grid-template-columns: 1fr auto;
		gap: 4px 12px;
	}

	dt {
		color: #6a6a6a;
		font-size: 10px;
	}

	dd {
		margin: 0;
		font-size: 10px;
		color: #1a1a1a;
		text-align: right;
	}

	dd.mono {
		font-size: 10px;
	}

	/* ---------- status pill ---------- */
	.status {
		display: inline-block;
		padding: 1px 8px;
		border-radius: 999px;
		font-size: 9px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.status-complete {
		background: #e6f7ee;
		color: #0a7a3b;
	}

	.status-failed {
		background: #fdeaea;
		color: #b42318;
	}

	.status-running {
		background: #e6f0fd;
		color: #1f5fbf;
	}

	.status-canceled,
	.status-pending {
		background: #f0f0f0;
		color: #555;
	}

	/* ---------- models list ---------- */
	.models-list {
		padding-top: 16px;
		border-top: 1px solid #ececec;
	}

	.models-list ul {
		margin: 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.models-list li {
		font-size: 10px;
		padding: 4px 10px;
		background: #fafafa;
		border: 1px solid #ececec;
		border-radius: 4px;
	}

	/* ---------- section headings ---------- */
	.section-head {
		margin-bottom: 16px;
		padding-bottom: 10px;
		border-bottom: 2px solid #0a0a0a;
	}

	h2 {
		font-size: 18px;
		font-weight: 600;
		margin: 0 0 4px 0;
		letter-spacing: -0.01em;
		color: #0a0a0a;
	}

	.section-sub {
		margin: 0;
		color: #6a6a6a;
		font-size: 10px;
	}

	/* ---------- data tables ---------- */
	table.data {
		width: 100%;
		border-collapse: collapse;
		font-size: 10px;
	}

	table.data th {
		text-align: right;
		font-weight: 600;
		font-size: 9px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: #888;
		padding: 8px 10px;
		border-bottom: 1px solid #d0d0d0;
		white-space: nowrap;
	}

	table.data th.left,
	table.data td.left {
		text-align: left;
	}

	table.data td {
		padding: 9px 10px;
		text-align: right;
		border-bottom: 1px solid #f0f0f0;
		color: #1a1a1a;
		vertical-align: middle;
		white-space: nowrap;
	}

	table.data tbody tr:nth-child(even) td {
		background: #fafafa;
	}

	.cell-model {
		display: flex;
		align-items: baseline;
		gap: 4px;
	}

	.model-name {
		font-weight: 600;
		color: #0a0a0a;
	}

	/* ---------- status dots ---------- */
	.dot {
		display: inline-block;
		width: 6px;
		height: 6px;
		border-radius: 50%;
		margin-right: 4px;
		vertical-align: middle;
	}
	.dot-success {
		background: #16a34a;
	}
	.dot-error {
		background: #dc2626;
	}
	.dot-timeout {
		background: #d97706;
	}

	.status-text {
		font-size: 9px;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: #555;
	}

	/* ---------- judge pills ---------- */
	.judge-pills {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}

	.pill {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 1px 6px;
		background: #f3f0fb;
		border: 1px solid #e6e0f5;
		border-radius: 999px;
		font-size: 9px;
		color: #4c1d95;
	}

	.pill-label {
		text-transform: lowercase;
	}

	.pill-value {
		font-weight: 600;
		color: #5b21b6;
	}

	/* ---------- charts ---------- */
	.chart-stack {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}

	.chart-card {
		margin: 0;
		padding: 16px 18px 18px;
		background: #ffffff;
		border: 1px solid #ececec;
		border-radius: 8px;
		page-break-inside: avoid;
		break-inside: avoid;
	}

	.chart-card figcaption {
		font-size: 10px;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: #888;
		font-weight: 600;
		margin-bottom: 10px;
	}

	.chart-box {
		width: 100%;
		height: 260px;
	}

	.chart-box-tall {
		height: 320px;
	}

	.radar-box {
		display: flex;
		justify-content: center;
		align-items: center;
		height: 360px;
	}

	/* ---------- responses ---------- */
	.response {
		margin-bottom: 18px;
		padding-bottom: 16px;
		border-bottom: 1px solid #ececec;
		page-break-inside: auto;
	}

	.response:last-child {
		border-bottom: none;
	}

	.response-head {
		margin-bottom: 8px;
		page-break-after: avoid;
	}

	h3 {
		font-size: 12px;
		margin: 0 0 4px 0;
		font-weight: 600;
		color: #0a0a0a;
	}

	.response-meta {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 10px;
		color: #6a6a6a;
	}

	.response-meta .sep {
		color: #ccc;
	}

	.response-body {
		font-family: 'JetBrains Mono', ui-monospace, monospace;
		font-size: 9.5px;
		line-height: 1.55;
		background: #fafafa;
		border: 1px solid #ececec;
		border-left: 3px solid #7c3aed;
		border-radius: 4px;
		padding: 10px 12px;
		white-space: pre-wrap;
		word-break: break-word;
		margin: 0;
		color: #1a1a1a;
	}

	.error-body {
		background: #fdf6f6;
		border-color: #f5d6d6;
		border-left-color: #dc2626;
		color: #8a1f1f;
	}

	.judge-block {
		margin-top: 10px;
		padding: 10px 12px;
		background: #fafafa;
		border: 1px solid #ececec;
		border-radius: 4px;
	}

	.judge-row {
		font-size: 10px;
	}

	.judge-row + .judge-row {
		margin-top: 8px;
		padding-top: 8px;
		border-top: 1px dashed #e0e0e0;
	}

	.judge-row-head {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 2px;
	}

	.judge-criterion {
		font-weight: 600;
		color: #0a0a0a;
		text-transform: capitalize;
	}

	.judge-score {
		font-size: 11px;
		font-weight: 600;
		color: #5b21b6;
	}

	.judge-max {
		color: #b8a8d8;
		font-weight: 400;
	}

	.judge-explanation {
		margin: 0;
		color: #555;
		font-size: 9.5px;
		line-height: 1.5;
	}

	/* ---------- footer ---------- */
	.report-footer {
		margin-top: 32px;
		padding-top: 12px;
		border-top: 1px solid #ececec;
		font-size: 9px;
		color: #aaa;
		text-align: center;
		letter-spacing: 0.04em;
	}
</style>

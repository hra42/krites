<script lang="ts">
	import type { Result } from '$lib/types';
	import JudgeScoreBadge from '$lib/components/JudgeScoreBadge.svelte';

	interface Props {
		result: Result;
	}

	let { result }: Props = $props();

	function truncate(text: string, maxLen: number): string {
		if (!text) return '';
		if (text.length <= maxLen) return text;
		return text.slice(0, maxLen) + '...';
	}
</script>

<div class="result-card" class:error={result.status !== 'success'}>
	<div class="result-header">
		<span class="prompt-name">{result.prompt_name || 'Prompt'}</span>
		<span class="iteration mono">#{result.iteration}</span>
		<span class="status-dot {result.status}"></span>
	</div>
	{#if result.status === 'success'}
		<div class="metrics mono">
			<span class="metric">{result.metrics.total_latency_ms.toFixed(0)}ms</span>
			<span class="metric">{result.metrics.tokens_per_second.toFixed(1)} tok/s</span>
			<span class="metric">{result.metrics.completion_tokens} tokens</span>
		</div>
		{#if result.judge_scores?.length}
			<div class="judge-badges">
				{#each result.judge_scores as score}
					<JudgeScoreBadge criterion={score.criterion} score={score.score} />
				{/each}
			</div>
		{/if}
		<pre class="response-preview mono">{truncate(result.response, 300)}</pre>
	{:else}
		<div class="error-msg mono">{result.error || result.status}</div>
	{/if}
</div>

<style>
	.result-card {
		background: var(--color-bg-card);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		padding: 12px 16px;
		animation: fadeIn 0.2s ease;
	}

	.result-card.error {
		border-color: var(--color-error);
	}

	.result-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 8px;
	}

	.prompt-name {
		font-weight: 500;
		font-size: 13px;
	}

	.iteration {
		font-size: 12px;
		color: var(--color-text-dim);
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		margin-left: auto;
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

	.metrics {
		display: flex;
		gap: 16px;
		margin-bottom: 8px;
		font-size: 12px;
	}

	.metric {
		color: var(--color-text-muted);
	}

	.judge-badges {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin-bottom: 8px;
	}

	.response-preview {
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		padding: 8px 10px;
		font-size: 11px;
		white-space: pre-wrap;
		word-break: break-word;
		color: var(--color-text-dim);
		max-height: 80px;
		overflow: hidden;
	}

	.error-msg {
		color: var(--color-error);
		font-size: 12px;
	}
</style>

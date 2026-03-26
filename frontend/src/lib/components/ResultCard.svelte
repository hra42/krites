<script lang="ts">
	import type { Result } from '$lib/types';
	import JudgeScoreBadge from '$lib/components/JudgeScoreBadge.svelte';

	interface Props {
		result: Result;
	}

	let { result }: Props = $props();

	const statusDotColor: Record<string, string> = {
		success: 'bg-success',
		error: 'bg-error',
		timeout: 'bg-warning'
	};

	function truncate(text: string, maxLen: number): string {
		if (!text) return '';
		if (text.length <= maxLen) return text;
		return text.slice(0, maxLen) + '...';
	}
</script>

<div class="bg-bg-card border rounded-[--radius] px-4 py-3 fade-in {result.status !== 'success' ? 'border-error' : 'border-border'}">
	<div class="flex items-center gap-2 mb-2">
		<span class="font-medium text-base">{result.prompt_name || 'Prompt'}</span>
		<span class="text-sm text-text-dim mono">#{result.iteration}</span>
		<span class="w-2 h-2 rounded-full ml-auto {statusDotColor[result.status] || 'bg-text-muted'}"></span>
	</div>
	{#if result.status === 'success'}
		<div class="flex gap-4 mb-2 text-sm text-text-muted mono">
			<span>{result.metrics.total_latency_ms.toFixed(0)}ms</span>
			<span>{result.metrics.tokens_per_second.toFixed(1)} tok/s</span>
			<span>{result.metrics.completion_tokens} tokens</span>
		</div>
		{#if result.judge_scores?.length}
			<div class="flex flex-wrap gap-1 mb-2">
				{#each result.judge_scores as score}
					<JudgeScoreBadge criterion={score.criterion} score={score.score} />
				{/each}
			</div>
		{/if}
		<pre class="bg-bg border border-border rounded-[--radius-sm] px-3 py-2.5 text-sm whitespace-pre-wrap break-words text-text-dim max-h-24 overflow-hidden mono">{truncate(result.response, 300)}</pre>
	{:else}
		<div class="text-error text-sm mono">{result.error || result.status}</div>
	{/if}
</div>

<script lang="ts">
	import type { SuiteSummary } from '$lib/types';

	interface Props {
		suite: SuiteSummary;
	}

	let { suite }: Props = $props();

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric'
		});
	}
</script>

<a href="/suites/{suite.id}" class="card fade-in">
	<div class="header">
		<h3>{suite.name}</h3>
	</div>
	{#if suite.description}
		<p class="description">{suite.description}</p>
	{/if}
	<div class="meta">
		<span class="meta-item mono">{suite.model_count} Models</span>
		<span class="dot">&middot;</span>
		<span class="meta-item mono">{suite.prompt_count} Prompts</span>
		<span class="dot">&middot;</span>
		<span class="meta-item">{formatDate(suite.created_at)}</span>
	</div>
</a>

<style>
	.card {
		display: block;
		background: var(--color-bg-card);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: 20px;
		transition: all 0.15s ease;
		color: var(--color-text);
	}

	.card:hover {
		border-color: var(--color-accent);
		transform: translateY(-1px);
	}

	.header {
		margin-bottom: 8px;
	}

	h3 {
		font-size: 16px;
		font-weight: 600;
	}

	.description {
		color: var(--color-text-muted);
		font-size: 13px;
		margin-bottom: 12px;
		line-height: 1.4;
	}

	.meta {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 12px;
		color: var(--color-text-dim);
	}

	.meta-item {
		font-size: 12px;
	}

	.dot {
		color: var(--color-text-dim);
	}
</style>

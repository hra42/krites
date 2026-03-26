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

<a href="/suites/{suite.id}" class="block bg-bg-card border border-border rounded-[--radius-lg] p-5 transition-all duration-150 text-text hover:border-accent hover:-translate-y-px fade-in">
	<div class="mb-2">
		<h3 class="text-lg font-semibold">{suite.name}</h3>
	</div>
	{#if suite.description}
		<p class="text-text-muted text-base mb-3 leading-[1.4]">{suite.description}</p>
	{/if}
	<div class="flex items-center gap-2 text-sm text-text-dim">
		<span class="mono">{suite.model_count} Models</span>
		<span>&middot;</span>
		<span class="mono">{suite.prompt_count} Prompts</span>
		<span>&middot;</span>
		<span>{formatDate(suite.created_at)}</span>
	</div>
</a>

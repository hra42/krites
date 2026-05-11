<script lang="ts">
	import type { SuiteSummary } from '$lib/types';

	interface Props {
		suite: SuiteSummary;
		onduplicate?: (suite: SuiteSummary) => void;
	}

	let { suite, onduplicate }: Props = $props();

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric'
		});
	}

	function handleDuplicate(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		onduplicate?.(suite);
	}
</script>

<a href="/suites/{suite.id}" class="relative block bg-bg-card border border-border rounded-[--radius-lg] p-5 transition-all duration-150 text-text hover:border-accent hover:-translate-y-px fade-in">
	{#if onduplicate}
		<button
			type="button"
			onclick={handleDuplicate}
			class="absolute top-3 right-3 px-2 py-1 text-xs text-text-muted hover:text-accent hover:bg-bg-elevated rounded-[--radius-sm] transition-colors mono"
			title="Duplicate suite"
			aria-label="Duplicate suite"
		>
			Copy
		</button>
	{/if}
	<div class="mb-2 pr-12">
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

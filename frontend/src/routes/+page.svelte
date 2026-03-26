<script lang="ts">
	import { onMount } from 'svelte';
	import { suites, loadSuites } from '$lib/stores/benchmark';
	import { getAnalyticsOverview } from '$lib/api/client';
	import type { AnalyticsOverview } from '$lib/types';

	let overview = $state<AnalyticsOverview | null>(null);

	onMount(async () => {
		loadSuites();
		try {
			overview = await getAnalyticsOverview();
		} catch {
			// analytics not available yet
		}
	});
</script>

<div>
	<h1 class="text-3xl mb-1">Dashboard</h1>
	<p class="text-text-muted mb-8">LLM Benchmark Platform</p>

	<div class="grid grid-cols-3 gap-4 mb-8">
		<div class="card flex flex-col gap-1 text-center p-6">
			<span class="text-[32px] font-bold text-accent mono">{$suites.length}</span>
			<span class="text-sm text-text-muted uppercase tracking-wide">Suites</span>
		</div>
		<div class="card flex flex-col gap-1 text-center p-6">
			<span class="text-[32px] font-bold text-accent mono">{overview?.total_runs ?? 0}</span>
			<span class="text-sm text-text-muted uppercase tracking-wide">Runs</span>
		</div>
		<div class="card flex flex-col gap-1 text-center p-6">
			<span class="text-[32px] font-bold text-accent mono">{overview?.completed_runs ?? 0}</span>
			<span class="text-sm text-text-muted uppercase tracking-wide">Completed</span>
		</div>
	</div>

	<div class="mb-10">
		<a href="/suites" class="btn btn-primary">Create New Suite</a>
		<a href="/analytics" class="btn ml-2">Analytics</a>
	</div>

	{#if $suites.length > 0}
		<div class="mt-4">
			<h2 class="text-xl mb-3">Recent Suites</h2>
			<div class="flex flex-col gap-2">
				{#each $suites.slice(0, 5) as suite}
					<a href="/suites/{suite.id}" class="card flex justify-between items-center py-3.5 px-5 text-text hover:border-accent">
						<span class="font-medium">{suite.name}</span>
						<span class="text-sm text-text-muted mono">{suite.model_count} Models &middot; {suite.prompt_count} Prompts</span>
					</a>
				{/each}
			</div>
		</div>
	{/if}
</div>

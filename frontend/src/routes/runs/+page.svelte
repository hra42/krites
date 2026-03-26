<script lang="ts">
	import { onMount } from 'svelte';
	import * as api from '$lib/api/client';
	import type { RunSummary } from '$lib/types';
	import StatusBadge from '$lib/components/StatusBadge.svelte';

	let runs = $state<RunSummary[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			runs = await api.listRuns();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load runs';
		} finally {
			loading = false;
		}
	});

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleString('en-US', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<div>
	<h1 class="text-2xl mb-6">Runs</h1>

	{#if loading}
		<p class="text-text-muted text-center py-10">Loading runs...</p>
	{:else if error}
		<div class="bg-error/10 border border-error rounded-[--radius] px-4 py-3 text-error">{error}</div>
	{:else if runs.length === 0}
		<div class="card text-center py-10">
			<p class="text-text-muted mb-4">No benchmark runs yet.</p>
			<a href="/suites" class="btn btn-primary">Create Suite</a>
		</div>
	{:else}
		<div class="overflow-x-auto">
			<table class="w-full border-collapse">
				<thead>
					<tr>
						<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Suite</th>
						<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Status</th>
						<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Results</th>
						<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border">Created</th>
						<th class="text-left text-sm text-text-dim uppercase tracking-wide px-3 py-2 border-b border-border"></th>
					</tr>
				</thead>
				<tbody>
					{#each runs as run}
						<tr class="hover:bg-bg-elevated">
							<td class="px-3 py-3 border-b border-border text-[15px] font-medium">{run.suite_name}</td>
							<td class="px-3 py-3 border-b border-border"><StatusBadge status={run.status} /></td>
							<td class="px-3 py-3 border-b border-border mono">{run.result_count}</td>
							<td class="px-3 py-3 border-b border-border text-sm text-text-muted mono">{formatDate(run.created_at)}</td>
							<td class="px-3 py-3 border-b border-border">
								<a href="/runs/{run.id}" class="text-[15px] text-accent hover:underline">Details &rarr;</a>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

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

<div class="page">
	<h1>Runs</h1>

	{#if loading}
		<p class="loading">Loading runs...</p>
	{:else if error}
		<div class="error-banner">{error}</div>
	{:else if runs.length === 0}
		<div class="empty card">
			<p>No benchmark runs yet.</p>
			<a href="/suites" class="btn btn-primary">Create Suite</a>
		</div>
	{:else}
		<div class="table-wrapper">
			<table>
				<thead>
					<tr>
						<th>Suite</th>
						<th>Status</th>
						<th>Results</th>
						<th>Created</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each runs as run}
						<tr>
							<td class="suite-name">{run.suite_name}</td>
							<td><StatusBadge status={run.status} /></td>
							<td class="mono">{run.result_count}</td>
							<td class="mono date">{formatDate(run.created_at)}</td>
							<td>
								<a href="/runs/{run.id}" class="detail-link">Details &rarr;</a>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<style>
	h1 {
		font-size: 22px;
		margin-bottom: 24px;
	}

	.table-wrapper {
		overflow-x: auto;
	}

	table {
		width: 100%;
		border-collapse: collapse;
	}

	th {
		text-align: left;
		font-size: 11px;
		color: var(--color-text-dim);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 8px 12px;
		border-bottom: 1px solid var(--color-border);
	}

	td {
		padding: 12px;
		border-bottom: 1px solid var(--color-border);
		font-size: 14px;
	}

	tr:hover {
		background: var(--color-bg-elevated);
	}

	.suite-name {
		font-weight: 500;
	}

	.date {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.detail-link {
		font-size: 13px;
		color: var(--color-accent);
	}

	.detail-link:hover {
		text-decoration: underline;
	}

	.loading {
		color: var(--color-text-muted);
		text-align: center;
		padding: 40px;
	}

	.error-banner {
		background: rgba(248, 113, 113, 0.1);
		border: 1px solid var(--color-error);
		border-radius: var(--radius);
		padding: 12px 16px;
		color: var(--color-error);
	}

	.empty {
		text-align: center;
		padding: 40px;
		color: var(--color-text-muted);
	}

	.empty p {
		margin-bottom: 16px;
	}
</style>

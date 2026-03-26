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

<div class="dashboard">
	<h1>Dashboard</h1>
	<p class="subtitle">LLM Benchmark Platform</p>

	<div class="stats">
		<div class="stat-card card">
			<span class="stat-value mono">{$suites.length}</span>
			<span class="stat-label">Suites</span>
		</div>
		<div class="stat-card card">
			<span class="stat-value mono">{overview?.total_runs ?? 0}</span>
			<span class="stat-label">Runs</span>
		</div>
		<div class="stat-card card">
			<span class="stat-value mono">{overview?.completed_runs ?? 0}</span>
			<span class="stat-label">Completed</span>
		</div>
	</div>

	<div class="actions">
		<a href="/suites" class="btn btn-primary">Create New Suite</a>
		<a href="/analytics" class="btn" style="margin-left: 8px;">Analytics</a>
	</div>

	{#if $suites.length > 0}
		<div class="recent">
			<h2>Recent Suites</h2>
			<div class="recent-list">
				{#each $suites.slice(0, 5) as suite}
					<a href="/suites/{suite.id}" class="recent-item card">
						<span class="recent-name">{suite.name}</span>
						<span class="recent-meta mono">{suite.model_count} Models &middot; {suite.prompt_count} Prompts</span>
					</a>
				{/each}
			</div>
		</div>
	{/if}
</div>

<style>
	.dashboard {
		max-width: 800px;
	}

	h1 {
		font-size: 28px;
		margin-bottom: 4px;
	}

	.subtitle {
		color: var(--color-text-muted);
		margin-bottom: 32px;
	}

	.stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;
		margin-bottom: 32px;
	}

	.stat-card {
		display: flex;
		flex-direction: column;
		gap: 4px;
		text-align: center;
		padding: 24px;
	}

	.stat-value {
		font-size: 32px;
		font-weight: 700;
		color: var(--color-accent);
	}

	.stat-label {
		font-size: 12px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.actions {
		margin-bottom: 40px;
	}

	.recent {
		margin-top: 16px;
	}

	.recent h2 {
		font-size: 18px;
		margin-bottom: 12px;
	}

	.recent-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.recent-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 14px 20px;
		color: var(--color-text);
	}

	.recent-item:hover {
		border-color: var(--color-accent);
	}

	.recent-name {
		font-weight: 500;
	}

	.recent-meta {
		font-size: 12px;
		color: var(--color-text-muted);
	}
</style>

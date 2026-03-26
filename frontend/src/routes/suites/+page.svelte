<script lang="ts">
	import { onMount } from 'svelte';
	import { suites, loading, error, loadSuites, createNewSuite } from '$lib/stores/benchmark';
	import type { Prompt } from '$lib/types';
	import SuiteCard from '$lib/components/SuiteCard.svelte';
	import ModelPicker from '$lib/components/ModelPicker.svelte';
	import PromptEditor from '$lib/components/PromptEditor.svelte';

	let showForm = $state(false);
	let name = $state('');
	let description = $state('');
	let models = $state<string[]>([]);
	let prompts = $state<Omit<Prompt, 'id'>[]>([
		{ name: '', system_message: '', user_message: '', expected_output: '', category: '' }
	]);
	let temperature = $state(0.7);
	let maxTokens = $state(1024);
	let iterations = $state(1);
	let concurrency = $state(3);
	let judgeEnabled = $state(false);
	let judgeModel = $state('');
	let judgeCriteria = $state('');

	onMount(() => {
		loadSuites();
	});

	function resetForm() {
		name = '';
		description = '';
		models = [];
		prompts = [{ name: '', system_message: '', user_message: '', expected_output: '', category: '' }];
		temperature = 0.7;
		maxTokens = 1024;
		iterations = 1;
		concurrency = 3;
		judgeEnabled = false;
		judgeModel = '';
		judgeCriteria = '';
	}

	async function handleSubmit() {
		const criteria = judgeCriteria
			.split(',')
			.map((c) => c.trim())
			.filter(Boolean);

		const suite = await createNewSuite({
			name,
			description,
			prompts,
			models,
			config: {
				temperature,
				max_tokens: maxTokens,
				iterations,
				concurrency,
				judge_enabled: judgeEnabled,
				...(judgeEnabled && judgeModel ? { judge_model: judgeModel } : {}),
				...(judgeEnabled && criteria.length > 0 ? { judge_criteria: criteria } : {})
			}
		});

		if (suite) {
			resetForm();
			showForm = false;
		}
	}
</script>

<div class="page">
	<div class="header">
		<div>
			<h1>Suites</h1>
			<p class="subtitle">Manage benchmark test suites</p>
		</div>
		<button class="btn btn-primary" onclick={() => (showForm = !showForm)}>
			{showForm ? 'Cancel' : '+ New Suite'}
		</button>
	</div>

	{#if $error}
		<div class="error-banner">{$error}</div>
	{/if}

	{#if showForm}
		<form class="form card fade-in" onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
			<h2>Create New Suite</h2>

			<div class="form-section">
				<div class="field">
					<label class="label" for="suite-name">Name *</label>
					<input id="suite-name" class="input" bind:value={name} placeholder="e.g. GPT vs Claude comparison" required />
				</div>

				<div class="field">
					<label class="label" for="suite-desc">Description</label>
					<textarea id="suite-desc" class="input" bind:value={description} placeholder="Brief description of the suite..." rows="2"></textarea>
				</div>
			</div>

			<div class="form-section">
				<h3>Models</h3>
				<ModelPicker selected={models} onchange={(m) => (models = m)} />
			</div>

			<div class="form-section">
				<h3>Prompts</h3>
				<PromptEditor bind:prompts />
			</div>

			<div class="form-section">
				<h3>Configuration</h3>
				<div class="config-grid">
					<div class="field">
						<label class="label" for="cfg-temp">Temperature</label>
						<input id="cfg-temp" type="number" class="input mono" bind:value={temperature} min="0" max="2" step="0.1" />
					</div>
					<div class="field">
						<label class="label" for="cfg-tokens">Max Tokens</label>
						<input id="cfg-tokens" type="number" class="input mono" bind:value={maxTokens} min="1" max="32768" />
					</div>
					<div class="field">
						<label class="label" for="cfg-iter">Iterations</label>
						<input id="cfg-iter" type="number" class="input mono" bind:value={iterations} min="1" max="100" />
					</div>
					<div class="field">
						<label class="label" for="cfg-conc">Parallelism</label>
						<input id="cfg-conc" type="number" class="input mono" bind:value={concurrency} min="1" max="20" />
					</div>
				</div>
			</div>

			<div class="form-section">
				<h3>LLM-as-Judge</h3>
				<label class="checkbox-label">
					<input type="checkbox" bind:checked={judgeEnabled} />
					<span>Enable judge scoring</span>
				</label>
				{#if judgeEnabled}
					<div class="judge-fields fade-in">
						<div class="field">
							<label class="label" for="judge-model">Judge Model</label>
							<input id="judge-model" class="input" bind:value={judgeModel} placeholder="e.g. openai/gpt-4o" />
						</div>
						<div class="field">
							<label class="label" for="judge-criteria">Criteria (comma-separated)</label>
							<input id="judge-criteria" class="input" bind:value={judgeCriteria} placeholder="e.g. accuracy, coherence, helpfulness" />
						</div>
					</div>
				{/if}
			</div>

			<div class="form-actions">
				<button type="button" class="btn" onclick={() => { resetForm(); showForm = false; }}>Cancel</button>
				<button
					type="submit"
					class="btn btn-primary"
					disabled={!name || models.length === 0 || prompts.length === 0 || !prompts.some((p) => p.user_message) || $loading}
				>
					{$loading ? 'Creating...' : 'Create Suite'}
				</button>
			</div>
		</form>
	{/if}

	{#if $loading && !showForm}
		<p class="loading">Loading suites...</p>
	{:else if $suites.length === 0 && !showForm}
		<div class="empty card">
			<p>No suites yet.</p>
			<button class="btn btn-primary" onclick={() => (showForm = true)}>Create First Suite</button>
		</div>
	{:else}
		<div class="suite-grid">
			{#each $suites as suite}
				<SuiteCard {suite} />
			{/each}
		</div>
	{/if}
</div>

<style>
	.page {
		max-width: 900px;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 24px;
	}

	h1 {
		font-size: 24px;
		margin-bottom: 2px;
	}

	.subtitle {
		color: var(--color-text-muted);
		font-size: 14px;
	}

	.error-banner {
		background: rgba(248, 113, 113, 0.1);
		border: 1px solid var(--color-error);
		border-radius: var(--radius);
		padding: 12px 16px;
		color: var(--color-error);
		font-size: 13px;
		margin-bottom: 16px;
	}

	.form {
		margin-bottom: 32px;
		padding: 24px;
	}

	.form h2 {
		font-size: 18px;
		margin-bottom: 20px;
	}

	.form-section {
		margin-bottom: 24px;
		padding-bottom: 24px;
		border-bottom: 1px solid var(--color-border);
	}

	.form-section:last-of-type {
		border-bottom: none;
		margin-bottom: 16px;
		padding-bottom: 0;
	}

	.form-section h3 {
		font-size: 14px;
		color: var(--color-text-muted);
		margin-bottom: 12px;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.field {
		margin-bottom: 12px;
	}

	.config-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 12px;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
		font-size: 14px;
	}

	.checkbox-label input[type='checkbox'] {
		accent-color: var(--color-accent);
		width: 16px;
		height: 16px;
	}

	.judge-fields {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		padding-top: 16px;
		border-top: 1px solid var(--color-border);
	}

	.suite-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 16px;
	}

	.empty {
		text-align: center;
		padding: 40px;
	}

	.empty p {
		color: var(--color-text-muted);
		margin-bottom: 16px;
	}

	.loading {
		color: var(--color-text-muted);
		text-align: center;
		padding: 40px;
	}
</style>

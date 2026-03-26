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

<div>
	<div class="flex justify-between items-start mb-6">
		<div>
			<h1 class="text-3xl mb-0.5">Suites</h1>
			<p class="text-text-muted text-base">Manage benchmark test suites</p>
		</div>
		<button class="btn btn-primary" onclick={() => (showForm = !showForm)}>
			{showForm ? 'Cancel' : '+ New Suite'}
		</button>
	</div>

	{#if $error}
		<div class="bg-error/10 border border-error rounded-[--radius] px-4 py-3 text-error text-[15px] mb-4">{$error}</div>
	{/if}

	{#if showForm}
		<form class="card mb-8 p-6 fade-in" onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
			<h2 class="text-xl mb-5">Create New Suite</h2>

			<div class="mb-6 pb-6 border-b border-border">
				<div class="mb-3">
					<label class="label" for="suite-name">Name *</label>
					<input id="suite-name" class="input" bind:value={name} placeholder="e.g. GPT vs Claude comparison" required />
				</div>
				<div>
					<label class="label" for="suite-desc">Description</label>
					<textarea id="suite-desc" class="input" bind:value={description} placeholder="Brief description of the suite..." rows="2"></textarea>
				</div>
			</div>

			<div class="mb-6 pb-6 border-b border-border">
				<h3 class="text-sm text-text-muted uppercase tracking-wide mb-3">Models</h3>
				<ModelPicker selected={models} onchange={(m) => (models = m)} />
			</div>

			<div class="mb-6 pb-6 border-b border-border">
				<h3 class="text-sm text-text-muted uppercase tracking-wide mb-3">Prompts</h3>
				<PromptEditor bind:prompts />
			</div>

			<div class="mb-6 pb-6 border-b border-border">
				<h3 class="text-sm text-text-muted uppercase tracking-wide mb-3">Configuration</h3>
				<div class="grid grid-cols-2 gap-3">
					<div class="mb-3">
						<label class="label" for="cfg-temp">Temperature</label>
						<input id="cfg-temp" type="number" class="input mono" bind:value={temperature} min="0" max="2" step="0.1" />
					</div>
					<div class="mb-3">
						<label class="label" for="cfg-tokens">Max Tokens</label>
						<input id="cfg-tokens" type="number" class="input mono" bind:value={maxTokens} min="1" max="32768" />
					</div>
					<div class="mb-3">
						<label class="label" for="cfg-iter">Iterations</label>
						<input id="cfg-iter" type="number" class="input mono" bind:value={iterations} min="1" max="100" />
					</div>
					<div class="mb-3">
						<label class="label" for="cfg-conc">Parallelism</label>
						<input id="cfg-conc" type="number" class="input mono" bind:value={concurrency} min="1" max="20" />
					</div>
				</div>
			</div>

			<div class="mb-4">
				<h3 class="text-sm text-text-muted uppercase tracking-wide mb-3">LLM-as-Judge</h3>
				<label class="flex items-center gap-2 cursor-pointer text-sm">
					<input type="checkbox" class="accent-accent w-4 h-4" bind:checked={judgeEnabled} />
					<span>Enable judge scoring</span>
				</label>
				{#if judgeEnabled}
					<div class="mt-3 flex flex-col fade-in">
						<div class="mb-3">
							<label class="label" for="judge-model">Judge Model</label>
							<input id="judge-model" class="input" bind:value={judgeModel} placeholder="e.g. openai/gpt-4o" />
						</div>
						<div>
							<label class="label" for="judge-criteria">Criteria (comma-separated)</label>
							<input id="judge-criteria" class="input" bind:value={judgeCriteria} placeholder="e.g. accuracy, coherence, helpfulness" />
						</div>
					</div>
				{/if}
			</div>

			<div class="flex justify-end gap-2 pt-4 border-t border-border">
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
		<p class="text-text-muted text-center py-10">Loading suites...</p>
	{:else if $suites.length === 0 && !showForm}
		<div class="card text-center py-10">
			<p class="text-text-muted mb-4">No suites yet.</p>
			<button class="btn btn-primary" onclick={() => (showForm = true)}>Create First Suite</button>
		</div>
	{:else}
		<div class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-4">
			{#each $suites as suite}
				<SuiteCard {suite} />
			{/each}
		</div>
	{/if}
</div>

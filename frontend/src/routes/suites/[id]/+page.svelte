<script lang="ts">
	import { onMount } from 'svelte';
	import { currentSuite, loading, error, loadSuite, removeSuite, editSuite } from '$lib/stores/benchmark';
	import ModelChip from '$lib/components/ModelChip.svelte';
	import ModelPicker from '$lib/components/ModelPicker.svelte';
	import PromptEditor from '$lib/components/PromptEditor.svelte';
	import { goto } from '$app/navigation';
	import type { Prompt } from '$lib/types';

	let { data } = $props();

	let editing = $state(false);

	let editName = $state('');
	let editDescription = $state('');
	let editModels = $state<string[]>([]);
	let editPrompts = $state<Omit<Prompt, 'id'>[]>([]);
	let editTemperature = $state(0.7);
	let editMaxTokens = $state(1024);
	let editIterations = $state(1);
	let editConcurrency = $state(3);
	let editJudgeEnabled = $state(false);
	let editJudgeModel = $state('');
	let editJudgeCriteria = $state('');

	onMount(() => {
		loadSuite(data.id);
	});

	function startEditing() {
		if (!$currentSuite) return;
		editName = $currentSuite.name;
		editDescription = $currentSuite.description;
		editModels = [...$currentSuite.models];
		editPrompts = $currentSuite.prompts.map((p) => ({
			name: p.name,
			system_message: p.system_message,
			user_message: p.user_message,
			expected_output: p.expected_output ?? '',
			category: p.category ?? ''
		}));
		editTemperature = $currentSuite.config.temperature;
		editMaxTokens = $currentSuite.config.max_tokens;
		editIterations = $currentSuite.config.iterations;
		editConcurrency = $currentSuite.config.concurrency;
		editJudgeEnabled = $currentSuite.config.judge_enabled;
		editJudgeModel = $currentSuite.config.judge_model ?? '';
		editJudgeCriteria = $currentSuite.config.judge_criteria?.join(', ') ?? '';
		editing = true;
	}

	function cancelEditing() {
		editing = false;
	}

	async function handleSave() {
		if (!$currentSuite) return;
		const criteria = editJudgeCriteria
			.split(',')
			.map((c) => c.trim())
			.filter(Boolean);

		const result = await editSuite($currentSuite.id, {
			name: editName,
			description: editDescription,
			prompts: editPrompts as Prompt[],
			models: editModels,
			config: {
				temperature: editTemperature,
				max_tokens: editMaxTokens,
				iterations: editIterations,
				concurrency: editConcurrency,
				judge_enabled: editJudgeEnabled,
				...(editJudgeEnabled && editJudgeModel ? { judge_model: editJudgeModel } : {}),
				...(editJudgeEnabled && criteria.length > 0 ? { judge_criteria: criteria } : {})
			}
		});
		if (result) {
			editing = false;
		}
	}

	let showDeleteConfirm = $state(false);

	async function handleDelete() {
		if ($currentSuite) {
			await removeSuite($currentSuite.id);
			goto('/suites');
		}
	}
</script>

{#if $loading}
	<p class="text-text-muted text-center py-10">Loading suite...</p>
{:else if $error}
	<div class="bg-error/10 border border-error rounded-[--radius] px-4 py-3 text-error">{$error}</div>
{:else if $currentSuite}
	<div>
		<div class="flex justify-between items-start mb-8">
			<div>
				<a href="/suites" class="text-[15px] text-text-muted hover:text-accent mb-2 inline-block">&larr; Back</a>
				{#if editing}
					<h1 class="text-3xl mb-1">Edit Suite</h1>
				{:else}
					<h1 class="text-3xl mb-1">{$currentSuite.name}</h1>
					{#if $currentSuite.description}
						<p class="text-text-muted">{$currentSuite.description}</p>
					{/if}
				{/if}
			</div>
			<div class="flex gap-2 shrink-0">
				{#if editing}
					<button class="btn" onclick={cancelEditing}>Cancel</button>
					<button
						class="btn btn-primary"
						onclick={handleSave}
						disabled={!editName || editModels.length === 0 || editPrompts.length === 0 || !editPrompts.some((p) => p.user_message) || $loading}
					>
						{$loading ? 'Saving...' : 'Save'}
					</button>
				{:else}
					<button class="btn" onclick={startEditing}>Edit</button>
					<button class="btn btn-primary" onclick={() => goto(`/suites/${$currentSuite.id}/run`)}>Start Benchmark</button>
					<button class="btn btn-danger" onclick={() => (showDeleteConfirm = true)}>Delete</button>
				{/if}
			</div>
		</div>

		{#if editing}
			<!-- Edit Form -->
			<form class="flex flex-col fade-in" onsubmit={(e) => { e.preventDefault(); handleSave(); }}>
				<div class="mb-6 pb-6 border-b border-border">
					<div class="mb-3">
						<label class="label" for="edit-name">Name *</label>
						<input id="edit-name" class="input" bind:value={editName} required />
					</div>
					<div>
						<label class="label" for="edit-desc">Description</label>
						<textarea id="edit-desc" class="input" bind:value={editDescription} rows="2"></textarea>
					</div>
				</div>

				<div class="mb-6 pb-6 border-b border-border">
					<h3 class="text-sm text-text-muted uppercase tracking-wide mb-3">Models</h3>
					<ModelPicker selected={editModels} onchange={(m) => (editModels = m)} />
				</div>

				<div class="mb-6 pb-6 border-b border-border">
					<h3 class="text-sm text-text-muted uppercase tracking-wide mb-3">Prompts</h3>
					<PromptEditor bind:prompts={editPrompts} />
				</div>

				<div class="mb-6 pb-6 border-b border-border">
					<h3 class="text-sm text-text-muted uppercase tracking-wide mb-3">Configuration</h3>
					<div class="grid grid-cols-2 gap-3">
						<div class="mb-3">
							<label class="label" for="edit-temp">Temperature</label>
							<input id="edit-temp" type="number" class="input mono" bind:value={editTemperature} min="0" max="2" step="0.1" />
						</div>
						<div class="mb-3">
							<label class="label" for="edit-tokens">Max Tokens</label>
							<input id="edit-tokens" type="number" class="input mono" bind:value={editMaxTokens} min="1" max="32768" />
						</div>
						<div class="mb-3">
							<label class="label" for="edit-iter">Iterations</label>
							<input id="edit-iter" type="number" class="input mono" bind:value={editIterations} min="1" max="100" />
						</div>
						<div class="mb-3">
							<label class="label" for="edit-conc">Parallelism</label>
							<input id="edit-conc" type="number" class="input mono" bind:value={editConcurrency} min="1" max="20" />
						</div>
					</div>
				</div>

				<div>
					<h3 class="text-sm text-text-muted uppercase tracking-wide mb-3">LLM-as-Judge</h3>
					<label class="flex items-center gap-2 cursor-pointer text-sm">
						<input type="checkbox" class="accent-accent w-4 h-4" bind:checked={editJudgeEnabled} />
						<span>Enable judge scoring</span>
					</label>
					{#if editJudgeEnabled}
						<div class="mt-3 flex flex-col fade-in">
							<div class="mb-3">
								<label class="label" for="edit-judge-model">Judge Model</label>
								<input id="edit-judge-model" class="input" bind:value={editJudgeModel} placeholder="e.g. openai/gpt-4o" />
							</div>
							<div>
								<label class="label" for="edit-judge-criteria">Criteria (comma-separated)</label>
								<input id="edit-judge-criteria" class="input" bind:value={editJudgeCriteria} placeholder="e.g. accuracy, coherence, helpfulness" />
							</div>
						</div>
					{/if}
				</div>
			</form>
		{:else}
			<!-- Read-only View -->
			<section class="mb-8">
				<h2 class="text-lg text-text-muted uppercase tracking-wide mb-3">Models</h2>
				<div class="flex flex-wrap gap-1.5">
					{#each $currentSuite.models as model}
						<ModelChip {model} />
					{/each}
				</div>
			</section>

			<section class="mb-8">
				<h2 class="text-lg text-text-muted uppercase tracking-wide mb-3">Prompts</h2>
				<div class="flex flex-col gap-3">
					{#each $currentSuite.prompts as prompt, i}
						<div class="card p-4">
							<div class="flex items-center gap-2.5 mb-3">
								<span class="text-accent font-semibold text-sm mono">#{i + 1}</span>
								<span class="font-medium">{prompt.name || 'Unnamed'}</span>
								{#if prompt.category}
									<span class="text-sm px-2 py-0.5 bg-bg-elevated rounded-[20px] text-text-muted">{prompt.category}</span>
								{/if}
							</div>

							{#if prompt.system_message}
								<div class="mb-2">
									<span class="block text-sm font-semibold uppercase tracking-wide mb-1 text-text-dim">System</span>
									<pre class="bg-bg border border-border rounded-[--radius-sm] px-3 py-2.5 text-sm whitespace-pre-wrap break-words text-text mono">{prompt.system_message}</pre>
								</div>
							{/if}

							<div class="mb-2">
								<span class="block text-sm font-semibold uppercase tracking-wide mb-1 text-accent">User</span>
								<pre class="bg-bg border border-border rounded-[--radius-sm] px-3 py-2.5 text-sm whitespace-pre-wrap break-words text-text mono">{prompt.user_message}</pre>
							</div>

							{#if prompt.expected_output}
								<div>
									<span class="block text-sm font-semibold uppercase tracking-wide mb-1 text-success">Expected Response</span>
									<pre class="bg-bg border border-border rounded-[--radius-sm] px-3 py-2.5 text-sm whitespace-pre-wrap break-words text-text mono">{prompt.expected_output}</pre>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			</section>

			<section class="mb-8">
				<h2 class="text-lg text-text-muted uppercase tracking-wide mb-3">Configuration</h2>
				<div class="grid grid-cols-3 gap-3">
					<div class="bg-bg-card border border-border rounded-[--radius] px-4 py-3">
						<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Temperature</span>
						<span class="text-xl font-semibold mono">{$currentSuite.config.temperature}</span>
					</div>
					<div class="bg-bg-card border border-border rounded-[--radius] px-4 py-3">
						<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Max Tokens</span>
						<span class="text-xl font-semibold mono">{$currentSuite.config.max_tokens}</span>
					</div>
					<div class="bg-bg-card border border-border rounded-[--radius] px-4 py-3">
						<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Iterations</span>
						<span class="text-xl font-semibold mono">{$currentSuite.config.iterations}</span>
					</div>
					<div class="bg-bg-card border border-border rounded-[--radius] px-4 py-3">
						<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Parallelism</span>
						<span class="text-xl font-semibold mono">{$currentSuite.config.concurrency}</span>
					</div>
					<div class="bg-bg-card border border-border rounded-[--radius] px-4 py-3">
						<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Timeout</span>
						<span class="text-xl font-semibold mono">{$currentSuite.config.timeout_seconds}s</span>
					</div>
					<div class="bg-bg-card border border-border rounded-[--radius] px-4 py-3">
						<span class="block text-sm text-text-muted uppercase tracking-wide mb-1">Judge</span>
						<span class="text-xl font-semibold mono">{$currentSuite.config.judge_enabled ? 'Active' : 'Disabled'}</span>
					</div>
				</div>
				{#if $currentSuite.config.judge_enabled && $currentSuite.config.judge_model}
					<div class="mt-3 px-4 py-3 bg-bg-card border border-border rounded-[--radius] text-[15px]">
						<p class="mb-1"><strong>Judge Model:</strong> <span class="mono">{$currentSuite.config.judge_model}</span></p>
						{#if $currentSuite.config.judge_criteria?.length}
							<p><strong>Criteria:</strong> {$currentSuite.config.judge_criteria.join(', ')}</p>
						{/if}
					</div>
				{/if}
			</section>
		{/if}
	</div>

	<!-- Delete Confirmation Modal -->
	{#if showDeleteConfirm}
		<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_noninteractive_element_interactions -->
		<div class="fixed inset-0 bg-black/60 flex items-center justify-center z-[9000]" onclick={() => (showDeleteConfirm = false)} role="dialog">
			<div class="bg-bg-card border border-border rounded-[--radius-lg] p-6 max-w-[420px] w-[90%] fade-in" onclick={(e) => e.stopPropagation()} role="document">
				<h3 class="text-lg mb-2">Delete Suite?</h3>
				<p class="text-text-muted text-base mb-5 leading-normal">Are you sure you want to delete <strong>{$currentSuite.name}</strong>? This action cannot be undone.</p>
				<div class="flex justify-end gap-2">
					<button class="btn" onclick={() => (showDeleteConfirm = false)}>Cancel</button>
					<button class="btn btn-danger" onclick={handleDelete}>Delete</button>
				</div>
			</div>
		</div>
	{/if}
{/if}

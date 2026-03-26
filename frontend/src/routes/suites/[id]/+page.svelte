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

	// Edit form state
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
	<p class="loading">Loading suite...</p>
{:else if $error}
	<div class="error-banner">{$error}</div>
{:else if $currentSuite}
	<div class="page">
		<div class="header">
			<div>
				<a href="/suites" class="back">&larr; Back</a>
				{#if editing}
					<h1>Edit Suite</h1>
				{:else}
					<h1>{$currentSuite.name}</h1>
					{#if $currentSuite.description}
						<p class="description">{$currentSuite.description}</p>
					{/if}
				{/if}
			</div>
			<div class="actions">
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
			<form class="edit-form fade-in" onsubmit={(e) => { e.preventDefault(); handleSave(); }}>
				<div class="form-section">
					<div class="field">
						<label class="label" for="edit-name">Name *</label>
						<input id="edit-name" class="input" bind:value={editName} required />
					</div>
					<div class="field">
						<label class="label" for="edit-desc">Description</label>
						<textarea id="edit-desc" class="input" bind:value={editDescription} rows="2"></textarea>
					</div>
				</div>

				<div class="form-section">
					<h3>Models</h3>
					<ModelPicker selected={editModels} onchange={(m) => (editModels = m)} />
				</div>

				<div class="form-section">
					<h3>Prompts</h3>
					<PromptEditor bind:prompts={editPrompts} />
				</div>

				<div class="form-section">
					<h3>Configuration</h3>
					<div class="config-edit-grid">
						<div class="field">
							<label class="label" for="edit-temp">Temperature</label>
							<input id="edit-temp" type="number" class="input mono" bind:value={editTemperature} min="0" max="2" step="0.1" />
						</div>
						<div class="field">
							<label class="label" for="edit-tokens">Max Tokens</label>
							<input id="edit-tokens" type="number" class="input mono" bind:value={editMaxTokens} min="1" max="32768" />
						</div>
						<div class="field">
							<label class="label" for="edit-iter">Iterations</label>
							<input id="edit-iter" type="number" class="input mono" bind:value={editIterations} min="1" max="100" />
						</div>
						<div class="field">
							<label class="label" for="edit-conc">Parallelism</label>
							<input id="edit-conc" type="number" class="input mono" bind:value={editConcurrency} min="1" max="20" />
						</div>
					</div>
				</div>

				<div class="form-section">
					<h3>LLM-as-Judge</h3>
					<label class="checkbox-label">
						<input type="checkbox" bind:checked={editJudgeEnabled} />
						<span>Enable judge scoring</span>
					</label>
					{#if editJudgeEnabled}
						<div class="judge-fields fade-in">
							<div class="field">
								<label class="label" for="edit-judge-model">Judge Model</label>
								<input id="edit-judge-model" class="input" bind:value={editJudgeModel} placeholder="e.g. openai/gpt-4o" />
							</div>
							<div class="field">
								<label class="label" for="edit-judge-criteria">Criteria (comma-separated)</label>
								<input id="edit-judge-criteria" class="input" bind:value={editJudgeCriteria} placeholder="e.g. accuracy, coherence, helpfulness" />
							</div>
						</div>
					{/if}
				</div>
			</form>
		{:else}
			<!-- Read-only View -->
			<section class="section">
				<h2>Models</h2>
				<div class="chips">
					{#each $currentSuite.models as model}
						<ModelChip {model} />
					{/each}
				</div>
			</section>

			<section class="section">
				<h2>Prompts</h2>
				<div class="prompts">
					{#each $currentSuite.prompts as prompt, i}
						<div class="prompt-block card">
							<div class="prompt-header">
								<span class="prompt-number mono">#{i + 1}</span>
								<span class="prompt-name">{prompt.name || 'Unnamed'}</span>
								{#if prompt.category}
									<span class="prompt-category">{prompt.category}</span>
								{/if}
							</div>

							{#if prompt.system_message}
								<div class="message-block system">
									<span class="message-label">System</span>
									<pre class="message-content mono">{prompt.system_message}</pre>
								</div>
							{/if}

							<div class="message-block user">
								<span class="message-label">User</span>
								<pre class="message-content mono">{prompt.user_message}</pre>
							</div>

							{#if prompt.expected_output}
								<div class="message-block expected">
									<span class="message-label">Expected Response</span>
									<pre class="message-content mono">{prompt.expected_output}</pre>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			</section>

			<section class="section">
				<h2>Configuration</h2>
				<div class="config-grid">
					<div class="config-item">
						<span class="config-label">Temperature</span>
						<span class="config-value mono">{$currentSuite.config.temperature}</span>
					</div>
					<div class="config-item">
						<span class="config-label">Max Tokens</span>
						<span class="config-value mono">{$currentSuite.config.max_tokens}</span>
					</div>
					<div class="config-item">
						<span class="config-label">Iterations</span>
						<span class="config-value mono">{$currentSuite.config.iterations}</span>
					</div>
					<div class="config-item">
						<span class="config-label">Parallelism</span>
						<span class="config-value mono">{$currentSuite.config.concurrency}</span>
					</div>
					<div class="config-item">
						<span class="config-label">Timeout</span>
						<span class="config-value mono">{$currentSuite.config.timeout_seconds}s</span>
					</div>
					<div class="config-item">
						<span class="config-label">Judge</span>
						<span class="config-value mono">{$currentSuite.config.judge_enabled ? 'Active' : 'Disabled'}</span>
					</div>
				</div>
				{#if $currentSuite.config.judge_enabled && $currentSuite.config.judge_model}
					<div class="judge-info">
						<p><strong>Judge Model:</strong> <span class="mono">{$currentSuite.config.judge_model}</span></p>
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
		<div class="modal-backdrop" onclick={() => (showDeleteConfirm = false)} role="dialog">
			<div class="modal fade-in" onclick={(e) => e.stopPropagation()} role="document">
				<h3>Delete Suite?</h3>
				<p>Are you sure you want to delete <strong>{$currentSuite.name}</strong>? This action cannot be undone.</p>
				<div class="modal-actions">
					<button class="btn" onclick={() => (showDeleteConfirm = false)}>Cancel</button>
					<button class="btn btn-danger" onclick={handleDelete}>Delete</button>
				</div>
			</div>
		</div>
	{/if}
{/if}

<style>
	.header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 32px;
	}

	.back {
		font-size: 13px;
		color: var(--color-text-muted);
		margin-bottom: 8px;
		display: inline-block;
	}

	.back:hover {
		color: var(--color-accent);
	}

	h1 {
		font-size: 24px;
		margin-bottom: 4px;
	}

	.description {
		color: var(--color-text-muted);
	}

	.actions {
		display: flex;
		gap: 8px;
		flex-shrink: 0;
	}

	/* Read-only view styles */
	.section {
		margin-bottom: 32px;
	}

	.section h2 {
		font-size: 16px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 12px;
	}

	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.prompts {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.prompt-block {
		padding: 16px;
	}

	.prompt-header {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-bottom: 12px;
	}

	.prompt-number {
		color: var(--color-accent);
		font-weight: 600;
		font-size: 13px;
	}

	.prompt-name {
		font-weight: 500;
	}

	.prompt-category {
		font-size: 11px;
		padding: 2px 8px;
		background: var(--color-bg-elevated);
		border-radius: 20px;
		color: var(--color-text-muted);
	}

	.message-block {
		margin-bottom: 8px;
	}

	.message-block:last-child {
		margin-bottom: 0;
	}

	.message-label {
		display: block;
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 4px;
	}

	.system .message-label {
		color: var(--color-text-dim);
	}

	.user .message-label {
		color: var(--color-accent);
	}

	.expected .message-label {
		color: var(--color-success);
	}

	.message-content {
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		padding: 10px 12px;
		font-size: 12px;
		white-space: pre-wrap;
		word-break: break-word;
		color: var(--color-text);
	}

	.config-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
	}

	.config-item {
		background: var(--color-bg-card);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		padding: 12px 16px;
	}

	.config-label {
		display: block;
		font-size: 11px;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 4px;
	}

	.config-value {
		font-size: 18px;
		font-weight: 600;
	}

	.judge-info {
		margin-top: 12px;
		padding: 12px 16px;
		background: var(--color-bg-card);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		font-size: 13px;
	}

	.judge-info p {
		margin-bottom: 4px;
	}

	/* Edit form styles */
	.edit-form {
		display: flex;
		flex-direction: column;
	}

	.form-section {
		margin-bottom: 24px;
		padding-bottom: 24px;
		border-bottom: 1px solid var(--color-border);
	}

	.form-section:last-of-type {
		border-bottom: none;
		margin-bottom: 0;
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

	.config-edit-grid {
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

	/* Delete confirmation modal */
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 9000;
	}

	.modal {
		background: var(--color-bg-card);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: 24px;
		max-width: 420px;
		width: 90%;
	}

	.modal h3 {
		font-size: 18px;
		margin-bottom: 8px;
	}

	.modal p {
		color: var(--color-text-muted);
		font-size: 14px;
		margin-bottom: 20px;
		line-height: 1.5;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}

	/* Shared */
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
</style>

<script lang="ts">
	import type { Prompt } from '$lib/types';

	interface Props {
		prompts: (Omit<Prompt, 'id'> & { id?: string })[];
	}

	let { prompts = $bindable() }: Props = $props();

	function addPrompt() {
		prompts = [
			...prompts,
			{ name: '', system_message: '', user_message: '', expected_output: '', category: '' }
		];
	}

	function removePrompt(index: number) {
		prompts = prompts.filter((_, i) => i !== index);
	}
</script>

<div class="prompts">
	{#each prompts as prompt, i}
		<div class="prompt-card fade-in">
			<div class="prompt-header">
				<span class="prompt-label">Prompt {i + 1}</span>
				{#if prompts.length > 1}
					<button class="btn-remove" onclick={() => removePrompt(i)}>&times;</button>
				{/if}
			</div>

			<div class="field">
				<label class="label" for="prompt-name-{i}">Name</label>
				<input id="prompt-name-{i}" class="input" bind:value={prompt.name} placeholder="e.g. Summarization" />
			</div>

			<div class="field">
				<label class="label" for="prompt-system-{i}">System Message</label>
				<textarea id="prompt-system-{i}" class="input" bind:value={prompt.system_message} placeholder="Optional system message..." rows="2"></textarea>
			</div>

			<div class="field">
				<label class="label" for="prompt-user-{i}">User Message *</label>
				<textarea id="prompt-user-{i}" class="input" bind:value={prompt.user_message} placeholder="The message to the model..." rows="3"></textarea>
			</div>

			<div class="field">
				<label class="label" for="prompt-expected-{i}">Expected Response</label>
				<textarea id="prompt-expected-{i}" class="input" bind:value={prompt.expected_output} placeholder="Optional: Expected output for judge scoring..." rows="2"></textarea>
			</div>

			<div class="field">
				<label class="label" for="prompt-category-{i}">Category</label>
				<input id="prompt-category-{i}" class="input" bind:value={prompt.category} placeholder="e.g. reasoning, coding, creative" />
			</div>
		</div>
	{/each}

	<button class="btn add-btn" onclick={addPrompt}>+ Add Prompt</button>
</div>

<style>
	.prompts {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.prompt-card {
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		padding: 16px;
	}

	.prompt-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}

	.prompt-label {
		font-size: 13px;
		font-weight: 600;
		color: var(--color-accent);
	}

	.btn-remove {
		width: 24px;
		height: 24px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--color-text-muted);
		font-size: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.15s;
	}

	.btn-remove:hover {
		border-color: var(--color-error);
		color: var(--color-error);
	}

	.field {
		margin-bottom: 12px;
	}

	.field:last-child {
		margin-bottom: 0;
	}

	.add-btn {
		align-self: flex-start;
	}
</style>

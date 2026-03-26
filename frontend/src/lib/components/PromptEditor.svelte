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

<div class="flex flex-col gap-4">
	{#each prompts as prompt, i}
		<div class="bg-bg-elevated border border-border rounded-[--radius] p-4 fade-in">
			<div class="flex justify-between items-center mb-3">
				<span class="text-base font-semibold text-accent">Prompt {i + 1}</span>
				{#if prompts.length > 1}
					<button
						class="w-6 h-6 border border-border rounded-[--radius-sm] bg-transparent text-text-muted text-base flex items-center justify-center transition-all duration-150 hover:border-error hover:text-error"
						onclick={() => removePrompt(i)}>&times;</button>
				{/if}
			</div>

			<div class="mb-3">
				<label class="label" for="prompt-name-{i}">Name</label>
				<input id="prompt-name-{i}" class="input" bind:value={prompt.name} placeholder="e.g. Summarization" />
			</div>

			<div class="mb-3">
				<label class="label" for="prompt-system-{i}">System Message</label>
				<textarea id="prompt-system-{i}" class="input" bind:value={prompt.system_message} placeholder="Optional system message..." rows="2"></textarea>
			</div>

			<div class="mb-3">
				<label class="label" for="prompt-user-{i}">User Message *</label>
				<textarea id="prompt-user-{i}" class="input" bind:value={prompt.user_message} placeholder="The message to the model..." rows="3"></textarea>
			</div>

			<div class="mb-3">
				<label class="label" for="prompt-expected-{i}">Expected Response</label>
				<textarea id="prompt-expected-{i}" class="input" bind:value={prompt.expected_output} placeholder="Optional: Expected output for judge scoring..." rows="2"></textarea>
			</div>

			<div>
				<label class="label" for="prompt-category-{i}">Category</label>
				<input id="prompt-category-{i}" class="input" bind:value={prompt.category} placeholder="e.g. reasoning, coding, creative" />
			</div>
		</div>
	{/each}

	<button class="btn self-start" onclick={addPrompt}>+ Add Prompt</button>
</div>

<script lang="ts">
	interface Props {
		open: boolean;
		suiteName: string;
		busy?: boolean;
		onclose: () => void;
		onconfirm: (name: string) => void;
	}

	let { open, suiteName, busy = false, onclose, onconfirm }: Props = $props();

	let name = $state('');
	let lastSuiteName = $state('');

	$effect(() => {
		if (open && suiteName !== lastSuiteName) {
			name = `${suiteName} (Copy)`;
			lastSuiteName = suiteName;
		}
		if (!open) {
			lastSuiteName = '';
		}
	});

	function submit() {
		const trimmed = name.trim();
		if (!trimmed || busy) return;
		onconfirm(trimmed);
	}
</script>

{#if open}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_noninteractive_element_interactions a11y_interactive_supports_focus a11y_autofocus -->
	<div class="fixed inset-0 bg-black/60 flex items-center justify-center z-[9000]" onclick={onclose} role="dialog">
		<div class="bg-bg-card border border-border rounded-[--radius-lg] p-6 max-w-[420px] w-[90%] fade-in" onclick={(e) => e.stopPropagation()} role="document">
			<h3 class="text-lg mb-2">Duplicate Suite</h3>
			<p class="text-text-muted text-base mb-4 leading-normal">Create a copy of <strong>{suiteName}</strong> with a new name.</p>
			<form onsubmit={(e) => { e.preventDefault(); submit(); }}>
				<div class="mb-5">
					<label class="label" for="duplicate-name">Name *</label>
					<input
						id="duplicate-name"
						class="input"
						bind:value={name}
						required
						autofocus
					/>
				</div>
				<div class="flex justify-end gap-2">
					<button type="button" class="btn" onclick={onclose} disabled={busy}>Cancel</button>
					<button type="submit" class="btn btn-primary" disabled={!name.trim() || busy}>
						{busy ? 'Duplicating...' : 'Duplicate'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

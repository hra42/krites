<script lang="ts">
	interface Props {
		model: string;
		removable?: boolean;
		onremove?: () => void;
	}

	let { model, removable = false, onremove }: Props = $props();
</script>

<span class="chip mono">
	{model}
	{#if removable}
		<button class="remove" onclick={onremove}>&times;</button>
	{/if}
</span>

<style>
	.chip {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 3px 10px;
		background: var(--color-accent-bg);
		border: 1px solid rgba(167, 139, 250, 0.2);
		border-radius: 20px;
		color: var(--color-accent);
		font-size: 12px;
		white-space: nowrap;
	}

	.remove {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 16px;
		height: 16px;
		padding: 0;
		border: none;
		border-radius: 50%;
		background: transparent;
		color: var(--color-accent);
		font-size: 14px;
		cursor: pointer;
		flex-shrink: 0;
		transition: background 0.15s;
	}

	.remove:hover {
		background: rgba(167, 139, 250, 0.2);
	}
</style>

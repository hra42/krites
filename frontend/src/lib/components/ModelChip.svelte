<script lang="ts">
	interface Props {
		model: string;
		removable?: boolean;
		onremove?: () => void;
	}

	let { model, removable = false, onremove }: Props = $props();
</script>

<span class="inline-flex items-center gap-1.5 px-3 py-1 bg-accent-bg border border-accent/20 rounded-[20px] text-accent text-sm whitespace-nowrap mono">
	{model}
	{#if removable}
		<button
			class="inline-flex items-center justify-center w-5 h-5 p-0 border-none rounded-full bg-transparent text-accent text-base shrink-0 transition-[background] duration-150 hover:bg-accent/20"
			onclick={onremove}>&times;</button>
	{/if}
</span>

<script lang="ts">
	interface Props {
		completed: number;
		total: number;
		percent: number;
	}

	let { completed, total, percent }: Props = $props();
</script>

<div class="flex items-center gap-3">
	<div class="flex-1 h-2 bg-bg-input rounded-[--radius-sm] overflow-hidden">
		<div class="h-full bg-accent rounded-[--radius-sm] transition-[width] duration-300 ease-in-out" style="width: {percent}%"></div>
	</div>
	<span class="text-text-muted whitespace-nowrap mono">{completed}/{total} ({percent}%)</span>
</div>

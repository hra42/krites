<script lang="ts">
	interface Props {
		criterion: string;
		score: number;
	}

	let { criterion, score }: Props = $props();

	const variantClasses = $derived(
		score >= 8
			? 'border-success bg-success/10'
			: score >= 5
				? 'border-warning bg-warning/10'
				: 'border-error bg-error/10'
	);

	const scoreColor = $derived(
		score >= 8 ? 'text-success' : score >= 5 ? 'text-warning' : 'text-error'
	);
</script>

<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-[--radius-lg] text-sm border {variantClasses}">
	<span class="capitalize text-text-muted">{criterion}</span>
	<span class="font-semibold mono {scoreColor}">{score.toFixed(1)}</span>
</span>

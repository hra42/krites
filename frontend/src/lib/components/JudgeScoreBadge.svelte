<script lang="ts">
	interface Props {
		criterion: string;
		score: number;
	}

	let { criterion, score }: Props = $props();

	const colorClass = $derived(
		score >= 8 ? 'score-high' : score >= 5 ? 'score-mid' : 'score-low'
	);
</script>

<span class="badge {colorClass}">
	<span class="criterion">{criterion}</span>
	<span class="score mono">{score.toFixed(1)}</span>
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 2px 8px;
		border-radius: 12px;
		font-size: 11px;
		border: 1px solid;
	}

	.criterion {
		text-transform: capitalize;
		color: var(--color-text-muted);
	}

	.score {
		font-weight: 600;
	}

	.score-high {
		border-color: var(--color-success);
		background: rgba(52, 211, 153, 0.1);
	}

	.score-high .score {
		color: var(--color-success);
	}

	.score-mid {
		border-color: var(--color-warning);
		background: rgba(251, 191, 36, 0.1);
	}

	.score-mid .score {
		color: var(--color-warning);
	}

	.score-low {
		border-color: var(--color-error);
		background: rgba(248, 113, 113, 0.1);
	}

	.score-low .score {
		color: var(--color-error);
	}
</style>

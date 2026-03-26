<script lang="ts">
	interface Props {
		completed: number;
		total: number;
		percent: number;
	}

	let { completed, total, percent }: Props = $props();
</script>

<div class="progress-container">
	<div class="progress-bar">
		<div class="progress-fill" style="width: {percent}%"></div>
	</div>
	<span class="progress-label mono">{completed}/{total} ({percent}%)</span>
</div>

<style>
	.progress-container {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.progress-bar {
		flex: 1;
		height: 8px;
		background: var(--color-bg-input);
		border-radius: 4px;
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: var(--color-accent);
		border-radius: 4px;
		transition: width 0.3s ease;
	}

	.progress-label {
		font-size: 13px;
		color: var(--color-text-muted);
		white-space: nowrap;
	}
</style>

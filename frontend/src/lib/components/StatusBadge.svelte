<script lang="ts">
	import type { RunStatus } from '$lib/types';

	interface Props {
		status: RunStatus;
	}

	let { status }: Props = $props();

	const labels: Record<RunStatus, string> = {
		pending: 'Pending',
		running: 'Running',
		complete: 'Completed',
		failed: 'Failed',
		canceled: 'Canceled'
	};
</script>

<span class="badge {status}" class:pulse={status === 'running'}>
	{labels[status]}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		padding: 2px 10px;
		border-radius: 20px;
		font-size: 12px;
		font-weight: 500;
	}

	.pending {
		background: rgba(139, 136, 148, 0.15);
		color: var(--color-text-muted);
	}

	.running {
		background: rgba(251, 191, 36, 0.15);
		color: var(--color-warning);
	}

	.complete {
		background: rgba(52, 211, 153, 0.15);
		color: var(--color-success);
	}

	.failed {
		background: rgba(248, 113, 113, 0.15);
		color: var(--color-error);
	}

	.canceled {
		background: rgba(139, 136, 148, 0.15);
		color: var(--color-text-muted);
	}
</style>

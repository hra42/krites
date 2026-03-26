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

	const variantClasses: Record<RunStatus, string> = {
		pending: 'bg-text-muted/15 text-text-muted',
		running: 'bg-warning/15 text-warning',
		complete: 'bg-success/15 text-success',
		failed: 'bg-error/15 text-error',
		canceled: 'bg-text-muted/15 text-text-muted'
	};
</script>

<span
	class="inline-flex items-center px-3 py-1 rounded-[20px] text-sm font-medium {variantClasses[status]}"
	class:pulse={status === 'running'}
>
	{labels[status]}
</span>

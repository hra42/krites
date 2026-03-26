<script lang="ts">
	import { LineChart } from 'layerchart';
	import type { Result } from '$lib/types';
	import { getModelColor } from '$lib/utils/chart-colors';

	interface Props {
		results: Result[];
		models: string[];
	}

	let { results, models }: Props = $props();

	const chartData = $derived.by(() => {
		const successful = results.filter((r) => r.status === 'success');
		if (successful.length === 0) return [];

		const maxIter = Math.max(...successful.map((r) => r.iteration));
		const rows: Record<string, unknown>[] = [];

		for (let iter = 1; iter <= maxIter; iter++) {
			const row: Record<string, unknown> = { iteration: `${iter}` };
			for (const model of models) {
				const key = model.split('/').pop() || model;
				const iterResults = successful.filter((r) => r.model === model && r.iteration === iter);
				if (iterResults.length > 0) {
					row[key] =
						iterResults.reduce((sum, r) => sum + r.metrics.total_latency_ms, 0) /
						iterResults.length;
				}
			}
			rows.push(row);
		}
		return rows;
	});

	const series = $derived(
		models.map((model, i) => ({
			key: model.split('/').pop() || model,
			label: model.split('/').pop() || model,
			color: getModelColor(i)
		}))
	);
</script>

<div class="relative w-full h-full">
	{#if chartData.length > 0}
		<LineChart
			data={chartData}
			x="iteration"
			padding={{ left: 56, top: 8, bottom: 36, right: 16 }}
			{series}
			legend
			points
			props={{
				yAxis: { format: (d: number) => `${d.toFixed(0)}ms` }
			}}
		/>
	{/if}
</div>

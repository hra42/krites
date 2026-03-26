<script lang="ts">
	import { LineChart } from 'layerchart';
	import type { ModelTrendPoint } from '$lib/types';

	interface Props {
		trends: ModelTrendPoint[];
		metric?: 'latency' | 'tps';
	}

	let { trends, metric = 'latency' }: Props = $props();

	const isLatency = $derived(metric === 'latency');
	const yKey = $derived(isLatency ? 'avg_latency_ms' : 'avg_tokens_per_second');
	const yLabel = $derived(isLatency ? 'Latency (ms)' : 'Tokens/s');

	const chartData = $derived(
		[...trends].reverse().map((t) => {
			const d = new Date(t.created_at);
			return {
				date: d.toLocaleDateString('en-US', { day: '2-digit', month: '2-digit' }),
				avg_latency_ms: t.avg_latency_ms,
				avg_tokens_per_second: t.avg_tokens_per_second
			};
		})
	);
</script>

<div class="relative w-full h-full">
	{#if chartData.length > 0}
		<LineChart
			data={chartData}
			x="date"
			padding={{ left: 56, top: 8, bottom: 36, right: 16 }}
			series={[{ key: yKey, label: yLabel, color: '#a78bfa' }]}
			points
			props={{
				yAxis: {
					format: (d: number) => (isLatency ? `${d.toFixed(0)}ms` : d.toFixed(1))
				}
			}}
		/>
	{/if}
</div>

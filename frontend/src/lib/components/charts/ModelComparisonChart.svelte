<script lang="ts">
	import { BarChart } from 'layerchart';
	import type { CrossRunModelStats } from '$lib/types';

	interface Props {
		stats: CrossRunModelStats[];
	}

	let { stats }: Props = $props();

	const chartData = $derived(
		stats.map((s) => ({
			model: s.model.split('/').pop() || s.model,
			avg_latency_ms: s.avg_latency_ms
		}))
	);
</script>

<div class="chart-wrapper">
	<BarChart
		data={chartData}
		x="avg_latency_ms"
		y="model"
		orientation="horizontal"
		padding={{ left: 140, top: 4, bottom: 36, right: 16 }}
		series={[{ key: 'avg_latency_ms', label: 'Avg Latency (ms)', color: '#a78bfa' }]}
		props={{
			xAxis: { format: (d: number) => `${d.toFixed(0)}ms` }
		}}
	/>
</div>

<style>
	.chart-wrapper {
		position: relative;
		width: 100%;
		height: 100%;
	}
</style>

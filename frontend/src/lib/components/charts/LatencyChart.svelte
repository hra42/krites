<script lang="ts">
	import { BarChart } from 'layerchart';
	import type { ModelSummary } from '$lib/types';
	import { MODEL_COLORS } from '$lib/utils/chart-colors';

	interface Props {
		modelSummaries: ModelSummary[];
	}

	let { modelSummaries }: Props = $props();

	const chartData = $derived(
		modelSummaries.map((m) => ({
			model: m.model.split('/').pop() || m.model,
			avg: m.avg_latency_ms,
			p50: m.p50_latency_ms,
			p95: m.p95_latency_ms
		}))
	);
</script>

<div class="chart-wrapper">
	<BarChart
		data={chartData}
		x="model"
		padding={{ left: 56, top: 8, bottom: 36, right: 16 }}
		seriesLayout="group"
		series={[
			{ key: 'avg', label: 'Avg Latency', color: MODEL_COLORS[0] },
			{ key: 'p50', label: 'P50 Latency', color: MODEL_COLORS[1] },
			{ key: 'p95', label: 'P95 Latency', color: MODEL_COLORS[2] }
		]}
		legend
		props={{
			yAxis: { format: (d: number) => `${d.toFixed(0)}ms` }
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

<script lang="ts">
	import { BarChart } from 'layerchart';
	import type { ModelSummary } from '$lib/types';

	interface Props {
		modelSummaries: ModelSummary[];
	}

	let { modelSummaries }: Props = $props();

	const chartData = $derived(
		modelSummaries.map((m) => ({
			model: m.model.split('/').pop() || m.model,
			total_cost: m.total_cost
		}))
	);
</script>

<div class="chart-wrapper">
	<BarChart
		data={chartData}
		x="total_cost"
		y="model"
		orientation="horizontal"
		padding={{ left: 140, top: 4, bottom: 36, right: 16 }}
		series={[{ key: 'total_cost', label: 'Total Cost', color: '#a78bfa' }]}
		props={{
			xAxis: { format: (d: number) => `$${d.toFixed(4)}` }
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

<script lang="ts">
	import { BarChart } from 'layerchart';
	import type { ModelSummary } from '$lib/types';

	interface Props {
		modelSummaries: ModelSummary[];
		pdfMode?: boolean;
	}

	let { modelSummaries, pdfMode = false }: Props = $props();

	const chartData = $derived(
		modelSummaries.map((m) => ({
			model: m.model.split('/').pop() || m.model,
			total_cost: m.total_cost
		}))
	);
</script>

<div class="relative w-full h-full">
	<BarChart
		data={chartData}
		x="total_cost"
		y="model"
		orientation="horizontal"
		padding={{ left: pdfMode ? 180 : 140, top: 8, bottom: 36, right: 24 }}
		series={[{ key: 'total_cost', label: 'Total Cost', color: '#a78bfa' }]}
		props={{
			xAxis: { format: (d: number) => `$${d.toFixed(4)}` }
		}}
	/>
</div>

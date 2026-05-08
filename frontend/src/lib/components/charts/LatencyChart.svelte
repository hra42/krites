<script lang="ts">
	import { BarChart } from 'layerchart';
	import type { ModelSummary } from '$lib/types';
	import { MODEL_COLORS } from '$lib/utils/chart-colors';

	interface Props {
		modelSummaries: ModelSummary[];
		pdfMode?: boolean;
	}

	let { modelSummaries, pdfMode = false }: Props = $props();

	const chartData = $derived(
		modelSummaries.map((m) => ({
			model: m.model.split('/').pop() || m.model,
			avg: m.avg_latency_ms,
			p50: m.p50_latency_ms,
			p95: m.p95_latency_ms
		}))
	);

	const xAxisProps = $derived(
		pdfMode
			? {
					tickLabelProps: {
						rotate: -28,
						textAnchor: 'end' as const,
						dx: -4,
						dy: 4
					}
				}
			: undefined
	);
</script>

<div class="relative w-full h-full">
	<BarChart
		data={chartData}
		x="model"
		padding={{ left: 60, top: 12, bottom: pdfMode ? 80 : 36, right: 20 }}
		seriesLayout="group"
		series={[
			{ key: 'avg', label: 'Avg Latency', color: MODEL_COLORS[0] },
			{ key: 'p50', label: 'P50 Latency', color: MODEL_COLORS[1] },
			{ key: 'p95', label: 'P95 Latency', color: MODEL_COLORS[2] }
		]}
		legend
		props={{
			yAxis: { format: (d: number) => `${d.toFixed(0)}ms` },
			xAxis: xAxisProps
		}}
	/>
</div>

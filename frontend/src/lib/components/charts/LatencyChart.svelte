<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		Chart,
		CategoryScale,
		LinearScale,
		BarElement,
		Title,
		Tooltip,
		Legend
	} from 'chart.js';
	import type { ModelSummary } from '$lib/types';
	import { getModelColor, getModelColorAlpha, DARK_CHART_DEFAULTS } from '$lib/utils/chart-theme';

	Chart.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend);

	interface Props {
		modelSummaries: ModelSummary[];
	}

	let { modelSummaries }: Props = $props();
	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	onMount(() => {
		const labels = modelSummaries.map((m) => m.model.split('/').pop() || m.model);

		chart = new Chart(canvas, {
			type: 'bar',
			data: {
				labels,
				datasets: [
					{
						label: 'Avg Latency',
						data: modelSummaries.map((m) => m.avg_latency_ms),
						backgroundColor: modelSummaries.map((_, i) => getModelColorAlpha(i, 0.8)),
						borderColor: modelSummaries.map((_, i) => getModelColor(i)),
						borderWidth: 1
					},
					{
						label: 'P50 Latency',
						data: modelSummaries.map((m) => m.p50_latency_ms),
						backgroundColor: modelSummaries.map((_, i) => getModelColorAlpha(i, 0.5)),
						borderColor: modelSummaries.map((_, i) => getModelColor(i)),
						borderWidth: 1
					},
					{
						label: 'P95 Latency',
						data: modelSummaries.map((m) => m.p95_latency_ms),
						backgroundColor: modelSummaries.map((_, i) => getModelColorAlpha(i, 0.3)),
						borderColor: modelSummaries.map((_, i) => getModelColor(i)),
						borderWidth: 1
					}
				]
			},
			options: {
				...DARK_CHART_DEFAULTS,
				scales: {
					x: {
						ticks: { color: '#5a5766' },
						grid: { color: '#2a2830' }
					},
					y: {
						ticks: { color: '#5a5766' },
						grid: { color: '#2a2830' },
						title: {
							display: true,
							text: 'Latency (ms)',
							color: '#8b8894'
						}
					}
				}
			}
		});
	});

	onDestroy(() => {
		chart?.destroy();
	});
</script>

<div class="chart-wrapper">
	<canvas bind:this={canvas}></canvas>
</div>

<style>
	.chart-wrapper {
		position: relative;
		width: 100%;
		height: 100%;
	}
</style>

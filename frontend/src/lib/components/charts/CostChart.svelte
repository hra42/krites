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
	import { getModelColor, getModelColorAlpha } from '$lib/utils/chart-theme';

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
						label: 'Total Cost',
						data: modelSummaries.map((m) => m.total_cost),
						backgroundColor: modelSummaries.map((_, i) => getModelColorAlpha(i, 0.8)),
						borderColor: modelSummaries.map((_, i) => getModelColor(i)),
						borderWidth: 1
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				indexAxis: 'y',
				scales: {
					x: {
						ticks: { color: '#5a5766' },
						grid: { color: '#2a2830' },
						title: {
							display: true,
							text: 'Cost ($)',
							color: '#8b8894'
						}
					},
					y: {
						ticks: { color: '#5a5766' },
						grid: { color: '#2a2830' }
					}
				},
				plugins: {
					legend: { display: false },
					tooltip: {
						backgroundColor: '#1e1c23',
						titleColor: '#e4e2e8',
						bodyColor: '#8b8894',
						borderColor: '#2a2830',
						borderWidth: 1,
						callbacks: {
							label: (ctx) => `$${(ctx.raw as number).toFixed(4)}`
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

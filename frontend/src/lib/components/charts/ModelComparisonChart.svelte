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
	import type { CrossRunModelStats } from '$lib/types';
	import { getModelColor, getModelColorAlpha, DARK_CHART_DEFAULTS } from '$lib/utils/chart-theme';

	Chart.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend);

	interface Props {
		stats: CrossRunModelStats[];
	}

	let { stats }: Props = $props();
	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	onMount(() => {
		const labels = stats.map((s) => s.model.split('/').pop() || s.model);

		chart = new Chart(canvas, {
			type: 'bar',
			data: {
				labels,
				datasets: [
					{
						label: 'Avg Latency (ms)',
						data: stats.map((s) => s.avg_latency_ms),
						backgroundColor: stats.map((_, i) => getModelColorAlpha(i, 0.8)),
						borderColor: stats.map((_, i) => getModelColor(i)),
						borderWidth: 1
					}
				]
			},
			options: {
				...DARK_CHART_DEFAULTS,
				indexAxis: 'y',
				scales: {
					x: {
						ticks: { color: '#5a5766' },
						grid: { color: '#2a2830' },
						title: {
							display: true,
							text: 'Latency (ms)',
							color: '#8b8894'
						}
					},
					y: {
						ticks: { color: '#5a5766' },
						grid: { color: '#2a2830' }
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

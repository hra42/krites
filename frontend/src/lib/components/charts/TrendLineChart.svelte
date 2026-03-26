<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		Chart,
		CategoryScale,
		LinearScale,
		PointElement,
		LineElement,
		Title,
		Tooltip,
		Legend
	} from 'chart.js';
	import type { ModelTrendPoint } from '$lib/types';
	import { DARK_CHART_DEFAULTS } from '$lib/utils/chart-theme';

	Chart.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend);

	interface Props {
		trends: ModelTrendPoint[];
		metric?: 'latency' | 'tps';
	}

	let { trends, metric = 'latency' }: Props = $props();
	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	onMount(() => {
		const sorted = [...trends].reverse();
		const labels = sorted.map((t) => {
			const d = new Date(t.created_at);
			return d.toLocaleDateString('en-US', { day: '2-digit', month: '2-digit' });
		});

		const isLatency = metric === 'latency';

		chart = new Chart(canvas, {
			type: 'line',
			data: {
				labels,
				datasets: [
					{
						label: isLatency ? 'Avg Latency (ms)' : 'Avg tok/s',
						data: sorted.map((t) => (isLatency ? t.avg_latency_ms : t.avg_tokens_per_second)),
						borderColor: '#a78bfa',
						backgroundColor: 'rgba(167, 139, 250, 0.1)',
						fill: true,
						tension: 0.3,
						pointRadius: 4,
						pointHoverRadius: 6
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
							text: isLatency ? 'Latency (ms)' : 'Tokens/s',
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

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
	import type { Result } from '$lib/types';
	import { getModelColor } from '$lib/utils/chart-theme';

	Chart.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend);

	interface Props {
		results: Result[];
		models: string[];
	}

	let { results, models }: Props = $props();
	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	onMount(() => {
		const successful = results.filter((r) => r.status === 'success');
		if (successful.length === 0) return;

		const maxIter = Math.max(...successful.map((r) => r.iteration));
		const labels = Array.from({ length: maxIter }, (_, i) => `${i + 1}`);

		const datasets = models.map((model, colorIdx) => {
			const modelResults = successful.filter((r) => r.model === model);
			const iterData: (number | null)[] = [];

			for (let iter = 1; iter <= maxIter; iter++) {
				const iterResults = modelResults.filter((r) => r.iteration === iter);
				if (iterResults.length === 0) {
					iterData.push(null);
				} else {
					const avg =
						iterResults.reduce((sum, r) => sum + r.metrics.total_latency_ms, 0) /
						iterResults.length;
					iterData.push(avg);
				}
			}

			return {
				label: model.split('/').pop() || model,
				data: iterData,
				borderColor: getModelColor(colorIdx),
				backgroundColor: getModelColor(colorIdx),
				borderWidth: 2,
				pointRadius: 4,
				tension: 0.2,
				spanGaps: true
			};
		});

		chart = new Chart(canvas, {
			type: 'line',
			data: { labels, datasets },
			options: {
				responsive: true,
				maintainAspectRatio: false,
				scales: {
					x: {
						ticks: { color: '#5a5766' },
						grid: { color: '#2a2830' },
						title: {
							display: true,
							text: 'Iteration',
							color: '#8b8894'
						}
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
				},
				plugins: {
					legend: {
						labels: {
							color: '#8b8894',
							font: { family: "'JetBrains Mono', monospace", size: 11 }
						}
					},
					tooltip: {
						backgroundColor: '#1e1c23',
						titleColor: '#e4e2e8',
						bodyColor: '#8b8894',
						borderColor: '#2a2830',
						borderWidth: 1
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

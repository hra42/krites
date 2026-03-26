<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		Chart,
		RadialLinearScale,
		PointElement,
		LineElement,
		Filler,
		Tooltip,
		Legend
	} from 'chart.js';
	import type { ModelSummary } from '$lib/types';
	import { getModelColor, getModelColorAlpha } from '$lib/utils/chart-theme';

	Chart.register(RadialLinearScale, PointElement, LineElement, Filler, Tooltip, Legend);

	interface Props {
		modelSummaries: ModelSummary[];
		criteria: string[];
	}

	let { modelSummaries, criteria }: Props = $props();
	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	const hasScores = $derived(
		modelSummaries.some((m) => m.avg_judge_scores && Object.keys(m.avg_judge_scores).length > 0)
	);

	onMount(() => {
		if (!hasScores || criteria.length === 0) return;

		const datasets = modelSummaries
			.filter((m) => m.avg_judge_scores && Object.keys(m.avg_judge_scores).length > 0)
			.map((m, i) => ({
				label: m.model.split('/').pop() || m.model,
				data: criteria.map((c) => m.avg_judge_scores?.[c] ?? 0),
				backgroundColor: getModelColorAlpha(i, 0.15),
				borderColor: getModelColor(i),
				borderWidth: 2,
				pointBackgroundColor: getModelColor(i),
				pointRadius: 4
			}));

		chart = new Chart(canvas!, {
			type: 'radar',
			data: {
				labels: criteria,
				datasets
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				scales: {
					r: {
						min: 0,
						max: 10,
						ticks: {
							stepSize: 2,
							color: '#5a5766',
							backdropColor: 'transparent'
						},
						grid: { color: '#2a2830' },
						angleLines: { color: '#2a2830' },
						pointLabels: {
							color: '#8b8894',
							font: { family: "'JetBrains Mono', monospace", size: 11 }
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

{#if hasScores && criteria.length > 0}
	<div class="chart-wrapper">
		<canvas bind:this={canvas}></canvas>
	</div>
{:else}
	<div class="no-data">No judge scores available</div>
{/if}

<style>
	.chart-wrapper {
		position: relative;
		width: 100%;
		height: 100%;
	}

	.no-data {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: var(--color-text-dim);
		font-size: 13px;
	}
</style>

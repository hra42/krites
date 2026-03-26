<script lang="ts">
	import type { ModelSummary } from '$lib/types';
	import { getModelColor } from '$lib/utils/chart-colors';

	interface Props {
		modelSummaries: ModelSummary[];
		criteria: string[];
	}

	let { modelSummaries, criteria }: Props = $props();

	const hasScores = $derived(
		modelSummaries.some((m) => m.avg_judge_scores && Object.keys(m.avg_judge_scores).length > 0)
	);

	const modelsWithScores = $derived(
		modelSummaries.filter(
			(m) => m.avg_judge_scores && Object.keys(m.avg_judge_scores).length > 0
		)
	);

	const maxScore = 10;
	const levels = [2, 4, 6, 8, 10];
	const cx = 150;
	const cy = 150;
	const radius = 120;

	function angleFor(i: number, total: number): number {
		return (Math.PI * 2 * i) / total - Math.PI / 2;
	}

	function pointOnCircle(angle: number, r: number): { x: number; y: number } {
		return {
			x: cx + r * Math.cos(angle),
			y: cy + r * Math.sin(angle)
		};
	}

	function polygonPoints(values: number[]): string {
		return values
			.map((v, i) => {
				const angle = angleFor(i, values.length);
				const r = (v / maxScore) * radius;
				const p = pointOnCircle(angle, r);
				return `${p.x},${p.y}`;
			})
			.join(' ');
	}

	function gridPolygon(level: number): string {
		const r = (level / maxScore) * radius;
		return criteria
			.map((_, i) => {
				const angle = angleFor(i, criteria.length);
				const p = pointOnCircle(angle, r);
				return `${p.x},${p.y}`;
			})
			.join(' ');
	}
</script>

{#if hasScores && criteria.length > 0}
	<div class="chart-wrapper">
		<svg viewBox="0 0 300 300" class="radar-svg">
			<!-- Grid levels -->
			{#each levels as level}
				<polygon points={gridPolygon(level)} fill="none" stroke="#2a2830" stroke-width="1" />
			{/each}

			<!-- Axis lines -->
			{#each criteria as _, i}
				{@const angle = angleFor(i, criteria.length)}
				{@const end = pointOnCircle(angle, radius)}
				<line x1={cx} y1={cy} x2={end.x} y2={end.y} stroke="#2a2830" stroke-width="1" />
			{/each}

			<!-- Data polygons -->
			{#each modelsWithScores as m, mi}
				{@const values = criteria.map((c) => m.avg_judge_scores?.[c] ?? 0)}
				{@const color = getModelColor(mi)}
				<polygon
					points={polygonPoints(values)}
					fill={color}
					fill-opacity="0.15"
					stroke={color}
					stroke-width="2"
				/>
				{#each values as v, vi}
					{@const angle = angleFor(vi, criteria.length)}
					{@const p = pointOnCircle(angle, (v / maxScore) * radius)}
					<circle cx={p.x} cy={p.y} r="3" fill={color} />
				{/each}
			{/each}

			<!-- Labels -->
			{#each criteria as c, i}
				{@const angle = angleFor(i, criteria.length)}
				{@const p = pointOnCircle(angle, radius + 16)}
				<text
					x={p.x}
					y={p.y}
					text-anchor="middle"
					dominant-baseline="central"
					fill="#8b8894"
					font-size="11"
					font-family="'JetBrains Mono', monospace"
				>
					{c}
				</text>
			{/each}
		</svg>

		<!-- Legend -->
		{#if modelsWithScores.length > 1}
			<div class="legend">
				{#each modelsWithScores as m, i}
					<div class="legend-item">
						<span class="legend-dot" style:background={getModelColor(i)}></span>
						<span class="legend-label">{m.model.split('/').pop() || m.model}</span>
					</div>
				{/each}
			</div>
		{/if}
	</div>
{:else}
	<div class="no-data">No judge scores available</div>
{/if}

<style>
	.chart-wrapper {
		position: relative;
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
	}

	.radar-svg {
		width: 100%;
		max-width: 300px;
		flex: 1;
		min-height: 0;
	}

	.legend {
		display: flex;
		gap: 16px;
		flex-wrap: wrap;
		justify-content: center;
		padding-top: 8px;
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 11px;
		color: #8b8894;
		font-family: 'JetBrains Mono', monospace;
	}

	.legend-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
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

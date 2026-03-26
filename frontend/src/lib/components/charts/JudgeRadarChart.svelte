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
	<div class="relative w-full h-full flex flex-col items-center">
		<svg viewBox="0 0 300 300" class="w-full max-w-[300px] flex-1 min-h-0">
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
					font-size="13"
					font-family="'JetBrains Mono', monospace"
				>
					{c}
				</text>
			{/each}
		</svg>

		<!-- Legend -->
		{#if modelsWithScores.length > 1}
			<div class="flex gap-4 flex-wrap justify-center pt-2">
				{#each modelsWithScores as m, i}
					<div class="flex items-center gap-1.5 text-sm text-text-muted font-mono">
						<span class="w-2 h-2 rounded-full shrink-0" style:background={getModelColor(i)}></span>
						<span>{m.model.split('/').pop() || m.model}</span>
					</div>
				{/each}
			</div>
		{/if}
	</div>
{:else}
	<div class="flex items-center justify-center h-full text-text-dim text-base">No judge scores available</div>
{/if}

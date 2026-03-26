import type { ChartOptions } from 'chart.js';

export const MODEL_COLORS = [
	'#a78bfa', // violet (accent)
	'#34d399', // emerald
	'#fbbf24', // amber
	'#f87171', // rose
	'#38bdf8', // sky
	'#fb923c', // orange
	'#a3e635', // lime
	'#e879f9'  // fuchsia
];

export function getModelColor(index: number): string {
	return MODEL_COLORS[index % MODEL_COLORS.length];
}

export function getModelColorAlpha(index: number, alpha: number): string {
	const hex = MODEL_COLORS[index % MODEL_COLORS.length];
	const r = parseInt(hex.slice(1, 3), 16);
	const g = parseInt(hex.slice(3, 5), 16);
	const b = parseInt(hex.slice(5, 7), 16);
	return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

export const DARK_CHART_DEFAULTS: Partial<ChartOptions> = {
	responsive: true,
	maintainAspectRatio: false,
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
	},
	scales: {
		x: {
			ticks: { color: '#5a5766' },
			grid: { color: '#2a2830' }
		},
		y: {
			ticks: { color: '#5a5766' },
			grid: { color: '#2a2830' }
		}
	}
};

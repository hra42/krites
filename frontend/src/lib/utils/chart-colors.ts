export const MODEL_COLORS = [
	'#a78bfa', // violet (accent)
	'#34d399', // emerald
	'#fbbf24', // amber
	'#f87171', // rose
	'#38bdf8', // sky
	'#fb923c', // orange
	'#a3e635', // lime
	'#e879f9' // fuchsia
];

export function getModelColor(index: number): string {
	return MODEL_COLORS[index % MODEL_COLORS.length];
}

<script lang="ts">
	import '../app.css';
	import Toast from '$lib/components/Toast.svelte';
	import { LayoutDashboardIcon, FolderOpenIcon, PlayIcon, BarChart3Icon, FlaskConicalIcon } from '@lucide/svelte';

	let { children } = $props();
</script>

<svelte:head>
	<title>Krites</title>
</svelte:head>

<div class="flex min-h-screen">
	<aside class="w-60 min-w-60 bg-bg-elevated border-r border-border py-6 px-4 flex flex-col gap-8">
		<div class="flex items-center gap-2.5 px-2">
			<FlaskConicalIcon size={22} color="var(--color-accent)" />
			<h1 class="text-lg font-bold text-text tracking-[-0.03em]">Krites</h1>
		</div>
		<nav class="flex flex-col gap-1">
			<a href="/" class="flex items-center gap-3 px-4 py-3.5 rounded-[--radius] text-text-muted text-lg font-medium transition-all duration-150 hover:text-text hover:bg-bg-card">
				<LayoutDashboardIcon size={20} />
				Dashboard
			</a>
			<a href="/suites" class="flex items-center gap-3 px-4 py-3.5 rounded-[--radius] text-text-muted text-lg font-medium transition-all duration-150 hover:text-text hover:bg-bg-card">
				<FolderOpenIcon size={20} />
				Suites
			</a>
			<a href="/runs" class="flex items-center gap-3 px-4 py-3.5 rounded-[--radius] text-text-muted text-lg font-medium transition-all duration-150 hover:text-text hover:bg-bg-card">
				<PlayIcon size={20} />
				Runs
			</a>
			<a href="/analytics" class="flex items-center gap-3 px-4 py-3.5 rounded-[--radius] text-text-muted text-lg font-medium transition-all duration-150 hover:text-text hover:bg-bg-card">
				<BarChart3Icon size={20} />
				Analytics
			</a>
		</nav>
	</aside>
	<main class="flex-1 py-8 px-10 overflow-y-auto">
		{@render children()}
	</main>
</div>

<Toast />

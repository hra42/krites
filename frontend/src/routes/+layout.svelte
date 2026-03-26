<script lang="ts">
	import '../app.css';
	import Toast from '$lib/components/Toast.svelte';
	import { LayoutDashboardIcon, FolderOpenIcon, PlayIcon, BarChart3Icon, FlaskConicalIcon } from '@lucide/svelte';

	let { children } = $props();
</script>

<svelte:head>
	<title>Krites</title>
</svelte:head>

<div class="app">
	<aside class="sidebar">
		<div class="logo">
			<FlaskConicalIcon size={22} color="var(--color-accent)" />
			<h1>Krites</h1>
		</div>
		<nav>
			<a href="/" class="nav-link">
				<LayoutDashboardIcon size={20} />
				Dashboard
			</a>
			<a href="/suites" class="nav-link">
				<FolderOpenIcon size={20} />
				Suites
			</a>
			<a href="/runs" class="nav-link">
				<PlayIcon size={20} />
				Runs
			</a>
			<a href="/analytics" class="nav-link">
				<BarChart3Icon size={20} />
				Analytics
			</a>
		</nav>
	</aside>
	<main class="content">
		{@render children()}
	</main>
</div>

<Toast />

<style>
	.app {
		display: flex;
		min-height: 100vh;
	}

	.sidebar {
		width: 240px;
		min-width: 240px;
		background: var(--color-bg-elevated);
		border-right: 1px solid var(--color-border);
		padding: 24px 16px;
		display: flex;
		flex-direction: column;
		gap: 32px;
	}

	.logo {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 0 8px;
	}

	.logo h1 {
		font-size: 16px;
		font-weight: 700;
		color: var(--color-text);
		letter-spacing: -0.03em;
	}

	nav {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.nav-link {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 16px;
		border-radius: var(--radius);
		color: var(--color-text-muted);
		font-size: 18px;
		font-weight: 500;
		transition: all 0.15s ease;
	}

	.nav-link:hover {
		color: var(--color-text);
		background: var(--color-bg-card);
	}


	.content {
		flex: 1;
		padding: 32px 40px;
		overflow-y: auto;
	}
</style>

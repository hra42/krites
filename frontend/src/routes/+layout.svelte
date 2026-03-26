<script lang="ts">
	import '../app.css';
	import Toast from '$lib/components/Toast.svelte';

	let { children } = $props();
</script>

<svelte:head>
	<title>Krites</title>
</svelte:head>

<div class="app">
	<aside class="sidebar">
		<div class="logo">
			<span class="logo-icon">&#9881;</span>
			<h1>Krites</h1>
		</div>
		<nav>
			<a href="/" class="nav-link">
				<span class="nav-icon">&#9632;</span>
				Dashboard
			</a>
			<a href="/suites" class="nav-link">
				<span class="nav-icon">&#9654;</span>
				Suites
			</a>
			<a href="/runs" class="nav-link">
				<span class="nav-icon">&#9200;</span>
				Runs
			</a>
			<a href="/analytics" class="nav-link">
				<span class="nav-icon">&#9733;</span>
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

	.logo-icon {
		font-size: 20px;
		color: var(--color-accent);
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
		gap: 2px;
	}

	.nav-link {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: var(--radius);
		color: var(--color-text-muted);
		font-size: 14px;
		font-weight: 500;
		transition: all 0.15s ease;
	}

	.nav-link:hover {
		color: var(--color-text);
		background: var(--color-bg-card);
	}


	.nav-icon {
		font-size: 12px;
		width: 20px;
		text-align: center;
	}

	.content {
		flex: 1;
		padding: 32px 40px;
		overflow-y: auto;
	}
</style>

<script lang="ts">
	import { onMount } from 'svelte';
	import * as api from '$lib/api/client';
	import type { OpenRouterModel } from '$lib/types';
	import ModelChip from './ModelChip.svelte';

	type SortKey = 'name' | 'context' | 'price';
	type SortDir = 'asc' | 'desc';

	interface Props {
		selected: string[];
		onchange: (models: string[]) => void;
	}

	let { selected, onchange }: Props = $props();

	let allModels = $state<OpenRouterModel[]>([]);
	let loadingModels = $state(true);
	let loadError = $state('');

	let search = $state('');
	let providerFilter = $state('');
	let sortKey = $state<SortKey>('name');
	let sortDir = $state<SortDir>('asc');
	let showDropdown = $state(false);
	let manualInput = $state('');

	let providers = $derived(() => {
		const set = new Set<string>();
		for (const m of allModels) {
			const provider = m.id.split('/')[0];
			if (provider) set.add(provider);
		}
		return [...set].sort();
	});

	let filtered = $derived(() => {
		let list = allModels;

		if (providerFilter) {
			list = list.filter((m) => m.id.startsWith(providerFilter + '/'));
		}

		if (search) {
			const q = search.toLowerCase();
			list = list.filter(
				(m) => m.id.toLowerCase().includes(q) || m.name.toLowerCase().includes(q)
			);
		}

		list = list.filter((m) => !selected.includes(m.id));

		const sorted = [...list];
		sorted.sort((a, b) => {
			let cmp = 0;
			if (sortKey === 'name') {
				cmp = a.name.localeCompare(b.name);
			} else if (sortKey === 'context') {
				cmp = (a.context_length ?? 0) - (b.context_length ?? 0);
			} else if (sortKey === 'price') {
				cmp = parseFloat(a.pricing.prompt || '0') - parseFloat(b.pricing.prompt || '0');
			}
			return sortDir === 'asc' ? cmp : -cmp;
		});

		return sorted;
	});

	onMount(async () => {
		try {
			const resp = await api.listModels();
			allModels = resp.data || [];
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load models';
		} finally {
			loadingModels = false;
		}
	});

	function addModel(id: string) {
		if (!selected.includes(id)) {
			onchange([...selected, id]);
		}
		search = '';
	}

	function removeModel(index: number) {
		onchange(selected.filter((_, i) => i !== index));
	}

	function addManualModel() {
		const trimmed = manualInput.trim();
		if (trimmed && !selected.includes(trimmed)) {
			onchange([...selected, trimmed]);
			manualInput = '';
		}
	}

	function handleManualKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			addManualModel();
		}
	}

	function toggleSort(key: SortKey) {
		if (sortKey === key) {
			sortDir = sortDir === 'asc' ? 'desc' : 'asc';
		} else {
			sortKey = key;
			sortDir = 'asc';
		}
	}

	function formatContext(len: number | null | undefined): string {
		if (!len) return '-';
		if (len >= 1000000) return `${(len / 1000000).toFixed(1)}M`;
		if (len >= 1000) return `${(len / 1000).toFixed(0)}K`;
		return `${len}`;
	}

	function formatPrice(price: string): string {
		const n = parseFloat(price);
		if (isNaN(n) || n === 0) return 'Free';
		if (n < 0.001) return `$${(n * 1000000).toFixed(2)}/M`;
		return `$${n.toFixed(4)}/tok`;
	}

	function sortIndicator(key: SortKey): string {
		if (sortKey !== key) return '';
		return sortDir === 'asc' ? ' ▲' : ' ▼';
	}
</script>

<div class="flex flex-col gap-2">
	{#if selected.length > 0}
		<div class="flex flex-wrap gap-1.5 mb-1">
			{#each selected as model, i}
				<ModelChip {model} removable onremove={() => removeModel(i)} />
			{/each}
		</div>
	{/if}

	<div class="flex gap-2 items-stretch">
		<button type="button" class="btn whitespace-nowrap flex items-center gap-1.5" onclick={() => (showDropdown = !showDropdown)}>
			{showDropdown ? 'Close' : 'Browse Models'}
			<span class="text-xs">{showDropdown ? '▴' : '▾'}</span>
		</button>
		<div class="flex gap-1 flex-1">
			<input
				class="input flex-1"
				bind:value={manualInput}
				onkeydown={handleManualKeydown}
				placeholder="Enter model ID manually"
			/>
			<button type="button" class="btn" onclick={addManualModel}>+</button>
		</div>
	</div>

	{#if showDropdown}
		<div class="bg-bg-elevated border border-border rounded-[--radius] overflow-hidden fade-in">
			{#if loadingModels}
				<p class="py-5 text-center text-text-muted text-base">Loading models...</p>
			{:else if loadError}
				<p class="py-5 text-center text-error text-base">{loadError}</p>
			{:else}
				<div class="flex gap-2 px-3 py-2.5 border-b border-border">
					<input
						class="input flex-1"
						bind:value={search}
						placeholder="Search by name or ID..."
					/>
					<select class="input w-40 cursor-pointer" bind:value={providerFilter}>
						<option value="">All Providers</option>
						{#each providers() as provider}
							<option value={provider}>{provider}</option>
						{/each}
					</select>
				</div>

				<div class="grid grid-cols-[1fr_80px_100px_32px] gap-2 px-3 py-1.5 border-b border-border bg-bg">
					<button type="button" class="bg-transparent border-none text-text-muted text-sm font-semibold uppercase tracking-wide text-left cursor-pointer py-1 hover:text-accent" onclick={() => toggleSort('name')}>
						Model{sortIndicator('name')}
					</button>
					<button type="button" class="bg-transparent border-none text-text-muted text-sm font-semibold uppercase tracking-wide text-left cursor-pointer py-1 hover:text-accent" onclick={() => toggleSort('context')}>
						Context{sortIndicator('context')}
					</button>
					<button type="button" class="bg-transparent border-none text-text-muted text-sm font-semibold uppercase tracking-wide text-left cursor-pointer py-1 hover:text-accent" onclick={() => toggleSort('price')}>
						Price/Prompt{sortIndicator('price')}
					</button>
					<span></span>
				</div>

				<div class="max-h-80 overflow-y-auto">
					{#each filtered().slice(0, 100) as model (model.id)}
						<button
							type="button"
							class="grid grid-cols-[1fr_80px_100px_32px] gap-2 items-center w-full py-2 px-3 border-none border-b border-border bg-transparent text-text cursor-pointer text-left transition-[background] duration-100 last:border-b-0 hover:bg-accent-bg"
							onclick={() => addModel(model.id)}
						>
							<div class="flex flex-col gap-px min-w-0">
								<span class="text-base font-medium truncate">{model.name}</span>
								<span class="text-sm text-text-muted truncate mono">{model.id}</span>
							</div>
							<span class="text-sm text-text-muted text-right mono">{formatContext(model.context_length)}</span>
							<span class="text-sm text-text-muted text-right mono">{formatPrice(model.pricing.prompt)}</span>
							<span class="text-lg text-accent text-center font-semibold">+</span>
						</button>
					{/each}
					{#if filtered().length === 0}
						<p class="py-5 text-center text-text-muted text-base">No models found</p>
					{/if}
					{#if filtered().length > 100}
						<p class="py-5 text-center text-text-muted text-base">
							{filtered().length - 100} more models — refine search
						</p>
					{/if}
				</div>

				<div class="flex justify-between px-3 py-2 border-t border-border text-sm text-text-dim">
					<span class="mono">{allModels.length} models available</span>
					<span class="mono">{selected.length} selected</span>
				</div>
			{/if}
		</div>
	{/if}
</div>

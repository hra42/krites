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

<div class="model-picker">
	{#if selected.length > 0}
		<div class="selected-chips">
			{#each selected as model, i}
				<ModelChip {model} removable onremove={() => removeModel(i)} />
			{/each}
		</div>
	{/if}

	<div class="picker-controls">
		<button type="button" class="btn btn-browse" onclick={() => (showDropdown = !showDropdown)}>
			{showDropdown ? 'Close' : 'Browse Models'}
			<span class="arrow">{showDropdown ? '▴' : '▾'}</span>
		</button>
		<div class="manual-add">
			<input
				class="input"
				bind:value={manualInput}
				onkeydown={handleManualKeydown}
				placeholder="Enter model ID manually"
			/>
			<button type="button" class="btn" onclick={addManualModel}>+</button>
		</div>
	</div>

	{#if showDropdown}
		<div class="dropdown fade-in">
			{#if loadingModels}
				<p class="dropdown-status">Loading models...</p>
			{:else if loadError}
				<p class="dropdown-status error">{loadError}</p>
			{:else}
				<div class="dropdown-filters">
					<input
						class="input search-input"
						bind:value={search}
						placeholder="Search by name or ID..."
					/>
					<select class="input provider-select" bind:value={providerFilter}>
						<option value="">All Providers</option>
						{#each providers() as provider}
							<option value={provider}>{provider}</option>
						{/each}
					</select>
				</div>

				<div class="dropdown-header">
					<button type="button" class="sort-btn" onclick={() => toggleSort('name')}>
						Model{sortIndicator('name')}
					</button>
					<button type="button" class="sort-btn" onclick={() => toggleSort('context')}>
						Context{sortIndicator('context')}
					</button>
					<button type="button" class="sort-btn" onclick={() => toggleSort('price')}>
						Price/Prompt{sortIndicator('price')}
					</button>
					<span></span>
				</div>

				<div class="dropdown-list">
					{#each filtered().slice(0, 100) as model (model.id)}
						<button type="button" class="model-row" onclick={() => addModel(model.id)}>
							<div class="model-info">
								<span class="model-name">{model.name}</span>
								<span class="model-id mono">{model.id}</span>
							</div>
							<span class="model-context mono">{formatContext(model.context_length)}</span>
							<span class="model-price mono">{formatPrice(model.pricing.prompt)}</span>
							<span class="model-add">+</span>
						</button>
					{/each}
					{#if filtered().length === 0}
						<p class="dropdown-status">No models found</p>
					{/if}
					{#if filtered().length > 100}
						<p class="dropdown-status">
							{filtered().length - 100} more models — refine search
						</p>
					{/if}
				</div>

				<div class="dropdown-footer">
					<span class="mono">{allModels.length} models available</span>
					<span class="mono">{selected.length} selected</span>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.model-picker {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.selected-chips {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-bottom: 4px;
	}

	.picker-controls {
		display: flex;
		gap: 8px;
		align-items: stretch;
	}

	.btn-browse {
		white-space: nowrap;
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.arrow {
		font-size: 10px;
	}

	.manual-add {
		display: flex;
		gap: 4px;
		flex: 1;
	}

	.manual-add .input {
		flex: 1;
		font-size: 13px;
	}

	.dropdown {
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		overflow: hidden;
	}

	.dropdown-filters {
		display: flex;
		gap: 8px;
		padding: 10px 12px;
		border-bottom: 1px solid var(--color-border);
	}

	.search-input {
		flex: 1;
	}

	.provider-select {
		width: 160px;
		cursor: pointer;
	}

	.dropdown-header {
		display: grid;
		grid-template-columns: 1fr 80px 100px 32px;
		gap: 8px;
		padding: 6px 12px;
		border-bottom: 1px solid var(--color-border);
		background: var(--color-bg);
	}

	.sort-btn {
		background: none;
		border: none;
		color: var(--color-text-muted);
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		text-align: left;
		cursor: pointer;
		padding: 4px 0;
	}

	.sort-btn:hover {
		color: var(--color-accent);
	}

	.dropdown-list {
		max-height: 320px;
		overflow-y: auto;
	}

	.model-row {
		display: grid;
		grid-template-columns: 1fr 80px 100px 32px;
		gap: 8px;
		align-items: center;
		width: 100%;
		padding: 8px 12px;
		border: none;
		border-bottom: 1px solid var(--color-border);
		background: transparent;
		color: var(--color-text);
		cursor: pointer;
		text-align: left;
		transition: background 0.1s;
	}

	.model-row:last-child {
		border-bottom: none;
	}

	.model-row:hover {
		background: var(--color-accent-bg);
	}

	.model-info {
		display: flex;
		flex-direction: column;
		gap: 1px;
		min-width: 0;
	}

	.model-name {
		font-size: 13px;
		font-weight: 500;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.model-id {
		font-size: 11px;
		color: var(--color-text-muted);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.model-context,
	.model-price {
		font-size: 12px;
		color: var(--color-text-muted);
		text-align: right;
	}

	.model-add {
		font-size: 16px;
		color: var(--color-accent);
		text-align: center;
		font-weight: 600;
	}

	.dropdown-footer {
		display: flex;
		justify-content: space-between;
		padding: 8px 12px;
		border-top: 1px solid var(--color-border);
		font-size: 11px;
		color: var(--color-text-dim);
	}

	.dropdown-status {
		padding: 20px;
		text-align: center;
		color: var(--color-text-muted);
		font-size: 13px;
	}

	.dropdown-status.error {
		color: var(--color-error);
	}
</style>

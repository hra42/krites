<script lang="ts">
	import { toasts } from '$lib/stores/toast';

	const typeClasses: Record<string, string> = {
		success: 'bg-success-bg border border-success text-success-text',
		error: 'bg-error-bg border border-error text-error-text',
		info: 'bg-bg-elevated border border-border text-text'
	};
</script>

{#if $toasts.length > 0}
	<div class="fixed bottom-6 right-6 z-[9999] flex flex-col gap-2 pointer-events-none">
		{#each $toasts as toast (toast.id)}
			<div class="flex items-center gap-3 px-5 py-3.5 rounded-[--radius] text-base font-medium pointer-events-auto min-w-[300px] max-w-[480px] shadow-[0_8px_24px_rgba(0,0,0,0.4)] fade-in {typeClasses[toast.type]}">
				<span class="text-lg shrink-0">
					{#if toast.type === 'success'}&#10003;{:else if toast.type === 'error'}&#10007;{:else}&#9432;{/if}
				</span>
				<span class="leading-[1.4]">{toast.message}</span>
			</div>
		{/each}
	</div>
{/if}

<script lang="ts">
	import { toasts } from '$lib/stores/toast';
</script>

{#if $toasts.length > 0}
	<div class="toast-container">
		{#each $toasts as toast (toast.id)}
			<div class="toast toast-{toast.type} fade-in">
				<span class="toast-icon">
					{#if toast.type === 'success'}&#10003;{:else if toast.type === 'error'}&#10007;{:else}&#9432;{/if}
				</span>
				<span class="toast-message">{toast.message}</span>
			</div>
		{/each}
	</div>
{/if}

<style>
	.toast-container {
		position: fixed;
		bottom: 24px;
		right: 24px;
		z-index: 9999;
		display: flex;
		flex-direction: column;
		gap: 8px;
		pointer-events: none;
	}

	.toast {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 20px;
		border-radius: var(--radius);
		font-size: 13px;
		font-weight: 500;
		pointer-events: auto;
		min-width: 280px;
		max-width: 420px;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
	}

	.toast-success {
		background: #065f46;
		border: 1px solid var(--color-success);
		color: #d1fae5;
	}

	.toast-error {
		background: #7f1d1d;
		border: 1px solid var(--color-error);
		color: #fecaca;
	}

	.toast-info {
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border);
		color: var(--color-text);
	}

	.toast-icon {
		font-size: 16px;
		flex-shrink: 0;
	}

	.toast-message {
		line-height: 1.4;
	}
</style>

import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info';

export interface Toast {
	id: number;
	type: ToastType;
	message: string;
}

let nextId = 0;

export const toasts = writable<Toast[]>([]);

function addToast(type: ToastType, message: string, duration = 4000) {
	const id = nextId++;
	toasts.update((all) => [...all, { id, type, message }]);
	setTimeout(() => {
		toasts.update((all) => all.filter((t) => t.id !== id));
	}, duration);
}

export function toastSuccess(message: string) {
	addToast('success', message);
}

export function toastError(message: string) {
	addToast('error', message, 6000);
}

export function toastInfo(message: string) {
	addToast('info', message);
}

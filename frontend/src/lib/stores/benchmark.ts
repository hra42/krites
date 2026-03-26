import { writable } from 'svelte/store';
import type { Suite, SuiteSummary, CreateSuiteRequest, UpdateSuiteRequest } from '$lib/types';
import * as api from '$lib/api/client';
import { toastSuccess, toastError } from '$lib/stores/toast';

export const suites = writable<SuiteSummary[]>([]);
export const currentSuite = writable<Suite | null>(null);
export const loading = writable(false);
export const error = writable<string | null>(null);

export async function loadSuites() {
	loading.set(true);
	error.set(null);
	try {
		const data = await api.listSuites();
		suites.set(data);
	} catch (e) {
		error.set(e instanceof Error ? e.message : 'Unknown error');
	} finally {
		loading.set(false);
	}
}

export async function loadSuite(id: string) {
	loading.set(true);
	error.set(null);
	try {
		const data = await api.getSuite(id);
		currentSuite.set(data);
	} catch (e) {
		error.set(e instanceof Error ? e.message : 'Unknown error');
	} finally {
		loading.set(false);
	}
}

export async function createNewSuite(data: CreateSuiteRequest): Promise<Suite | null> {
	loading.set(true);
	error.set(null);
	try {
		const suite = await api.createSuite(data);
		toastSuccess('Suite created');
		await loadSuites();
		return suite;
	} catch (e) {
		const msg = e instanceof Error ? e.message : 'Unknown error';
		error.set(msg);
		toastError(msg);
		return null;
	} finally {
		loading.set(false);
	}
}

export async function editSuite(id: string, data: UpdateSuiteRequest): Promise<Suite | null> {
	loading.set(true);
	error.set(null);
	try {
		const suite = await api.updateSuite(id, data);
		currentSuite.set(suite);
		toastSuccess('Suite updated');
		return suite;
	} catch (e) {
		const msg = e instanceof Error ? e.message : 'Unknown error';
		error.set(msg);
		toastError(msg);
		return null;
	} finally {
		loading.set(false);
	}
}

export async function removeSuite(id: string) {
	error.set(null);
	try {
		await api.deleteSuite(id);
		toastSuccess('Suite deleted');
		await loadSuites();
	} catch (e) {
		const msg = e instanceof Error ? e.message : 'Unknown error';
		error.set(msg);
		toastError(msg);
	}
}

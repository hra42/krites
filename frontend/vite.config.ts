import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			'/benchmarks': 'http://localhost:8080',
			'/v1': 'http://localhost:8080'
		}
	}
});

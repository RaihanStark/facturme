import { browser } from '$app/environment';
import { goto } from '$app/navigation';

export const ssr = false;

export function load() {
	if (browser) {
		const token = localStorage.getItem('authToken');

		if (token) {
			// Redirect to dashboard if already authenticated
			goto('/');
			return {};
		}
	}

	return {};
}

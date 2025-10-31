import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { api } from '$lib/api';
import { authStore } from '$lib/stores';

export const ssr = false; // Disable server-side rendering for client-only auth

export async function load() {
	if (browser) {
		const token = localStorage.getItem('authToken');

		if (!token) {
			// Redirect to login if not authenticated
			goto('/login');
			return {
				authenticated: false
			};
		}

		// Fetch fresh user data from the API
		try {
			const user = await api.getCurrentUser();
			authStore.setUser(user);
			return {
				authenticated: true,
				user
			};
		} catch (error) {
			// Token is invalid, clear and redirect
			authStore.clearAuth();
			goto('/login');
			return {
				authenticated: false
			};
		}
	}

	return {
		authenticated: false
	};
}

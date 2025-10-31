import { browser } from '$app/environment';
import { goto } from '$app/navigation';

export function load() {
	if (browser) {
		const token = localStorage.getItem('authToken');
		const userStr = localStorage.getItem('user');

		if (!token || !userStr) {
			// Not authenticated, redirect to login
			goto('/login');
			return;
		}

		const user = JSON.parse(userStr);

		// If already verified, redirect to dashboard
		if (user.email_verified) {
			goto('/');
			return;
		}
	}

	return {};
}

import { redirect } from '@sveltejs/kit';
import { api } from '$lib/api';
import type { PageLoad } from './$types';

export const ssr = false;

export const load: PageLoad = async () => {
	// Check if user is authenticated
	if (!api.isAuthenticated()) {
		throw redirect(302, '/login');
	}

	// Get user info from localStorage
	const user = api.getUser();

	// If user has already completed onboarding, redirect to dashboard
	if (user?.onboarding_completed) {
		throw redirect(302, '/');
	}

	return {};
};

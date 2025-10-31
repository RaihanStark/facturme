import { api } from '$lib/api';

export const ssr = false;

export function load({ depends }) {
	depends('clients');

	return {
		clients: api.getClients()
	};
}

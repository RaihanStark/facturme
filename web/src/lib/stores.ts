import { writable } from 'svelte/store';
import type { Client, TimeEntry, Invoice } from './types';
import type { UserInfo } from './api';
import { browser } from '$app/environment';

// Mock data for development
const mockClients: Client[] = [
	{
		id: '1',
		name: 'Acme Corporation',
		email: 'contact@acme.com',
		hourlyRate: 100,
		createdAt: new Date('2025-01-01')
	},
	{
		id: '2',
		name: 'TechStart Inc',
		email: 'hello@techstart.com',
		hourlyRate: 125,
		createdAt: new Date('2025-01-15')
	}
];

const mockTimeEntries: TimeEntry[] = [
	{
		id: '1',
		clientId: '1',
		date: new Date('2025-10-22'),
		hours: 4,
		notes: 'Frontend development',
		createdAt: new Date('2025-10-22')
	},
	{
		id: '2',
		clientId: '1',
		date: new Date('2025-10-23'),
		hours: 6,
		notes: 'API integration',
		createdAt: new Date('2025-10-23')
	},
	{
		id: '3',
		clientId: '2',
		date: new Date('2025-10-23'),
		hours: 3,
		notes: 'Consultation meeting',
		createdAt: new Date('2025-10-23')
	}
];

const mockInvoices: Invoice[] = [
	{
		id: '1',
		clientId: '1',
		invoiceNumber: 'INV-001',
		issueDate: new Date('2025-10-01'),
		dueDate: new Date('2025-10-31'),
		status: 'sent',
		timeEntries: [
			{
				id: '1',
				clientId: '1',
				date: new Date('2025-10-22'),
				hours: 4,
				notes: 'Frontend development',
				createdAt: new Date('2025-10-22')
			},
			{
				id: '2',
				clientId: '1',
				date: new Date('2025-10-23'),
				hours: 6,
				notes: 'API integration',
				createdAt: new Date('2025-10-23')
			}
		],
		totalHours: 10,
		totalAmount: 1000
	}
];

export const clients = writable<Client[]>(mockClients);
export const timeEntries = writable<TimeEntry[]>(mockTimeEntries);
export const invoices = writable<Invoice[]>(mockInvoices);

// Authentication store
function createAuthStore() {
	const { subscribe, set, update } = writable<{
		user: UserInfo | null;
		token: string | null;
		isAuthenticated: boolean;
	}>({
		user: null,
		token: null,
		isAuthenticated: false
	});

	// Initialize from localStorage on client side
	// Note: We only load the token here. User data will be fetched from API
	if (browser) {
		const token = localStorage.getItem('authToken');

		if (token) {
			set({
				user: null, // User data will be loaded by the layout
				token,
				isAuthenticated: true
			});
		}
	}

	return {
		subscribe,
		setAuth: (user: UserInfo, token: string) => {
			if (browser) {
				// Only store token, not user data
				localStorage.setItem('authToken', token);
			}
			set({
				user,
				token,
				isAuthenticated: true
			});
		},
		setUser: (user: UserInfo) => {
			// Update user data in store only, no localStorage
			update((state) => ({
				...state,
				user
			}));
		},
		clearAuth: () => {
			if (browser) {
				// Only remove token (user was never stored)
				localStorage.removeItem('authToken');
			}
			set({
				user: null,
				token: null,
				isAuthenticated: false
			});
		},
		checkAuth: () => {
			if (browser) {
				const token = localStorage.getItem('authToken');
				return !!token;
			}
			return false;
		}
	};
}

export const authStore = createAuthStore();

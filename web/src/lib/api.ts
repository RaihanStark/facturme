// API client configuration and utilities
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export interface ApiError {
	error: string;
}

export interface LoginRequest {
	email: string;
	password: string;
}

export interface RegisterRequest {
	email: string;
	password: string;
	name: string;
}

export interface VerifyEmailRequest {
	token: string;
}

export interface UserInfo {
	id: number;
	email: string;
	name: string;
	email_verified: boolean;
	onboarding_completed: boolean;
	tour_completed: boolean;
	currency: string;
}

export interface AuthResponse {
	token: string;
	user: UserInfo;
}

export interface Client {
	id: number;
	user_id: number;
	name: string;
	email: string;
	phone?: string;
	company?: string;
	address?: string;
	hourly_rate: number;
	currency: string;
	created_at: string;
	updated_at: string;
}

export interface CreateClientRequest {
	name: string;
	email: string;
	phone?: string;
	company?: string;
	address?: string;
	hourly_rate: number;
	currency: string;
}

export interface UpdateClientRequest {
	name: string;
	email: string;
	phone?: string;
	company?: string;
	address?: string;
	hourly_rate: number;
	currency: string;
}

export interface SupportedCurrency {
	code: string;
	symbol: string;
	name: string;
}

export interface TimeEntry {
	id: number;
	user_id: number;
	client_id: number;
	client_name?: string;
	client_currency?: string;
	date: string;
	hours: number;
	description?: string;
	hourly_rate: number;
	created_at: string;
	updated_at: string;
}

export interface CreateTimeEntryRequest {
	client_id: number;
	date: string;
	hours: number;
	description?: string;
}

export interface UpdateTimeEntryRequest {
	client_id: number;
	date: string;
	hours: number;
	description?: string;
}

export interface HeatmapResponse {
	start_date: string;
	end_date: string;
	data: Record<string, number>;
	entries: Record<string, TimeEntry[]>;
	total_hours: number;
	days_worked: number;
	days_off: number;
	average_hours: number;
}

export interface Invoice {
	id: number;
	user_id: number;
	client_id: number;
	client_name?: string;
	client_currency?: string;
	invoice_number: string;
	issue_date: string;
	due_date: string;
	status: 'draft' | 'sent' | 'paid' | 'overdue';
	notes?: string;
	time_entries: TimeEntry[];
	total_hours: number;
	total_amount: number;
	created_at: string;
	updated_at: string;
}

export interface CreateInvoiceRequest {
	client_id: number;
	invoice_number: string;
	issue_date: string;
	due_date: string;
	status: 'draft' | 'sent' | 'paid' | 'overdue';
	notes?: string;
	time_entry_ids: number[];
}

export interface UpdateInvoiceRequest {
	client_id: number;
	invoice_number: string;
	issue_date: string;
	due_date: string;
	status: 'draft' | 'sent' | 'paid' | 'overdue';
	notes?: string;
}

export interface UpdateInvoiceStatusRequest {
	status: 'draft' | 'sent' | 'paid' | 'overdue';
}

export interface CompleteOnboardingRequest {
	currency: string;
}

export interface DashboardStats {
	total_hours: number;
	total_revenue: number;
	unpaid_invoices: number;
	paid_invoices: number;
}

export interface TimeEntriesStats {
	entries: TimeEntry[];
	total_hours: number;
	total_revenue: number;
}

export interface InvoiceStats {
	invoices: Invoice[];
	total_invoices: number;
	total_amount: number;
	paid_amount: number;
	unpaid_amount: number;
}

class ApiClient {
	private baseUrl: string;

	constructor(baseUrl: string) {
		this.baseUrl = baseUrl;
	}

	private async request<T>(
		endpoint: string,
		options: RequestInit = {}
	): Promise<T> {
		const token = localStorage.getItem('authToken');

		const config: RequestInit = {
			...options,
			headers: {
				'Content-Type': 'application/json',
				...(token && { Authorization: `Bearer ${token}` }),
				...options.headers
			}
		};

		const response = await fetch(`${this.baseUrl}${endpoint}`, config);

		if (!response.ok) {
			const error: ApiError = await response.json();
			throw new Error(error.error || 'An error occurred');
		}

		return response.json();
	}

	// Authentication endpoints
	async login(credentials: LoginRequest): Promise<AuthResponse> {
		const response = await this.request<AuthResponse>('/api/auth/login', {
			method: 'POST',
			body: JSON.stringify(credentials)
		});

		// Store token in localStorage (user data will be loaded from API)
		localStorage.setItem('authToken', response.token);

		return response;
	}

	async register(data: RegisterRequest): Promise<AuthResponse> {
		const response = await this.request<AuthResponse>('/api/auth/register', {
			method: 'POST',
			body: JSON.stringify(data)
		});

		// Store token in localStorage (user data will be loaded from API)
		localStorage.setItem('authToken', response.token);

		return response;
	}

	async verifyEmail(data: VerifyEmailRequest): Promise<UserInfo> {
		const response = await this.request<UserInfo>('/api/auth/verify-email', {
			method: 'POST',
			body: JSON.stringify(data)
		});

		// User data will be managed by authStore
		return response;
	}

	async resendVerificationEmail(): Promise<{ message: string }> {
		return this.request<{ message: string }>('/api/auth/resend-verification', {
			method: 'POST'
		});
	}

	async forgotPassword(email: string): Promise<{ message: string }> {
		return this.request<{ message: string }>('/api/auth/forgot-password', {
			method: 'POST',
			body: JSON.stringify({ email })
		});
	}

	async resetPassword(token: string, password: string): Promise<{ message: string }> {
		return this.request<{ message: string }>('/api/auth/reset-password', {
			method: 'POST',
			body: JSON.stringify({ token, password })
		});
	}

	async changePassword(currentPassword: string, newPassword: string): Promise<{ message: string }> {
		return this.request<{ message: string }>('/api/users/change-password', {
			method: 'POST',
			body: JSON.stringify({ current_password: currentPassword, new_password: newPassword })
		});
	}

	async updateCurrency(currency: string): Promise<UserInfo> {
		const response = await this.request<UserInfo>('/api/users/currency', {
			method: 'POST',
			body: JSON.stringify({ currency })
		});

		// User data will be managed by authStore
		return response;
	}

	async getCurrentUser(): Promise<UserInfo> {
		return this.request<UserInfo>('/api/users/me', {
			method: 'GET'
		});
	}

	logout() {
		localStorage.removeItem('authToken');
		localStorage.removeItem('user');
	}

	async completeOnboarding(data: CompleteOnboardingRequest): Promise<UserInfo> {
		const response = await this.request<UserInfo>('/api/users/complete-onboarding', {
			method: 'POST',
			body: JSON.stringify(data)
		});

		// User data will be managed by authStore
		return response;
	}

	async completeTour(): Promise<UserInfo> {
		const response = await this.request<UserInfo>('/api/users/complete-tour', {
			method: 'POST'
		});

		// User data will be managed by authStore
		return response;
	}

	getToken(): string | null {
		return localStorage.getItem('authToken');
	}

	getUser(): UserInfo | null {
		const user = localStorage.getItem('user');
		return user ? JSON.parse(user) : null;
	}

	isAuthenticated(): boolean {
		return !!this.getToken();
	}

	// Client endpoints
	async getClients(): Promise<Client[]> {
		return this.request<Client[]>('/api/clients', {
			method: 'GET'
		});
	}

	async getClient(id: number): Promise<Client> {
		return this.request<Client>(`/api/clients/${id}`, {
			method: 'GET'
		});
	}

	async createClient(data: CreateClientRequest): Promise<Client> {
		return this.request<Client>('/api/clients', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async updateClient(id: number, data: UpdateClientRequest): Promise<Client> {
		return this.request<Client>(`/api/clients/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	async deleteClient(id: number): Promise<void> {
		await fetch(`${this.baseUrl}/api/clients/${id}`, {
			method: 'DELETE',
			headers: {
				Authorization: `Bearer ${this.getToken()}`
			}
		});
	}

	async getSupportedCurrencies(): Promise<SupportedCurrency[]> {
		return this.request<SupportedCurrency[]>('/api/supported-currencies', {
			method: 'GET'
		});
	}

	async convertCurrency(amount: number, from: string, to: string): Promise<number> {
		const response = await this.request<{ converted_amount: number }>(
			`/api/convert-currency?amount=${amount}&from=${from}&to=${to}`,
			{
				method: 'GET'
			}
		);
		return response.converted_amount;
	}

	// Time entry endpoints
	async getTimeEntries(): Promise<TimeEntry[]> {
		return this.request<TimeEntry[]>('/api/time-entries', {
			method: 'GET'
		});
	}

	async getTimeEntry(id: number): Promise<TimeEntry> {
		return this.request<TimeEntry>(`/api/time-entries/${id}`, {
			method: 'GET'
		});
	}

	async createTimeEntry(data: CreateTimeEntryRequest): Promise<TimeEntry> {
		return this.request<TimeEntry>('/api/time-entries', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async updateTimeEntry(id: number, data: UpdateTimeEntryRequest): Promise<TimeEntry> {
		return this.request<TimeEntry>(`/api/time-entries/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	async deleteTimeEntry(id: number): Promise<void> {
		await fetch(`${this.baseUrl}/api/time-entries/${id}`, {
			method: 'DELETE',
			headers: {
				Authorization: `Bearer ${this.getToken()}`
			}
		});
	}

	async getHeatmap(startDate: string, endDate: string): Promise<HeatmapResponse> {
		return this.request<HeatmapResponse>(
			`/api/time-entries/heatmap?start_date=${startDate}&end_date=${endDate}`,
			{
				method: 'GET'
			}
		);
	}

	// Invoice endpoints
	async getInvoices(): Promise<Invoice[]> {
		return this.request<Invoice[]>('/api/invoices', {
			method: 'GET'
		});
	}

	async getInvoice(id: number): Promise<Invoice> {
		return this.request<Invoice>(`/api/invoices/${id}`, {
			method: 'GET'
		});
	}

	async createInvoice(data: CreateInvoiceRequest): Promise<Invoice> {
		return this.request<Invoice>('/api/invoices', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async updateInvoice(id: number, data: UpdateInvoiceRequest): Promise<Invoice> {
		return this.request<Invoice>(`/api/invoices/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	async updateInvoiceStatus(id: number, data: UpdateInvoiceStatusRequest): Promise<Invoice> {
		return this.request<Invoice>(`/api/invoices/${id}/status`, {
			method: 'PATCH',
			body: JSON.stringify(data)
		});
	}

	async deleteInvoice(id: number): Promise<void> {
		await fetch(`${this.baseUrl}/api/invoices/${id}`, {
			method: 'DELETE',
			headers: {
				Authorization: `Bearer ${this.getToken()}`
			}
		});
	}

	async getAvailableTimeEntries(clientId: number): Promise<TimeEntry[]> {
		return this.request<TimeEntry[]>(`/api/invoices/available-time-entries?client_id=${clientId}`, {
			method: 'GET'
		});
	}

	async downloadInvoicePDF(id: number): Promise<void> {
		const token = this.getToken();
		const response = await fetch(`${this.baseUrl}/api/invoices/${id}/pdf`, {
			method: 'GET',
			headers: {
				...(token && { Authorization: `Bearer ${token}` })
			}
		});

		if (!response.ok) {
			throw new Error('Failed to download PDF');
		}

		// Get filename from Content-Disposition header or use default
		const contentDisposition = response.headers.get('Content-Disposition');
		let filename = `invoice-${id}.pdf`;
		if (contentDisposition) {
			const filenameMatch = contentDisposition.match(/filename="?(.+)"?/);
			if (filenameMatch) {
				filename = filenameMatch[1];
			}
		}

		// Create blob and trigger download
		const blob = await response.blob();
		const url = window.URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = filename;
		document.body.appendChild(a);
		a.click();
		window.URL.revokeObjectURL(url);
		document.body.removeChild(a);
	}

	// Demo endpoints
	async generateDemoData(): Promise<{ message: string }> {
		return this.request<{ message: string }>('/api/demo/generate', {
			method: 'POST'
		});
	}

	async deleteDemoData(): Promise<{ message: string }> {
		return this.request<{ message: string }>('/api/demo', {
			method: 'DELETE'
		});
	}

	// Stats endpoints
	async getDashboardStats(from?: string, to?: string): Promise<DashboardStats> {
		const params = new URLSearchParams();
		if (from) params.append('from', from);
		if (to) params.append('to', to);
		const query = params.toString() ? `?${params}` : '';
		return this.request<DashboardStats>(`/api/stats/dashboard${query}`, {
			method: 'GET'
		});
	}

	async getRecentTimeEntries(from?: string, to?: string, limit?: number): Promise<TimeEntry[]> {
		const params = new URLSearchParams();
		if (from) params.append('from', from);
		if (to) params.append('to', to);
		if (limit) params.append('limit', limit.toString());
		const query = params.toString() ? `?${params}` : '';
		return this.request<TimeEntry[]>(`/api/stats/recent-time-entries${query}`, {
			method: 'GET'
		});
	}

	async getRecentInvoices(from?: string, to?: string, limit?: number): Promise<Invoice[]> {
		const params = new URLSearchParams();
		if (from) params.append('from', from);
		if (to) params.append('to', to);
		if (limit) params.append('limit', limit.toString());
		const query = params.toString() ? `?${params}` : '';
		return this.request<Invoice[]>(`/api/stats/recent-invoices${query}`, {
			method: 'GET'
		});
	}

	async getTimeEntriesStats(
		viewMode: 'daily' | 'weekly' | 'monthly',
		date: string
	): Promise<TimeEntriesStats> {
		const params = new URLSearchParams();
		params.append('view_mode', viewMode);
		params.append('date', date);
		return this.request<TimeEntriesStats>(`/api/time-entries/stats?${params}`, {
			method: 'GET'
		});
	}

	async getInvoiceStats(status: 'all' | 'draft' | 'sent' | 'paid' | 'overdue' = 'all'): Promise<InvoiceStats> {
		const params = new URLSearchParams();
		if (status !== 'all') {
			params.append('status', status);
		}
		const query = params.toString() ? `?${params}` : '';
		return this.request<InvoiceStats>(`/api/stats/invoices${query}`, {
			method: 'GET'
		});
	}
}

// Export singleton instance
export const api = new ApiClient(API_BASE_URL);

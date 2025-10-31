export interface Client {
	id: string;
	name: string;
	email: string;
	hourlyRate: number;
	createdAt: Date;
}

export interface TimeEntry {
	id: string;
	clientId: string;
	date: Date;
	hours: number;
	notes: string;
	createdAt: Date;
}

export interface Invoice {
	id: string;
	clientId: string;
	invoiceNumber: string;
	issueDate: Date;
	dueDate: Date;
	status: 'draft' | 'sent' | 'paid' | 'overdue';
	timeEntries: TimeEntry[];
	totalHours: number;
	totalAmount: number;
	notes?: string;
}

export type InvoiceStatus = Invoice['status'];

export interface DashboardStats {
	totalHours: number;
	totalRevenue: number;
	activeClients: number;
	pendingInvoices: number;
}

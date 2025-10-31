<script lang="ts">
	import {
		Button,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell,
		Modal,
		Label,
		Input,
		Textarea,
		Select,
		Toast,
		Badge,
		ButtonGroup,
		Checkbox
	} from 'flowbite-svelte';
	import {
		PlusOutline,
		EyeOutline,
		TrashBinOutline,
		CheckCircleSolid,
		DownloadOutline,
		FileInvoiceSolid,
		CalendarMonthOutline,
		DollarOutline
	} from 'flowbite-svelte-icons';
	import {
		api,
		type Invoice,
		type TimeEntry,
		type Client,
		type CreateInvoiceRequest,
		type InvoiceStats
	} from '$lib/api';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { formatCurrency, formatCurrencyRate } from '$lib/utils/currency';
	import { authStore } from '$lib/stores';

	let showCreateModal = $state(false);
	let showViewModal = $state(false);
	let selectedInvoice: Invoice | null = $state(null);
	let showToast = $state(false);
	let toastMessage = $state('');
	let filterStatus: 'all' | 'draft' | 'sent' | 'paid' | 'overdue' = $state('all');
	let isSaving = $state(false);
	let availableTimeEntries: TimeEntry[] = $state([]);
	let selectedTimeEntryIds: Set<number> = $state(new Set());
	let isLoading = $state(true);
	let statsData: InvoiceStats | null = $state(null);
	let clients: Client[] = $state([]);

	let formData = $state({
		clientId: 0,
		invoiceNumber: '',
		issueDate: new Date().toISOString().split('T')[0],
		dueDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
		notes: ''
	});

	async function loadStats() {
		isLoading = true;
		try {
			statsData = await api.getInvoiceStats(filterStatus);
		} catch (error) {
			console.error('Failed to load invoice stats:', error);
		} finally {
			isLoading = false;
		}
	}

	async function loadClients() {
		if (clients.length > 0) return;
		try {
			clients = await api.getClients();
		} catch (error) {
			console.error('Failed to load clients:', error);
		}
	}

	$effect(() => {
		filterStatus;
		loadStats();
	});

	const clientOptions = $derived(clients.map((c) => ({ value: c.id, name: c.name })));

	async function openCreateModal() {
		await loadClients();

		// Generate suggested invoice number
		const totalInvoices = statsData?.total_invoices || 0;
		const suggestedNumber = `INV-${String(totalInvoices + 1).padStart(4, '0')}`;

		formData = {
			clientId: clients[0]?.id || 0,
			invoiceNumber: suggestedNumber,
			issueDate: new Date().toISOString().split('T')[0],
			dueDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
			notes: ''
		};
		selectedTimeEntryIds = new Set();

		if (formData.clientId) {
			await loadAvailableTimeEntries(formData.clientId);
		}

		showCreateModal = true;
	}

	async function loadAvailableTimeEntries(clientId: number) {
		try {
			availableTimeEntries = await api.getAvailableTimeEntries(clientId);
		} catch (error) {
			availableTimeEntries = [];
			console.error('Failed to load available time entries:', error);
		}
	}

	async function handleClientChange(clientId: number) {
		selectedTimeEntryIds = new Set();
		await loadAvailableTimeEntries(clientId);
	}

	function toggleTimeEntry(entryId: number) {
		const newSet = new Set(selectedTimeEntryIds);
		if (newSet.has(entryId)) {
			newSet.delete(entryId);
		} else {
			newSet.add(entryId);
		}
		selectedTimeEntryIds = newSet;
	}

	function toggleAllTimeEntries() {
		if (selectedTimeEntryIds.size === availableTimeEntries.length) {
			// Deselect all
			selectedTimeEntryIds = new Set();
		} else {
			// Select all
			selectedTimeEntryIds = new Set(availableTimeEntries.map((e) => e.id));
		}
	}

	function areAllSelected() {
		return (
			availableTimeEntries.length > 0 && selectedTimeEntryIds.size === availableTimeEntries.length
		);
	}

	async function generateInvoice() {
		if (selectedTimeEntryIds.size === 0) {
			showMessage('Please select at least one time entry');
			return;
		}

		if (!formData.invoiceNumber.trim()) {
			showMessage('Please enter an invoice number');
			return;
		}

		isSaving = true;
		try {
			await api.createInvoice({
				client_id: formData.clientId,
				invoice_number: formData.invoiceNumber.trim(),
				issue_date: formData.issueDate,
				due_date: formData.dueDate,
				status: 'draft',
				notes: formData.notes || undefined,
				time_entry_ids: Array.from(selectedTimeEntryIds)
			});

			showCreateModal = false;
			showMessage('Invoice created successfully');
			await loadStats();
		} catch (error) {
			showMessage(error instanceof Error ? error.message : 'Failed to create invoice');
		} finally {
			isSaving = false;
		}
	}

	async function viewInvoice(invoice: Invoice) {
		await loadClients();
		selectedInvoice = invoice;
		showViewModal = true;
	}

	async function updateInvoiceStatus(
		invoiceId: number,
		status: 'draft' | 'sent' | 'paid' | 'overdue'
	) {
		try {
			await api.updateInvoiceStatus(invoiceId, { status });
			showMessage(`Invoice marked as ${status}`);
			await loadStats();
		} catch (error) {
			showMessage('Failed to update invoice status');
		}
	}

	async function deleteInvoice(invoiceId: number) {
		if (!confirm('Are you sure you want to delete this invoice?')) return;

		try {
			await api.deleteInvoice(invoiceId);
			showMessage('Invoice deleted successfully');
			await loadStats();
		} catch (error) {
			showMessage('Failed to delete invoice');
		}
	}

	async function downloadPDF(invoiceId: number) {
		try {
			await api.downloadInvoicePDF(invoiceId);
			showMessage('PDF downloaded successfully');
		} catch (error) {
			showMessage('Failed to download PDF');
		}
	}

	const getClient = (clientId: number) => clients.find((c) => c.id === clientId);

	const formatDate = (date: string) =>
		new Date(date).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});

	const statusColors: Record<
		'draft' | 'sent' | 'paid' | 'overdue',
		'gray' | 'blue' | 'green' | 'red'
	> = {
		draft: 'gray',
		sent: 'blue',
		paid: 'green',
		overdue: 'red'
	};

	function showMessage(message: string) {
		toastMessage = message;
		showToast = true;
		setTimeout(() => (showToast = false), 3000);
	}
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-3xl font-bold text-white">Invoices</h1>
			<p class="mt-2 text-slate-400">Create and manage client invoices</p>
		</div>
		<Button onclick={() => openCreateModal()} color="primary">
			<PlusOutline class="w-4 h-4 mr-2" />
			Create Invoice
		</Button>
	</div>

	<!-- Stats Grid -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
		<div class="bg-slate-800 rounded-xl border border-slate-700 p-6">
			<div class="flex items-center justify-between">
				<div class="flex-1">
					<p class="text-sm font-medium text-slate-400">Total Invoices</p>
					{#if isLoading}
						<div class="h-9 bg-slate-700 rounded w-12 mt-2 animate-pulse"></div>
					{:else}
						<p class="text-3xl font-bold text-white mt-2">{statsData?.total_invoices || 0}</p>
					{/if}
				</div>
				<div class="p-3 bg-blue-500/10 rounded-lg">
					<FileInvoiceSolid class="w-8 h-8 text-blue-400" />
				</div>
			</div>
		</div>

		<div class="bg-slate-800 rounded-xl border border-slate-700 p-6">
			<div class="flex items-center justify-between">
				<div class="flex-1">
					<p class="text-sm font-medium text-slate-400">Total Amount</p>
					{#if isLoading}
						<div class="h-9 bg-slate-700 rounded w-32 mt-2 animate-pulse"></div>
					{:else}
						<p class="text-3xl font-bold text-white mt-2">
							{formatCurrency(statsData?.total_amount || 0, $authStore.user?.currency)}
						</p>
					{/if}
				</div>
				<div class="p-3 bg-violet-500/10 rounded-lg">
					<DollarOutline class="w-8 h-8 text-violet-400" />
				</div>
			</div>
		</div>

		<div class="bg-slate-800 rounded-xl border border-slate-700 p-6">
			<div class="flex items-center justify-between">
				<div class="flex-1">
					<p class="text-sm font-medium text-slate-400">Paid</p>
					{#if isLoading}
						<div class="h-9 bg-slate-700 rounded w-28 mt-2 animate-pulse"></div>
					{:else}
						<p class="text-3xl font-bold text-white mt-2">
							{formatCurrency(statsData?.paid_amount || 0, $authStore.user?.currency)}
						</p>
					{/if}
				</div>
				<div class="p-3 bg-emerald-500/10 rounded-lg">
					<svg
						class="w-8 h-8 text-emerald-400"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
				</div>
			</div>
		</div>

		<div class="bg-slate-800 rounded-xl border border-slate-700 p-6">
			<div class="flex items-center justify-between">
				<div class="flex-1">
					<p class="text-sm font-medium text-slate-400">Unpaid</p>
					{#if isLoading}
						<div class="h-9 bg-slate-700 rounded w-28 mt-2 animate-pulse"></div>
					{:else}
						<p class="text-3xl font-bold text-white mt-2">
							{formatCurrency(statsData?.unpaid_amount || 0, $authStore.user?.currency)}
						</p>
					{/if}
				</div>
				<div class="p-3 bg-amber-500/10 rounded-lg">
					<svg class="w-8 h-8 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
				</div>
			</div>
		</div>
	</div>

	<!-- Filter Buttons -->
	<div class="bg-slate-800 rounded-xl border border-slate-700 p-4">
		<ButtonGroup>
			<Button
				size="sm"
				color={filterStatus === 'all' ? 'primary' : 'light'}
				onclick={() => (filterStatus = 'all')}
			>
				All
			</Button>
			<Button
				size="sm"
				color={filterStatus === 'draft' ? 'primary' : 'light'}
				onclick={() => (filterStatus = 'draft')}
			>
				Draft
			</Button>
			<Button
				size="sm"
				color={filterStatus === 'sent' ? 'primary' : 'light'}
				onclick={() => (filterStatus = 'sent')}
			>
				Sent
			</Button>
			<Button
				size="sm"
				color={filterStatus === 'paid' ? 'primary' : 'light'}
				onclick={() => (filterStatus = 'paid')}
			>
				Paid
			</Button>
			<Button
				size="sm"
				color={filterStatus === 'overdue' ? 'primary' : 'light'}
				onclick={() => (filterStatus = 'overdue')}
			>
				Overdue
			</Button>
		</ButtonGroup>
	</div>

	<!-- Invoices List -->
	{#if isLoading}
		<!-- Skeleton -->
		<div class="grid grid-cols-1 gap-4">
			{#each Array(5) as _}
				<div class="bg-slate-800 rounded-xl border border-slate-700 p-6 animate-pulse">
					<div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
						<!-- Invoice Info Skeleton -->
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-3 mb-2">
								<div class="h-7 bg-slate-700 rounded w-28"></div>
								<div class="h-5 bg-slate-700 rounded w-16"></div>
							</div>
							<div class="flex flex-wrap items-center gap-4 mb-3">
								<div class="h-4 bg-slate-700/60 rounded w-32"></div>
								<div class="h-4 bg-slate-700/60 rounded w-36"></div>
								<div class="h-4 bg-slate-700/60 rounded w-28"></div>
							</div>
							<div class="flex items-center gap-3">
								<div class="h-4 bg-slate-700/60 rounded w-12"></div>
								<div class="w-1 h-1 bg-slate-700 rounded-full"></div>
								<div class="h-8 bg-slate-700 rounded w-24"></div>
							</div>
						</div>

						<!-- Actions Skeleton -->
						<div class="flex items-center gap-2">
							<div class="h-9 w-20 bg-slate-700 rounded"></div>
							<div class="h-9 w-28 bg-slate-700 rounded"></div>
							<div class="h-9 w-9 bg-slate-700 rounded"></div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{:else if !statsData || statsData.invoices.length === 0}
		<div class="bg-slate-800 rounded-xl border border-slate-700">
			<EmptyState title="No invoices yet" description="Create your first invoice to get started">
				{#snippet icon()}
					<FileInvoiceSolid class="w-8 h-8 text-slate-500" />
				{/snippet}
			</EmptyState>
		</div>
	{:else}
		{@const sortedInvoices = [...statsData.invoices].sort(
			(a, b) => new Date(b.issue_date).getTime() - new Date(a.issue_date).getTime()
		)}
		<div class="grid grid-cols-1 gap-4">
			{#each sortedInvoices as invoice}
				<div
					class="bg-slate-800 rounded-xl border border-slate-700 hover:border-slate-600 transition-all duration-200 group"
				>
					<div class="p-6">
						<div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
							<!-- Invoice Info -->
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-3 mb-2">
									<h3 class="font-bold text-white text-xl">{invoice.invoice_number}</h3>
									<Badge color={statusColors[invoice.status]} class="capitalize !text-xs">
										{invoice.status}
									</Badge>
								</div>
								<div class="flex flex-wrap items-center gap-4 text-sm text-slate-400">
									<div class="flex items-center gap-1.5">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
											/>
										</svg>
										<span class="text-slate-300">{invoice.client_name || 'Unknown'}</span>
									</div>
									<div class="flex items-center gap-1.5">
										<CalendarMonthOutline class="w-4 h-4" />
										<span>Issued: {formatDate(invoice.issue_date)}</span>
									</div>
									<div class="flex items-center gap-1.5">
										<CalendarMonthOutline class="w-4 h-4" />
										<span class={invoice.status === 'overdue' ? 'text-red-400 font-medium' : ''}>
											Due: {formatDate(invoice.due_date)}
										</span>
									</div>
								</div>
								<div class="mt-3 flex items-center gap-3">
									<span class="text-sm text-slate-400">{invoice.total_hours}h</span>
									<span class="text-slate-600">•</span>
									<span class="text-2xl font-bold text-white"
										>{formatCurrency(invoice.total_amount, invoice.client_currency || 'USD')}</span
									>
								</div>
							</div>

							<!-- Actions -->
							<div
								class="flex items-center gap-2 lg:opacity-0 lg:group-hover:opacity-100 transition-opacity"
							>
								<Button size="sm" color="light" onclick={() => viewInvoice(invoice)}>
									<EyeOutline class="w-4 h-4 mr-2" />
									View
								</Button>
								{#if invoice.status === 'draft'}
									<Button
										size="sm"
										color="blue"
										onclick={() => updateInvoiceStatus(invoice.id, 'sent')}
									>
										Set to Sent
									</Button>
								{:else if invoice.status === 'sent' || invoice.status === 'overdue'}
									<Button
										size="sm"
										color="green"
										onclick={() => updateInvoiceStatus(invoice.id, 'paid')}
									>
										Mark as Paid
									</Button>
								{:else if invoice.status === 'paid'}
									<Button
										size="sm"
										color="yellow"
										onclick={() => updateInvoiceStatus(invoice.id, 'sent')}
									>
										Mark as Unpaid
									</Button>
								{/if}
								<Button size="sm" color="red" onclick={() => deleteInvoice(invoice.id)}>
									<TrashBinOutline class="w-4 h-4" />
								</Button>
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Invoice Modal -->
<Modal bind:open={showCreateModal} size="lg" autoclose={false} class="dark:bg-slate-800">
	<form
		class="space-y-6"
		onsubmit={(e) => {
			e.preventDefault();
			generateInvoice();
		}}
	>
		<!-- Header with Icon -->
		<div class="flex items-center gap-4 pb-4 border-b border-slate-700">
			<div class="p-3 bg-blue-500/10 rounded-xl">
				<FileInvoiceSolid class="w-7 h-7 text-blue-400" />
			</div>
			<div>
				<h3 class="text-2xl font-bold text-white">Create Invoice</h3>
				<p class="text-sm text-slate-400 mt-1">Generate an invoice from your time entries</p>
			</div>
		</div>

		<!-- Invoice Number -->
		<div class="space-y-2">
			<Label for="invoiceNumber" class="text-base font-semibold text-white flex items-center gap-2">
				<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14"
					/>
				</svg>
				Invoice Number
			</Label>
			<Input
				id="invoiceNumber"
				type="text"
				bind:value={formData.invoiceNumber}
				placeholder="INV-0001"
				required
				class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
			/>
		</div>

		<!-- Client Selection -->
		<div class="space-y-2">
			<Label for="client" class="text-base font-semibold text-white flex items-center gap-2">
				<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
					/>
				</svg>
				Client
			</Label>
			<Select
				id="client"
				bind:value={formData.clientId}
				onchange={() => handleClientChange(formData.clientId)}
				required
				items={clientOptions}
				class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
			/>
		</div>

		<!-- Date Fields -->
		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<div class="space-y-2">
				<Label for="issueDate" class="text-base font-semibold text-white flex items-center gap-2">
					<CalendarMonthOutline class="w-4 h-4 text-slate-400" />
					Issue Date
				</Label>
				<Input
					id="issueDate"
					type="date"
					bind:value={formData.issueDate}
					required
					class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
				/>
			</div>

			<div class="space-y-2">
				<Label for="dueDate" class="text-base font-semibold text-white flex items-center gap-2">
					<CalendarMonthOutline class="w-4 h-4 text-slate-400" />
					Due Date
				</Label>
				<Input
					id="dueDate"
					type="date"
					bind:value={formData.dueDate}
					required
					class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
				/>
			</div>
		</div>

		<!-- Notes -->
		<div class="space-y-2">
			<Label for="notes" class="text-base font-semibold text-white flex items-center gap-2">
				<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
					/>
				</svg>
				Notes
				<span class="text-xs text-slate-500 font-normal">(optional)</span>
			</Label>
			<Textarea
				id="notes"
				bind:value={formData.notes}
				placeholder="Add any additional notes for the invoice..."
				rows={3}
				class="!bg-slate-700 !border-slate-600 !text-white placeholder:!text-slate-500 focus:!ring-blue-500 focus:!border-blue-500 resize-none !w-full"
			/>
		</div>

		<!-- Select Work Logs -->
		<div class="space-y-3">
			<Label class="text-base font-semibold text-white flex items-center gap-2">
				<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
					/>
				</svg>
				Select Time Entries
			</Label>
			{#if availableTimeEntries.length === 0}
				<div
					class="bg-slate-700/50 border border-slate-600 rounded-lg p-6 text-center text-slate-400"
				>
					<svg
						class="w-12 h-12 mx-auto mb-3 text-slate-500"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<p class="font-medium">No available time entries for this client</p>
					<p class="text-sm mt-1">Log some hours first before creating an invoice</p>
				</div>
			{:else}
				<div class="border border-slate-600 rounded-lg overflow-hidden bg-slate-700/30">
					<Table hoverable={true} class="!bg-transparent">
						<TableHead class="!bg-slate-700/50 !border-slate-600">
							<TableHeadCell class="!p-4 !bg-transparent !text-slate-300">
								<Checkbox checked={areAllSelected()} onchange={toggleAllTimeEntries} />
							</TableHeadCell>
							<TableHeadCell class="!bg-transparent !text-slate-300">Date</TableHeadCell>
							<TableHeadCell class="!bg-transparent !text-slate-300">Description</TableHeadCell>
							<TableHeadCell class="!bg-transparent !text-slate-300">Hours</TableHeadCell>
						</TableHead>
						<TableBody class="!bg-transparent">
							{#each availableTimeEntries as entry}
								<TableBodyRow
									class="!bg-transparent !border-slate-600 hover:!bg-slate-600/30 cursor-pointer transition-colors"
									onclick={() => toggleTimeEntry(entry.id)}
								>
									<TableBodyCell class="!p-4 !bg-transparent !text-white">
										<Checkbox
											checked={selectedTimeEntryIds.has(entry.id)}
											onclick={(e) => e.stopPropagation()}
											onchange={() => toggleTimeEntry(entry.id)}
										/>
									</TableBodyCell>
									<TableBodyCell class="!bg-transparent !text-white"
										>{formatDate(entry.date)}</TableBodyCell
									>
									<TableBodyCell class="!bg-transparent !text-slate-300"
										>{entry.description || 'No description'}</TableBodyCell
									>
									<TableBodyCell class="!bg-transparent !text-white font-medium"
										>{entry.hours}h</TableBodyCell
									>
								</TableBodyRow>
							{/each}
						</TableBody>
					</Table>
				</div>
				<div class="flex items-center gap-2 text-sm">
					<div
						class="px-3 py-1.5 bg-blue-500/10 border border-blue-500/20 rounded-lg text-blue-300 font-medium"
					>
						{selectedTimeEntryIds.size} selected
					</div>
					<span class="text-slate-500">•</span>
					<span class="text-slate-400">{availableTimeEntries.length} total</span>
				</div>

				<!-- Invoice Preview -->
				{#if selectedTimeEntryIds.size > 0}
					{@const selectedEntries = availableTimeEntries.filter((e) =>
						selectedTimeEntryIds.has(e.id)
					)}
					{@const totalHours = selectedEntries.reduce((sum, e) => sum + e.hours, 0)}
					{@const selectedClient = clients.find((c) => c.id === formData.clientId)}
					{@const hourlyRate = selectedClient?.hourly_rate || 0}
					{@const totalAmount = totalHours * hourlyRate}

					<div
						class="bg-gradient-to-br from-blue-500/10 to-violet-500/10 border border-blue-500/20 rounded-xl p-5"
					>
						<div class="flex items-center gap-3 mb-4">
							<div class="p-2 bg-blue-500/20 rounded-lg">
								<svg
									class="w-5 h-5 text-blue-400"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"
									/>
								</svg>
							</div>
							<h4 class="text-base font-semibold text-white">Invoice Preview</h4>
						</div>

						<div class="grid grid-cols-2 gap-4">
							<div class="bg-slate-700/30 rounded-lg p-4">
								<div class="flex items-center gap-2 mb-2">
									<svg
										class="w-4 h-4 text-slate-400"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
										/>
									</svg>
									<span class="text-sm text-slate-400">Total Hours</span>
								</div>
								<p class="text-2xl font-bold text-white">{totalHours.toFixed(2)}h</p>
							</div>

							<div class="bg-slate-700/30 rounded-lg p-4">
								<div class="flex items-center gap-2 mb-2">
									<svg
										class="w-4 h-4 text-slate-400"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
										/>
									</svg>
									<span class="text-sm text-slate-400">Total Amount</span>
								</div>
								<p class="text-2xl font-bold text-emerald-400">
									{formatCurrency(totalAmount, selectedClient?.currency || 'USD')}
								</p>
							</div>
						</div>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Action Buttons -->
		<div class="flex gap-3 pt-4">
			<Button
				type="submit"
				color="blue"
				class="flex-1 !py-3 !text-base !font-semibold"
				disabled={isSaving || selectedTimeEntryIds.size === 0}
			>
				{#if isSaving}
					<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
						></circle>
						<path
							class="opacity-75"
							fill="currentColor"
							d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
						></path>
					</svg>
					Creating Invoice...
				{:else}
					<CheckCircleSolid class="w-4 h-4 mr-2" />
					Create Invoice
				{/if}
			</Button>
			<Button
				type="button"
				color="alternative"
				class="!py-3 px-6 !text-base !font-semibold"
				onclick={() => (showCreateModal = false)}
				disabled={isSaving}
			>
				Cancel
			</Button>
		</div>
	</form>
</Modal>

<!-- View Invoice Modal -->
<Modal bind:open={showViewModal} size="lg" autoclose={false} class="dark:bg-slate-800">
	{#if selectedInvoice}
		{@const client = getClient(selectedInvoice.client_id)}
		<div class="space-y-6">
			<!-- Header with Icon and Status -->
			<div class="pb-4 border-b border-slate-700">
				<div class="flex items-center gap-4 mb-2">
					<div class="p-3 bg-blue-500/10 rounded-xl">
						<FileInvoiceSolid class="w-7 h-7 text-blue-400" />
					</div>
					<div class="flex-1">
						<div class="flex items-center gap-3">
							<h3 class="text-2xl font-bold text-white">{selectedInvoice.invoice_number}</h3>
							<Badge
								color={statusColors[selectedInvoice.status]}
								class="capitalize !text-sm !px-3 !py-1.5"
							>
								{selectedInvoice.status}
							</Badge>
						</div>
						<p class="text-sm text-slate-400 mt-1">Invoice Details</p>
					</div>
				</div>
			</div>

			<!-- Bill To & Dates Section -->
			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<!-- Bill To -->
				<div class="bg-slate-700/30 rounded-xl p-5 border border-slate-600">
					<div class="flex items-center gap-2 mb-3">
						<svg
							class="w-5 h-5 text-blue-400"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
							/>
						</svg>
						<h4 class="text-sm font-semibold text-slate-400 uppercase tracking-wide">Bill To</h4>
					</div>
					<p class="text-lg font-bold text-white mb-1">{client?.name}</p>
					<p class="text-sm text-slate-400">{client?.email}</p>
					<div class="mt-3 pt-3 border-t border-slate-600">
						<p class="text-xs text-slate-500">Hourly Rate</p>
						<p class="text-base font-semibold text-emerald-400">
							{formatCurrencyRate(client?.hourly_rate || 0, client?.currency || 'USD')}
						</p>
					</div>
				</div>

				<!-- Dates & Summary -->
				<div class="space-y-3">
					<div class="bg-slate-700/30 rounded-xl p-4 border border-slate-600">
						<div class="flex items-center gap-2 mb-2">
							<CalendarMonthOutline class="w-4 h-4 text-slate-400" />
							<span class="text-xs text-slate-400 uppercase tracking-wide">Issue Date</span>
						</div>
						<p class="text-base font-semibold text-white">
							{formatDate(selectedInvoice.issue_date)}
						</p>
					</div>
					<div class="bg-slate-700/30 rounded-xl p-4 border border-slate-600">
						<div class="flex items-center gap-2 mb-2">
							<CalendarMonthOutline class="w-4 h-4 text-slate-400" />
							<span class="text-xs text-slate-400 uppercase tracking-wide">Due Date</span>
						</div>
						<p class="text-base font-semibold text-white">
							{formatDate(selectedInvoice.due_date)}
						</p>
					</div>
				</div>
			</div>

			<!-- Time Entries -->
			<div class="space-y-3">
				<div class="flex items-center gap-2">
					<svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<h4 class="text-base font-semibold text-white">Time Entries</h4>
					<span class="text-sm text-slate-400">({selectedInvoice.time_entries.length})</span>
				</div>
				<div class="border border-slate-600 rounded-xl overflow-hidden bg-slate-700/30">
					<Table class="!bg-transparent">
						<TableHead class="!bg-slate-700/50 !border-slate-600">
							<TableHeadCell class="!bg-transparent !text-slate-300">Date</TableHeadCell>
							<TableHeadCell class="!bg-transparent !text-slate-300">Description</TableHeadCell>
							<TableHeadCell class="!bg-transparent !text-slate-300">Hours</TableHeadCell>
							<TableHeadCell class="!bg-transparent !text-slate-300">Rate</TableHeadCell>
							<TableHeadCell class="!bg-transparent !text-slate-300">Amount</TableHeadCell>
						</TableHead>
						<TableBody class="!bg-transparent">
							{#each selectedInvoice.time_entries as entry}
								<TableBodyRow class="!bg-transparent !border-slate-600">
									<TableBodyCell class="!bg-transparent !text-white"
										>{formatDate(entry.date)}</TableBodyCell
									>
									<TableBodyCell class="!bg-transparent !text-slate-300"
										>{entry.description || 'No description'}</TableBodyCell
									>
									<TableBodyCell class="!bg-transparent !text-white font-medium"
										>{entry.hours}h</TableBodyCell
									>
									<TableBodyCell class="!bg-transparent !text-slate-300"
										>{formatCurrencyRate(
											client?.hourly_rate || 0,
											client?.currency || 'USD'
										)}</TableBodyCell
									>
									<TableBodyCell class="!bg-transparent !text-emerald-400 font-semibold"
										>{formatCurrency(
											entry.hours * (client?.hourly_rate || 0),
											client?.currency || 'USD'
										)}</TableBodyCell
									>
								</TableBodyRow>
							{/each}
						</TableBody>
					</Table>
				</div>
			</div>

			<!-- Summary -->
			<div
				class="bg-gradient-to-br from-blue-500/10 to-violet-500/10 border border-blue-500/20 rounded-xl p-6"
			>
				<div class="flex items-center justify-between mb-4">
					<div class="flex items-center gap-2">
						<svg
							class="w-5 h-5 text-blue-400"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"
							/>
						</svg>
						<span class="text-sm text-slate-400">Total Hours</span>
					</div>
					<span class="text-xl font-bold text-white">{selectedInvoice.total_hours}h</span>
				</div>
				<div class="flex items-center justify-between pt-4 border-t border-slate-600">
					<div class="flex items-center gap-2">
						<svg
							class="w-6 h-6 text-emerald-400"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
						<span class="text-base font-semibold text-white">Total Amount</span>
					</div>
					<span class="text-3xl font-bold text-emerald-400"
						>{formatCurrency(selectedInvoice.total_amount, client?.currency || 'USD')}</span
					>
				</div>
			</div>

			<!-- Notes -->
			{#if selectedInvoice.notes}
				<div class="bg-slate-700/30 border border-slate-600 rounded-xl p-5">
					<div class="flex items-center gap-2 mb-3">
						<svg
							class="w-5 h-5 text-slate-400"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
							/>
						</svg>
						<h4 class="text-base font-semibold text-white">Notes</h4>
					</div>
					<p class="text-sm text-slate-300 leading-relaxed">{selectedInvoice.notes}</p>
				</div>
			{/if}

			<!-- Action Buttons -->
			<div class="flex gap-3 pt-4">
				<Button
					color="blue"
					class="flex-1 !py-3 !text-base !font-semibold"
					onclick={() => (showViewModal = false)}
				>
					Close
				</Button>
				<Button
					color="alternative"
					class="!py-3 px-6 !text-base !font-semibold"
					onclick={() => selectedInvoice && downloadPDF(selectedInvoice.id)}
				>
					<DownloadOutline class="w-4 h-4 mr-2" />
					Download PDF
				</Button>
			</div>
		</div>
	{/if}
</Modal>

<!-- Toast Notification -->
{#if showToast}
	<Toast color="green" class="fixed bottom-4 right-4">
		{#snippet icon()}
			<CheckCircleSolid class="w-5 h-5" />
		{/snippet}
		{toastMessage}
	</Toast>
{/if}

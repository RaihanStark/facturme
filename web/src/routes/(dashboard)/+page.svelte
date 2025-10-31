<script lang="ts">
	import { Badge, Datepicker, Button, ButtonGroup } from 'flowbite-svelte';
	import {
		ClockSolid,
		DollarOutline,
		UsersGroupSolid,
		FileInvoiceSolid,
		CalendarMonthOutline
	} from 'flowbite-svelte-icons';
	import type { TimeEntry, Invoice } from '$lib/api';
	import { api } from '$lib/api';
	import CardWithHeader from '$lib/components/CardWithHeader.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { formatCurrency } from '$lib/utils/currency';
	import { authStore } from '$lib/stores';

	let dateRange = $state<{ from: Date | undefined; to: Date | undefined }>({
		from: new Date(new Date().getFullYear(), new Date().getMonth(), 1),
		to: new Date()
	});
	let activeRange = $state<'today' | 'week' | 'month' | 'year' | 'all' | 'custom'>('month');
	let stats = $state<{ totalHours: number; totalRevenue: number; unpaidInvoices: number; paidInvoices: number } | null>(null);
	let isLoadingStats = $state(false);
	let recentTimeEntries = $state<TimeEntry[]>([]);
	let isLoadingTimeEntries = $state(false);
	let recentInvoices = $state<Invoice[]>([]);
	let isLoadingInvoices = $state(false);

	// Derive formatted date strings from dateRange
	const dateParams = $derived({
		from: dateRange.from?.toISOString().split('T')[0],
		to: dateRange.to?.toISOString().split('T')[0]
	});

	// Generic async data loader
	async function loadData<T>(
		fetcher: () => Promise<T>,
		onSuccess: (data: T) => void,
		setLoading: (loading: boolean) => void
	) {
		setLoading(true);
		try {
			const data = await fetcher();
			onSuccess(data);
		} catch (error) {
			console.error('Failed to fetch data:', error);
		} finally {
			setLoading(false);
		}
	}

	function setDateRange(range: 'today' | 'week' | 'month' | 'year' | 'all') {
		const today = new Date();
		today.setHours(0, 0, 0, 0);

		activeRange = range;

		switch (range) {
			case 'today':
				dateRange.from = new Date(today);
				dateRange.to = new Date();
				break;
			case 'week':
				const weekStart = new Date(today);
				weekStart.setDate(today.getDate() - today.getDay());
				dateRange.from = weekStart;
				dateRange.to = new Date();
				break;
			case 'month':
				dateRange.from = new Date(today.getFullYear(), today.getMonth(), 1);
				dateRange.to = new Date();
				break;
			case 'year':
				dateRange.from = new Date(today.getFullYear(), 0, 1);
				dateRange.to = new Date();
				break;
			case 'all':
				dateRange.from = undefined;
				dateRange.to = undefined;
				break;
		}
	}

	function handleDateChange() {
		activeRange = 'custom';
	}

	// Single effect that loads all dashboard data when date range changes
	$effect(() => {
		// React to dateParams changes
		dateParams.from;
		dateParams.to;

		// Load stats
		loadData(
			() => api.getDashboardStats(dateParams.from, dateParams.to),
			(data) => {
				stats = {
					totalHours: data.total_hours,
					totalRevenue: data.total_revenue,
					unpaidInvoices: data.unpaid_invoices,
					paidInvoices: data.paid_invoices
				};
			},
			(loading) => isLoadingStats = loading
		);

		// Load recent time entries
		loadData(
			() => api.getRecentTimeEntries(dateParams.from, dateParams.to, 5),
			(data) => recentTimeEntries = data,
			(loading) => isLoadingTimeEntries = loading
		);

		// Load recent invoices
		loadData(
			() => api.getRecentInvoices(dateParams.from, dateParams.to, 5),
			(data) => recentInvoices = data,
			(loading) => isLoadingInvoices = loading
		);
	});

	function formatDate(date: string): string {
		return new Date(date).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function getStatusColor(status: string): 'gray' | 'blue' | 'green' | 'red' {
		const colors: Record<string, 'gray' | 'blue' | 'green' | 'red'> = {
			draft: 'gray',
			sent: 'blue',
			paid: 'green',
			overdue: 'red'
		};
		return colors[status] || 'gray';
	}
</script>

<div class="space-y-6">
	<!-- Header (Always visible) -->
	<div>
		<h1 class="text-3xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
		<p class="mt-2 text-gray-600 dark:text-gray-400">Overview of your work and invoices</p>
	</div>

	<!-- Date Range Filter (Always visible) -->
	<div id="date-filter" class="bg-slate-800 rounded-xl border border-slate-700 p-4">
		<div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
			<div class="flex items-center gap-3">
				<CalendarMonthOutline class="w-5 h-5 text-slate-400" />
				<span class="text-sm font-medium text-white">Filter by Date Range</span>
			</div>
			<div class="flex flex-col sm:flex-row items-start sm:items-center gap-3">
				<ButtonGroup>
					<Button
						size="xs"
						color={activeRange === 'today' ? 'primary' : 'light'}
						onclick={() => setDateRange('today')}
					>
						Today
					</Button>
					<Button
						size="xs"
						color={activeRange === 'week' ? 'primary' : 'light'}
						onclick={() => setDateRange('week')}
					>
						This Week
					</Button>
					<Button
						size="xs"
						color={activeRange === 'month' ? 'primary' : 'light'}
						onclick={() => setDateRange('month')}
					>
						This Month
					</Button>
					<Button
						size="xs"
						color={activeRange === 'year' ? 'primary' : 'light'}
						onclick={() => setDateRange('year')}
					>
						This Year
					</Button>
					<Button
						size="xs"
						color={activeRange === 'all' ? 'primary' : 'light'}
						onclick={() => setDateRange('all')}
					>
						All Time
					</Button>
				</ButtonGroup>
				<div class="flex items-center gap-2">
					<Datepicker
						inputClass="min-w-[288px]"
						range
						bind:rangeFrom={dateRange.from}
						bind:rangeTo={dateRange.to}
						onchange={handleDateChange}
					/>
				</div>
			</div>
		</div>
	</div>

	<!-- Stats Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4" id="stats-grid">
			<div
				class="bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl p-6 shadow-lg border border-blue-400/20"
			>
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-blue-100">Total Hours</p>
						{#if isLoadingStats}
							<div class="h-9 bg-blue-300/40 rounded w-20 mt-2 animate-pulse"></div>
						{:else}
							<p class="text-3xl font-bold text-white mt-2">{stats?.totalHours ?? 0}</p>
						{/if}
					</div>
					<div class="p-3 bg-white/20 backdrop-blur-sm rounded-lg">
						<ClockSolid class="w-7 h-7 text-white" />
					</div>
				</div>
			</div>

			<div
				class="bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-xl p-6 shadow-lg border border-emerald-400/20"
			>
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-emerald-100">Total Revenue</p>
						{#if isLoadingStats}
							<div class="h-9 bg-emerald-300/40 rounded w-32 mt-2 animate-pulse"></div>
						{:else}
							<p class="text-3xl font-bold text-white mt-2">
								{formatCurrency(stats?.totalRevenue ?? 0, $authStore.user?.currency)}
							</p>
						{/if}
					</div>
					<div class="p-3 bg-white/20 backdrop-blur-sm rounded-lg">
						<DollarOutline class="w-7 h-7 text-white" />
					</div>
				</div>
			</div>

			<div
				class="bg-gradient-to-br from-violet-500 to-violet-600 rounded-xl p-6 shadow-lg border border-violet-400/20"
			>
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-violet-100">Unpaid Invoices</p>
						{#if isLoadingStats}
							<div class="h-9 bg-violet-300/40 rounded w-28 mt-2 animate-pulse"></div>
						{:else}
							<p class="text-3xl font-bold text-white mt-2">{formatCurrency(stats?.unpaidInvoices ?? 0, $authStore.user?.currency)}</p>
						{/if}
					</div>
					<div class="p-3 bg-white/20 backdrop-blur-sm rounded-lg">
						<FileInvoiceSolid class="w-7 h-7 text-white" />
					</div>
				</div>
			</div>

			<div
				class="bg-gradient-to-br from-amber-500 to-amber-600 rounded-xl p-6 shadow-lg border border-amber-400/20"
			>
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-amber-100">Paid Invoices</p>
						{#if isLoadingStats}
							<div class="h-9 bg-amber-300/40 rounded w-28 mt-2 animate-pulse"></div>
						{:else}
							<p class="text-3xl font-bold text-white mt-2">
								{formatCurrency(stats?.paidInvoices ?? 0, $authStore.user?.currency)}
							</p>
						{/if}
					</div>
					<div class="p-3 bg-white/20 backdrop-blur-sm rounded-lg">
						<FileInvoiceSolid class="w-7 h-7 text-white" />
					</div>
				</div>
			</div>
		</div>

		<!-- Recent Activity -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Recent Time Entries -->
			<div id="recent-time">
				<CardWithHeader title="Recent Time Entries" href="/time">
					{#if isLoadingTimeEntries}
						<div class="space-y-3 animate-pulse">
							{#each Array(5) as _, i}
								<div class="bg-slate-700/30 rounded-lg p-4">
									<div class="flex items-start justify-between gap-4">
										<div class="flex-1 min-w-0">
											<div class="h-5 bg-slate-600 rounded w-32 mb-2"></div>
											<div class="h-4 bg-slate-600/60 rounded w-48 mb-2"></div>
											<div class="h-3 bg-slate-600/40 rounded w-24"></div>
										</div>
										<div class="flex-shrink-0">
											<div class="w-[60px] h-8 bg-slate-600 rounded-lg"></div>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{:else if recentTimeEntries.length === 0}
						<EmptyState title="No time entries yet" description="Start tracking your work hours">
							{#snippet icon()}
								<ClockSolid class="w-8 h-8 text-slate-500" />
							{/snippet}
						</EmptyState>
					{:else}
						<div class="space-y-3">
							{#each recentTimeEntries as entry}
								<div class="bg-slate-700/30 hover:bg-slate-700/50 rounded-lg p-4 transition-colors">
									<div class="flex items-start justify-between gap-4">
										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-2 mb-1">
												<p class="font-medium text-white truncate">
													{entry.client_name || 'Unknown'}
												</p>
											</div>
											<p class="text-sm text-slate-300 line-clamp-2 mb-2">
												{entry.description || 'No description'}
											</p>
											<p class="text-xs text-slate-400">{formatDate(entry.date)}</p>
										</div>
										<div class="flex-shrink-0">
											<span
												class="inline-flex items-center justify-center min-w-[60px] px-3 py-1.5 bg-blue-500/10 border border-blue-500/20 rounded-lg text-blue-400 font-semibold text-sm"
											>
												{entry.hours}h
											</span>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</CardWithHeader>
			</div>

			<!-- Recent Invoices -->
			<div id="recent-invoices">
				<CardWithHeader title="Recent Invoices" href="/invoices">
					{#if isLoadingInvoices}
						<div class="space-y-3 animate-pulse">
							{#each Array(5) as _, i}
								<div class="bg-slate-700/30 rounded-lg p-4">
									<div class="flex items-start justify-between gap-4">
										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-2 mb-2">
												<div class="h-5 bg-slate-600 rounded w-24"></div>
												<div class="h-5 bg-slate-600 rounded w-16"></div>
											</div>
											<div class="h-4 bg-slate-600/60 rounded w-40 mb-2"></div>
											<div class="h-3 bg-slate-600/40 rounded w-32"></div>
										</div>
										<div class="flex-shrink-0">
											<div class="w-[80px] h-8 bg-slate-600 rounded-lg"></div>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{:else if recentInvoices.length === 0}
						<EmptyState title="No invoices yet" description="Create your first invoice">
							{#snippet icon()}
								<FileInvoiceSolid class="w-8 h-8 text-slate-500" />
							{/snippet}
						</EmptyState>
					{:else}
						<div class="space-y-3">
							{#each recentInvoices as invoice}
								<div class="bg-slate-700/30 hover:bg-slate-700/50 rounded-lg p-4 transition-colors">
									<div class="flex items-start justify-between gap-4">
										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-2 mb-1">
												<p class="font-medium text-white">
													{invoice.invoice_number}
												</p>
												<Badge color={getStatusColor(invoice.status)} class="capitalize">
													{invoice.status}
												</Badge>
											</div>
											<p class="text-sm text-slate-300 mb-2 truncate">
												{invoice.client_name || 'Unknown'}
											</p>
											<p class="text-xs text-slate-400">Due: {formatDate(invoice.due_date)}</p>
										</div>
										<div class="flex-shrink-0">
											<span
												class="inline-flex items-center justify-center min-w-[80px] px-3 py-1.5 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-emerald-400 font-semibold text-sm"
											>
												{formatCurrency(invoice.total_amount, invoice.client_currency || 'USD')}
											</span>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</CardWithHeader>
			</div>
		</div>
</div>

<script lang="ts">
	import { Button, Toast, ButtonGroup, Badge } from 'flowbite-svelte';
	import {
		PlusOutline,
		EditOutline,
		TrashBinOutline,
		CheckCircleSolid,
		ChevronLeftOutline,
		ChevronRightOutline,
		ClockSolid,
		CalendarMonthOutline
	} from 'flowbite-svelte-icons';
	import { api, type TimeEntry, type Client, type CreateTimeEntryRequest, type TimeEntriesStats } from '$lib/api';
	import CardWithHeader from '$lib/components/CardWithHeader.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import TimeEntryModal from '$lib/components/TimeEntryModal.svelte';
	import { formatCurrency, formatCurrencyRate } from '$lib/utils/currency';
	import { authStore } from '$lib/stores';

	let showModal = $state(false);
	let editingEntry: TimeEntry | null = $state(null);
	let showToast = $state(false);
	let toastMessage = $state('');
	let isSaving = $state(false);
	let viewMode: 'daily' | 'weekly' | 'monthly' = $state('weekly');
	let currentDate = $state(new Date());
	let statsData = $state<TimeEntriesStats | null>(null);
	let isLoading = $state(false);
	let clients = $state<Client[]>([]);

	let formData = $state({
		clientId: 0,
		date: new Date().toISOString().split('T')[0],
		hours: 0,
		notes: ''
	});

	// Load stats from backend
	async function loadStats() {
		isLoading = true;
		try {
			const dateStr = currentDate.toISOString().split('T')[0];
			statsData = await api.getTimeEntriesStats(viewMode, dateStr);
		} catch (error) {
			console.error('Failed to load time entries stats:', error);
			statsData = null;
		} finally {
			isLoading = false;
		}
	}

	// Load stats when view mode or date changes
	$effect(() => {
		viewMode;
		currentDate;
		Promise.resolve().then(() => loadStats());
	});

	// Load clients for the modal
	async function loadClients() {
		if (clients.length > 0) return;
		try {
			clients = await api.getClients();
		} catch (error) {
			console.error('Failed to load clients:', error);
		}
	}

	// Derive grouped entries from statsData
	const groupedEntries = $derived.by(() => {
		if (!statsData || !statsData.entries || statsData.entries.length === 0) return [];

		const groups = new Map<string, TimeEntry[]>();
		const entries = [...statsData.entries]; // Create a copy to avoid mutating original
		entries
			.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
			.forEach((entry) => {
				const dateKey = entry.date;
				if (!groups.has(dateKey)) {
					groups.set(dateKey, []);
				}
				groups.get(dateKey)!.push(entry);
			});
		return Array.from(groups.entries());
	});

	// Derive stats values
	const totalHours = $derived(statsData?.total_hours ?? 0);
	const totalRevenue = $derived(statsData?.total_revenue ?? 0);

	async function openAddModal() {
		await loadClients();
		editingEntry = null;
		formData = {
			clientId: clients[0]?.id || 0,
			date: currentDate.toISOString().split('T')[0],
			hours: 0,
			notes: ''
		};
		showModal = true;
	}

	async function openEditModal(entry: TimeEntry) {
		await loadClients();
		editingEntry = entry;
		formData = {
			clientId: entry.client_id,
			date: entry.date,
			hours: entry.hours,
			notes: entry.description || ''
		};
		showModal = true;
	}

	async function handleSubmit() {
		isSaving = true;
		try {
			const requestData: CreateTimeEntryRequest = {
				client_id: formData.clientId,
				date: formData.date,
				hours: formData.hours,
				description: formData.notes || undefined
			};

			if (editingEntry) {
				await api.updateTimeEntry(editingEntry.id, requestData);
				toastMessage = 'Time entry updated successfully';
			} else {
				await api.createTimeEntry(requestData);
				toastMessage = 'Time entry added successfully';
			}

			showModal = false;
			showToast = true;
			await loadStats();
		} catch (error) {
			toastMessage = error instanceof Error ? error.message : 'Operation failed';
			showToast = true;
		} finally {
			isSaving = false;
		}
	}

	async function handleDelete(entryId: number) {
		if (!confirm('Are you sure you want to delete this time entry?')) return;

		try {
			await api.deleteTimeEntry(entryId);
			toastMessage = 'Time entry deleted successfully';
			showToast = true;
			await loadStats();
		} catch (error) {
			toastMessage = 'Failed to delete time entry';
			showToast = true;
		}
	}

	function navigatePrevious() {
		if (viewMode === 'daily') {
			currentDate = new Date(currentDate.setDate(currentDate.getDate() - 1));
		} else if (viewMode === 'weekly') {
			currentDate = new Date(currentDate.setDate(currentDate.getDate() - 7));
		} else {
			currentDate = new Date(currentDate.setMonth(currentDate.getMonth() - 1));
		}
	}

	function navigateNext() {
		if (viewMode === 'daily') {
			currentDate = new Date(currentDate.setDate(currentDate.getDate() + 1));
		} else if (viewMode === 'weekly') {
			currentDate = new Date(currentDate.setDate(currentDate.getDate() + 7));
		} else {
			currentDate = new Date(currentDate.setMonth(currentDate.getMonth() + 1));
		}
	}

	function navigateToday() {
		currentDate = new Date();
	}

	function getClientColor(
		clients: Client[],
		clientId: number
	): 'blue' | 'emerald' | 'violet' | 'amber' | 'rose' | 'cyan' | 'pink' | 'indigo' {
		const colors: Array<
			'blue' | 'emerald' | 'violet' | 'amber' | 'rose' | 'cyan' | 'pink' | 'indigo'
		> = ['blue', 'emerald', 'violet', 'amber', 'rose', 'cyan', 'pink', 'indigo'];
		const index = clients.findIndex((c) => c.id === clientId);
		return colors[index % colors.length];
	}

	function formatDate(date: Date | string): string {
		return new Date(date).toLocaleDateString('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function formatDateShort(date: Date | string): string {
		return new Date(date).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric'
		});
	}

	function getDateRangeDisplay(): string {
		if (viewMode === 'daily') {
			return formatDate(currentDate);
		} else if (viewMode === 'weekly') {
			// Start week on Monday
			const day = currentDate.getDay();
			const daysToSubtract = day === 0 ? 6 : day - 1; // Sunday is 0, so go back 6 days

			const startOfWeek = new Date(currentDate);
			startOfWeek.setDate(currentDate.getDate() - daysToSubtract);

			const endOfWeek = new Date(startOfWeek);
			endOfWeek.setDate(startOfWeek.getDate() + 6);

			return `${formatDateShort(startOfWeek)} - ${formatDateShort(endOfWeek)}`;
		} else {
			return currentDate.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
		}
	}

	$effect(() => {
		if (showToast) {
			setTimeout(() => {
				showToast = false;
			}, 3000);
		}
	});
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-3xl font-bold text-white">Time Tracking</h1>
			<p class="mt-2 text-slate-400">Log and manage your work hours</p>
		</div>
		<Button onclick={() => openAddModal()} color="primary">
			<PlusOutline class="w-4 h-4 mr-2" />
			Log Time
		</Button>
	</div>

		<!-- View Controls -->
		<div class="bg-slate-800 rounded-xl border border-slate-700 p-4">
			<div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
				<div class="flex items-center gap-3">
					<ButtonGroup>
						<Button
							size="sm"
							color={viewMode === 'daily' ? 'primary' : 'light'}
							onclick={() => (viewMode = 'daily')}
						>
							Daily
						</Button>
						<Button
							size="sm"
							color={viewMode === 'weekly' ? 'primary' : 'light'}
							onclick={() => (viewMode = 'weekly')}
						>
							Weekly
						</Button>
						<Button
							size="sm"
							color={viewMode === 'monthly' ? 'primary' : 'light'}
							onclick={() => (viewMode = 'monthly')}
						>
							Monthly
						</Button>
					</ButtonGroup>
				</div>

				<div class="flex items-center gap-3">
					<Button size="sm" color="light" onclick={navigatePrevious}>
						<ChevronLeftOutline class="w-4 h-4" />
					</Button>
					<div class="flex items-center gap-2 min-w-[200px] justify-center">
						<CalendarMonthOutline class="w-4 h-4 text-slate-400" />
						<span class="text-sm font-medium text-white">{getDateRangeDisplay()}</span>
					</div>
					<Button size="sm" color="light" onclick={navigateToday}>Today</Button>
					<Button size="sm" color="light" onclick={navigateNext}>
						<ChevronRightOutline class="w-4 h-4" />
					</Button>
				</div>
			</div>
		</div>

		<!-- Summary Stats -->
		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<div class="bg-slate-800 rounded-xl border border-slate-700 p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-slate-400">Total Hours</p>
						{#if isLoading}
							<div class="h-9 bg-slate-700 rounded w-20 mt-2 animate-pulse"></div>
						{:else}
							<p class="text-3xl font-bold text-white mt-2">{totalHours.toFixed(2)}h</p>
						{/if}
					</div>
					<div class="p-3 bg-blue-500/10 rounded-lg">
						<ClockSolid class="w-8 h-8 text-blue-400" />
					</div>
				</div>
			</div>

			<div class="bg-slate-800 rounded-xl border border-slate-700 p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-slate-400">Estimated Revenue</p>
						{#if isLoading}
							<div class="h-9 bg-slate-700 rounded w-32 mt-2 animate-pulse"></div>
						{:else}
							<p class="text-3xl font-bold text-white mt-2">
								{formatCurrency(totalRevenue, $authStore.user?.currency)}
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
								d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
					</div>
				</div>
			</div>
		</div>

		<!-- Time Entries -->
		{#if isLoading}
			<div class="space-y-4">
				{#each Array(2) as _}
					<div class="bg-slate-800 rounded-xl border border-slate-700 p-6">
						<div class="flex items-center justify-between mb-4">
							<div class="h-5 bg-slate-700 rounded w-48 animate-pulse"></div>
							<div class="h-5 bg-slate-700 rounded w-16 animate-pulse"></div>
						</div>
						<div class="space-y-3">
							{#each Array(2) as _}
								<div class="h-20 bg-slate-700/30 rounded-lg animate-pulse"></div>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		{:else if groupedEntries.length === 0}
			<CardWithHeader title="Time Entries">
				<EmptyState
					title="No time entries for this period"
					description="Log your first entry to get started"
				>
					{#snippet icon()}
						<ClockSolid class="w-8 h-8 text-slate-500" />
					{/snippet}
				</EmptyState>
			</CardWithHeader>
		{:else}
			<div class="space-y-4">
				{#each groupedEntries as [dateKey, entries]}
					{@const dayTotal = entries.reduce((sum, e) => sum + e.hours, 0)}
					<div class="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
						<!-- Date Header -->
						<div class="bg-slate-700/50 px-6 py-3 border-b border-slate-700">
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-3">
									<CalendarMonthOutline class="w-5 h-5 text-slate-400" />
									<h3 class="font-semibold text-white">{formatDate(new Date(dateKey))}</h3>
								</div>
								<div class="flex items-center gap-4">
									<span class="text-sm text-slate-400">{entries.length} entries</span>
									<Badge color="blue">{dayTotal.toFixed(2)}h</Badge>
								</div>
							</div>
						</div>

						<!-- Entries List -->
						<div class="p-4 space-y-3">
							{#each entries as entry}
								<div
									class="bg-slate-700/30 hover:bg-slate-700/50 rounded-lg p-4 transition-colors group"
								>
									<div class="flex items-start justify-between gap-4">
										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-2 mb-2">
												<Badge color={getClientColor(clients, entry.client_id)} class="!text-xs">
													{entry.client_name || 'Unknown'}
												</Badge>
												<span
													class="text-sm font-semibold text-blue-400 bg-blue-500/10 px-2 py-0.5 rounded"
												>
													{entry.hours}h
												</span>
												<span
													class="text-sm font-medium text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded"
												>
													{formatCurrencyRate(entry.hourly_rate, entry.client_currency || 'USD')}
												</span>
											</div>
											<p class="text-slate-300 text-sm leading-relaxed">
												{entry.description || 'No description'}
											</p>
										</div>
										<div
											class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity"
										>
											<Button size="xs" color="light" onclick={() => openEditModal(entry)}>
												<EditOutline class="w-3.5 h-3.5" />
											</Button>
											<Button size="xs" color="red" onclick={() => handleDelete(entry.id)}>
												<TrashBinOutline class="w-3.5 h-3.5" />
											</Button>
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		{/if}

	<!-- Add/Edit Time Entry Modal -->
	<TimeEntryModal
		bind:open={showModal}
		{clients}
		{editingEntry}
		{formData}
		{isSaving}
		onSubmit={handleSubmit}
		onCancel={() => (showModal = false)}
		onUpdateFormData={(updates) => {
			if (updates.clientId !== undefined) formData.clientId = updates.clientId;
			if (updates.date !== undefined) formData.date = updates.date;
			if (updates.hours !== undefined) formData.hours = updates.hours;
			if (updates.notes !== undefined) formData.notes = updates.notes;
		}}
	/>
</div>

<!-- Toast Notification -->
{#if showToast}
	<Toast color="green" class="fixed bottom-4 right-4">
		{#snippet icon()}
			<CheckCircleSolid class="w-5 h-5" />
		{/snippet}
		{toastMessage}
	</Toast>
{/if}

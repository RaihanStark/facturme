<script lang="ts">
	import { Button, Modal } from 'flowbite-svelte';
	import {
		ChevronLeftOutline,
		ChevronRightOutline,
		CalendarMonthOutline,
		ClockSolid,
		EditOutline
	} from 'flowbite-svelte-icons';
	import { goto } from '$app/navigation';
	import { api, type Client, type TimeEntry } from '$lib/api';
	import { invalidate } from '$app/navigation';
	import TimeEntryModal from '$lib/components/TimeEntryModal.svelte';

	let { data } = $props();
	let currentOffset = $derived(data.offset || 0);
	let dateRange = $derived(getDisplayDateRange(currentOffset));
	let showModal = $state(false);
	let showSelectionModal = $state(false);
	let selectedDate = $state('');
	let clients: Client[] = $state([]);
	let dayTimeEntries: TimeEntry[] = $state([]);
	let editingEntry: TimeEntry | null = $state(null);
	let isSaving = $state(false);
	let hoveredDate = $state<string | null>(null);
	let tooltipPosition = $state({ x: 0, y: 0 });
	let arrowPosition = $state({ x: 0 });
	let arrowOffset = $derived(arrowPosition.x - tooltipPosition.x);
	let heatmapData = $state<Awaited<ReturnType<typeof api.getHeatmap>> | null>(null);

	let formData = $state({
		clientId: 0,
		date: '',
		hours: 0,
		notes: ''
	});

	// Load heatmap data once it's available
	$effect(() => {
		data.heatmap.then((heatmap) => {
			heatmapData = heatmap;
		});
	});

	// Lazy load clients only when modal opens
	async function ensureClientsLoaded() {
		if (clients.length === 0) {
			try {
				clients = await api.getClients();
			} catch (error) {
				console.error('Failed to load clients:', error);
			}
		}
	}

	function getEntriesForDate(date: string): TimeEntry[] {
		if (!heatmapData?.entries) return [];
		return heatmapData.entries[date] || [];
	}

	function handleDayHover(date: Date, event: MouseEvent) {
		const dateKey = formatDateKey(date);
		const entries = getEntriesForDate(dateKey);

		if (entries.length > 0) {
			hoveredDate = dateKey;
			const target = event.currentTarget as HTMLElement;
			const rect = target.getBoundingClientRect();

			// Estimate tooltip width (min-w-[280px], max-w-[400px])
			const tooltipWidth = 300; // approximate width
			const tooltipPadding = 20; // padding from screen edge

			// Calculate element center (where arrow should point)
			const elementCenterX = rect.left + rect.width / 2;

			// Calculate initial centered position for tooltip
			let tooltipX = elementCenterX;

			// Adjust if tooltip would overflow right edge
			if (tooltipX + tooltipWidth / 2 > window.innerWidth - tooltipPadding) {
				tooltipX = window.innerWidth - tooltipWidth / 2 - tooltipPadding;
			}

			// Adjust if tooltip would overflow left edge
			if (tooltipX - tooltipWidth / 2 < tooltipPadding) {
				tooltipX = tooltipWidth / 2 + tooltipPadding;
			}

			tooltipPosition = {
				x: tooltipX,
				y: rect.top - 10
			};

			// Arrow position is relative to tooltip, pointing to element center
			arrowPosition = {
				x: elementCenterX
			};
		}
	}

	function handleDayLeave() {
		hoveredDate = null;
	}

	function getWeekStart(date: Date): Date {
		const day = date.getDay();
		const diff = day === 0 ? -6 : 1 - day;
		const monday = new Date(date);
		monday.setDate(date.getDate() + diff);
		monday.setHours(0, 0, 0, 0);
		return monday;
	}

	function formatDateKey(date: Date): string {
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		return `${year}-${month}-${day}`;
	}

	function getLast4Weeks(
		heatmapData: Record<string, number>,
		offset: number
	): Array<Array<{ date: Date; day: number; month: string; hours: number }>> {
		const weeks: Array<Array<{ date: Date; day: number; month: string; hours: number }>> = [];

		// Get today and apply offset
		const today = new Date();
		const offsetDate = new Date(today);
		offsetDate.setDate(today.getDate() + offset * 7);

		const currentWeekStart = getWeekStart(offsetDate);

		// Generate 4 weeks starting from 3 weeks ago
		for (let weekOffset = -3; weekOffset <= 0; weekOffset++) {
			const weekStart = new Date(currentWeekStart);
			weekStart.setDate(currentWeekStart.getDate() + weekOffset * 7);

			const week: Array<{ date: Date; day: number; month: string; hours: number }> = [];
			for (let i = 0; i < 7; i++) {
				const currentDay = new Date(weekStart);
				currentDay.setDate(weekStart.getDate() + i);
				const dateKey = formatDateKey(currentDay);
				const hours = heatmapData[dateKey] || 0;
				const month = currentDay.toLocaleDateString('en-US', { month: 'short' });
				week.push({ date: currentDay, day: currentDay.getDate(), month, hours });
			}
			weeks.push(week);
		}

		return weeks;
	}

	function getDisplayDateRange(offset: number): { start: string; end: string } {
		const today = new Date();
		const offsetDate = new Date(today);
		offsetDate.setDate(today.getDate() + offset * 7);

		const currentWeekStart = getWeekStart(offsetDate);

		// Start date: 3 weeks ago Monday
		const start = new Date(currentWeekStart);
		start.setDate(currentWeekStart.getDate() - 21);

		// End date: Current week Sunday
		const end = new Date(currentWeekStart);
		end.setDate(currentWeekStart.getDate() + 6);

		return {
			start: start.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
			end: end.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
		};
	}

	function getColorClass(hours: number, date: Date): string {
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const currentDate = new Date(date);
		currentDate.setHours(0, 0, 0, 0);

		if (currentDate > today) {
			return 'bg-slate-700/30 border-slate-600/30 text-slate-500';
		}

		if (hours === 0) {
			return 'bg-red-500/20 border-red-500/30 text-red-300';
		} else if (hours < 5) {
			return 'bg-yellow-500/20 border-yellow-500/30 text-yellow-300';
		} else {
			return 'bg-green-500/20 border-green-500/30 text-green-300';
		}
	}

	async function navigatePrevWeek() {
		const newOffset = currentOffset - 1;
		await goto(`/calendar?offset=${newOffset}`);
	}

	async function navigateNextWeek() {
		const newOffset = currentOffset + 1;
		await goto(`/calendar?offset=${newOffset}`);
	}

	async function navigateToday() {
		await goto('/calendar');
	}

	async function handleDayClick(date: Date, hours: number) {
		selectedDate = formatDateKey(date);

		// Get entries from heatmap data
		dayTimeEntries = getEntriesForDate(selectedDate);

		if (dayTimeEntries.length === 0) {
			// No entries, open create modal directly
			await openCreateModal();
		} else {
			// One or more entries, show selection modal
			showSelectionModal = true;
		}
	}

	async function openEditModal(entry: TimeEntry) {
		// Load clients only when modal opens
		await ensureClientsLoaded();

		editingEntry = entry;
		formData = {
			clientId: entry.client_id,
			date: selectedDate,
			hours: entry.hours,
			notes: entry.description || ''
		};
		showSelectionModal = false;
		showModal = true;
	}

	async function openCreateModal() {
		// Load clients only when modal opens
		await ensureClientsLoaded();

		editingEntry = null;
		formData = {
			clientId: clients[0]?.id || 0,
			date: selectedDate,
			hours: 0,
			notes: ''
		};
		showSelectionModal = false;
		showModal = true;
	}

	async function handleSubmit() {
		isSaving = true;
		try {
			const requestData = {
				client_id: formData.clientId,
				date: formData.date,
				hours: formData.hours,
				description: formData.notes
			};

			if (editingEntry) {
				await api.updateTimeEntry(editingEntry.id, requestData);
			} else {
				await api.createTimeEntry(requestData);
			}

			showModal = false;
			await invalidate('heatmap');
		} catch (error) {
			console.error('Failed to save time entry:', error);
		} finally {
			isSaving = false;
		}
	}

	async function handleDelete() {
		if (!editingEntry) return;
		if (!confirm('Are you sure you want to delete this time entry?')) return;

		isSaving = true;
		try {
			await api.deleteTimeEntry(editingEntry.id);
			showModal = false;
			await invalidate('heatmap');
		} catch (error) {
			console.error('Failed to delete time entry:', error);
		} finally {
			isSaving = false;
		}
	}

	const weekDayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
</script>

<div class="space-y-6">
	<!-- Header (Always visible) -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-3xl font-bold text-white">Calendar</h1>
			<p class="mt-2 text-slate-400">Track your work hours with a visual heatmap</p>
		</div>
	</div>

	<!-- Navigation Controls (Always visible) -->
	<div class="bg-slate-800 rounded-xl border border-slate-700 p-4">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-3">
				<CalendarMonthOutline class="w-5 h-5 text-slate-400" />
				<h2 class="text-xl font-semibold text-white">
					{dateRange.start} - {dateRange.end}
				</h2>
			</div>
			<div class="flex items-center gap-2">
				<Button size="sm" color="light" onclick={navigatePrevWeek}>
					<ChevronLeftOutline class="w-4 h-4" />
				</Button>
				<Button size="sm" color="light" onclick={navigateToday}>
					{currentOffset === 0 ? 'Refresh' : 'Today'}
				</Button>
				<Button size="sm" color="light" onclick={navigateNextWeek} disabled={currentOffset >= 0}>
					<ChevronRightOutline class="w-4 h-4" />
				</Button>
			</div>
		</div>
	</div>

	<!-- Legend (Always visible) -->
	<div id="calendar-legend" class="bg-slate-800 rounded-lg border border-slate-700 p-3">
		<div class="flex flex-wrap items-center gap-4">
			<span class="text-xs font-medium text-slate-400">Legend:</span>
			<div class="flex items-center gap-1.5">
				<div class="w-6 h-6 bg-red-500/20 border border-red-500/30 rounded"></div>
				<span class="text-xs text-slate-300">0 hours</span>
			</div>
			<div class="flex items-center gap-1.5">
				<div class="w-6 h-6 bg-yellow-500/20 border border-yellow-500/30 rounded"></div>
				<span class="text-xs text-slate-300">&lt; 5h</span>
			</div>
			<div class="flex items-center gap-1.5">
				<div class="w-6 h-6 bg-green-500/20 border border-green-500/30 rounded"></div>
				<span class="text-xs text-slate-300">≥ 5h</span>
			</div>
			<div class="flex items-center gap-1.5">
				<div class="w-6 h-6 bg-slate-700/30 border border-slate-600/30 rounded"></div>
				<span class="text-xs text-slate-300">Future</span>
			</div>
		</div>
	</div>

	{#await data.heatmap}
		<div class="grid grid-cols-1 xl:grid-cols-2 gap-3">
			<!-- Stats Summary Skeleton -->
			<div>
				<div class="grid grid-cols-4 gap-3 animate-pulse">
					{#each Array(4) as _}
						<div class="bg-slate-800 rounded-lg border border-slate-700 p-4">
							<div class="h-3 bg-slate-700 rounded w-20 mb-2"></div>
							<div class="h-8 bg-slate-700 rounded w-16"></div>
						</div>
					{/each}
				</div>
			</div>

			<!-- Calendar Heatmap Skeleton -->
			<div>
				<div class="bg-slate-800 rounded-lg border border-slate-700 p-4 animate-pulse">
					<!-- Week day headers -->
					<div class="grid grid-cols-7 gap-2 mb-3">
						{#each Array(7) as _}
							<div class="h-4 bg-slate-700 rounded"></div>
						{/each}
					</div>

					<!-- Calendar weeks skeleton -->
					<div class="space-y-2">
						{#each Array(4) as _}
							<div class="grid grid-cols-7 gap-2">
								{#each Array(7) as _}
									<div class="aspect-square bg-slate-700 rounded-md"></div>
								{/each}
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	{:then heatmap}
		{@const weeks = getLast4Weeks(heatmap.data, currentOffset)}

		<div class="grid grid-cols-1 xl:grid-cols-2 gap-3">
			<!-- Stats Summary -->
			<div>
				<div class="grid grid-cols-4 gap-3">
					<div class="bg-slate-800 rounded-lg border border-slate-700 p-4">
						<p class="text-xs font-medium text-slate-400">Total Hours</p>
						<p class="text-2xl font-bold text-white mt-1">{heatmap.total_hours.toFixed(1)}h</p>
					</div>

					<div class="bg-slate-800 rounded-lg border border-slate-700 p-4">
						<p class="text-xs font-medium text-slate-400">Days Worked</p>
						<p class="text-2xl font-bold text-white mt-1">{heatmap.days_worked}</p>
					</div>

					<div class="bg-slate-800 rounded-lg border border-slate-700 p-4">
						<p class="text-xs font-medium text-slate-400">Days Off</p>
						<p class="text-2xl font-bold text-white mt-1">{heatmap.days_off}</p>
					</div>

					<div class="bg-slate-800 rounded-lg border border-slate-700 p-4">
						<p class="text-xs font-medium text-slate-400">Avg Hours/Day</p>
						<p class="text-2xl font-bold text-white mt-1">{heatmap.average_hours.toFixed(1)}h</p>
					</div>
				</div>
			</div>

			<!-- Calendar Heatmap -->
			<div>
				<div id="calendar-grid" class="bg-slate-800 rounded-lg border border-slate-700 p-4">
					<!-- Week day headers -->
					<div class="grid grid-cols-7 gap-2 mb-3">
						{#each weekDayNames as dayName}
							<div class="text-center text-xs font-medium text-slate-400 py-1">
								{dayName}
							</div>
						{/each}
					</div>

					<!-- Calendar weeks -->
					<div class="space-y-2">
						{#each weeks as week}
							<div class="grid grid-cols-7 gap-2">
								{#each week as { date, day, month, hours }}
									<button
										type="button"
										onclick={() => handleDayClick(date, hours)}
										onmouseenter={(e) => handleDayHover(date, e)}
										onmouseleave={handleDayLeave}
										class="aspect-square border-2 rounded-md flex flex-col items-center justify-center transition-all duration-200 hover:scale-105 hover:shadow-lg cursor-pointer {getColorClass(
											hours,
											date
										)}"
									>
										<div class="text-xs font-medium text-slate-400 mb-1">{day} {month}</div>
										<div class="text-lg font-bold">{hours > 0 ? `${hours}h` : '-'}</div>
									</button>
								{/each}
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>

	<!-- Tooltip for day details -->
	{#if hoveredDate}
		{@const entries = getEntriesForDate(hoveredDate)}
		{@const totalHours = entries.reduce((sum, entry) => sum + entry.hours, 0)}
		{@const clientGroups = entries.reduce(
			(acc, entry) => {
				const clientName = entry.client_name || 'Unknown Client';
				if (!acc[clientName]) {
					acc[clientName] = { hours: 0, entries: [] };
				}
				acc[clientName].hours += entry.hours;
				acc[clientName].entries.push(entry);
				return acc;
			},
			{} as Record<string, { hours: number; entries: TimeEntry[] }>
		)}

		<div
			class="fixed z-50 pointer-events-none"
			style="left: {tooltipPosition.x}px; top: {tooltipPosition.y}px; transform: translate(-50%, -100%);"
		>
			<div
				class="bg-slate-800 border-2 border-blue-500/50 rounded-xl shadow-2xl shadow-blue-500/20 p-4 min-w-[280px] max-w-[400px]"
			>
				<!-- Header -->
				<div class="flex items-center justify-between mb-3 pb-3 border-b border-slate-700">
					<div class="flex items-center gap-2">
						<CalendarMonthOutline class="w-4 h-4 text-blue-400" />
						<span class="text-sm font-semibold text-white">{hoveredDate}</span>
					</div>
					<div class="flex items-center gap-1.5">
						<ClockSolid class="w-4 h-4 text-emerald-400" />
						<span class="text-lg font-bold text-emerald-400">{totalHours}h</span>
					</div>
				</div>

				<!-- Entries grouped by client -->
				<div class="space-y-2">
					{#each Object.entries(clientGroups) as [clientName, group]}
						<div class="bg-slate-700/30 rounded-lg p-3">
							<div class="flex items-center justify-between mb-2">
								<div class="flex items-center gap-2">
									<div class="w-2 h-2 bg-blue-400 rounded-full"></div>
									<span class="text-sm font-semibold text-white">{clientName}</span>
								</div>
								<span class="text-sm font-bold text-emerald-400">{group.hours}h</span>
							</div>
							{#if group.entries.length > 1}
								<div class="space-y-1 mt-2 pl-4">
									{#each group.entries as entry}
										<div class="text-xs text-slate-400">
											• {entry.hours}h
											{#if entry.description}
												- {entry.description.length > 40
													? entry.description.substring(0, 40) + '...'
													: entry.description}
											{/if}
										</div>
									{/each}
								</div>
							{:else if group.entries[0].description}
								<div class="text-xs text-slate-400 mt-1 pl-4">
									{group.entries[0].description.length > 50
										? group.entries[0].description.substring(0, 50) + '...'
										: group.entries[0].description}
								</div>
							{/if}
						</div>
					{/each}
				</div>

				<!-- Footer hint -->
				<div class="mt-3 pt-3 border-t border-slate-700">
					<p class="text-xs text-slate-500 text-center">Click to edit or add entry</p>
				</div>
			</div>
			<!-- Tooltip arrow -->
			<div class="relative w-full">
				<div
					class="absolute w-3 h-3 bg-slate-800 border-r-2 border-b-2 border-blue-500/50 -mt-1.5"
					style="left: 50%; margin-left: {arrowOffset}px; transform: translateX(-50%) rotate(45deg);"
				></div>
			</div>
		</div>
	{/if}

	<!-- Entry Selection Modal (for multiple entries in a day) -->
	<Modal bind:open={showSelectionModal} size="md" autoclose={false} class="dark:bg-slate-800">
		<div class="space-y-6">
			<!-- Header -->
			<div class="flex items-center gap-4 pb-4 border-b border-slate-700">
				<div class="p-3 bg-blue-500/10 rounded-xl">
					<ClockSolid class="w-7 h-7 text-blue-400" />
				</div>
				<div>
					<h3 class="text-2xl font-bold text-white">Select Time Entry</h3>
					<p class="text-sm text-slate-400 mt-1">
						{dayTimeEntries.length === 1
							? `1 entry on ${selectedDate}`
							: `Choose which entry to edit from ${dayTimeEntries.length} entries on ${selectedDate}`}
					</p>
				</div>
			</div>

			<!-- Entry List -->
			<div class="space-y-3">
				{#each dayTimeEntries as entry (entry.id)}
					<button
						type="button"
						onclick={() => openEditModal(entry)}
						class="w-full bg-slate-700/30 hover:bg-slate-700/50 border border-slate-600 hover:border-blue-500/50 rounded-xl p-4 transition-all duration-200 text-left group"
					>
						<div class="flex items-center justify-between">
							<div class="flex-1">
								<div class="flex items-center gap-3 mb-2">
									<div
										class="p-2 bg-blue-500/10 rounded-lg group-hover:bg-blue-500/20 transition-colors"
									>
										<svg
											class="w-4 h-4 text-blue-400"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
											/>
										</svg>
									</div>
									<div>
										<h4 class="font-semibold text-white">
											{entry.client_name || 'Unknown Client'}
										</h4>
										<p class="text-sm text-slate-400 mt-0.5">
											{entry.description || 'No description'}
										</p>
									</div>
								</div>
							</div>
							<div class="flex items-center gap-4">
								<div class="text-right">
									<div class="flex items-center gap-2">
										<ClockSolid class="w-4 h-4 text-emerald-400" />
										<span class="text-xl font-bold text-emerald-400">{entry.hours}h</span>
									</div>
								</div>
								<EditOutline
									class="w-5 h-5 text-slate-400 group-hover:text-blue-400 transition-colors"
								/>
							</div>
						</div>
					</button>
				{/each}
			</div>

			<div class="space-y-3">
				<!-- Add New Entry Button -->
				<div class="pt-4 border-t border-slate-700">
					<Button
						color="blue"
						class="w-full !py-3 !text-base !font-semibold"
						onclick={openCreateModal}
					>
						<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 4v16m8-8H4"
							/>
						</svg>
						Add New Entry for This Day
					</Button>
				</div>

				<!-- Cancel Button -->
				<div class="flex gap-3">
					<Button
						type="button"
						color="alternative"
						class="flex-1 !py-3 !text-base !font-semibold"
						onclick={() => (showSelectionModal = false)}
					>
						Cancel
					</Button>
				</div>
			</div>
		</div>
	</Modal>

	<!-- Time Entry Modal -->
	<TimeEntryModal
		bind:open={showModal}
		{clients}
		{editingEntry}
		{formData}
		{isSaving}
		dateDisabled={true}
		onSubmit={handleSubmit}
		onCancel={() => (showModal = false)}
		onDelete={handleDelete}
		onUpdateFormData={(updates) => {
			if (updates.clientId !== undefined) formData.clientId = updates.clientId;
			if (updates.date !== undefined) formData.date = updates.date;
			if (updates.hours !== undefined) formData.hours = updates.hours;
			if (updates.notes !== undefined) formData.notes = updates.notes;
		}}
	/>
	{:catch error}
		<div class="bg-red-500/10 border border-red-500/20 rounded-xl p-6">
			<p class="text-red-400">Failed to load heatmap data: {error.message}</p>
		</div>
	{/await}
</div>

<script lang="ts">
	import { Modal, Label, Input, Textarea, Select, Button } from 'flowbite-svelte';
	import {
		ClockSolid,
		CalendarMonthOutline,
		CheckCircleSolid
	} from 'flowbite-svelte-icons';
	import type { Client, TimeEntry } from '$lib/api';

	interface Props {
		open: boolean;
		clients: Client[];
		editingEntry: TimeEntry | null;
		formData: {
			clientId: number;
			date: string;
			hours: number;
			notes: string;
		};
		isSaving?: boolean;
		dateDisabled?: boolean;
		onSubmit: () => void;
		onCancel: () => void;
		onDelete?: () => void;
		onUpdateFormData: (data: { clientId?: number; date?: string; hours?: number; notes?: string }) => void;
	}

	let {
		open = $bindable(),
		clients,
		editingEntry,
		formData,
		isSaving = false,
		dateDisabled = false,
		onSubmit,
		onCancel,
		onDelete,
		onUpdateFormData
	}: Props = $props();

	const clientOptions = $derived(clients.map((c) => ({ value: c.id, name: c.name })));

	function updateHours(hours: number) {
		onUpdateFormData({ hours });
	}

	function updateClientId(clientId: number) {
		onUpdateFormData({ clientId });
	}

	function updateDate(date: string) {
		onUpdateFormData({ date });
	}

	function updateNotes(notes: string) {
		onUpdateFormData({ notes });
	}
</script>

<Modal bind:open size="md" autoclose={false} class="dark:bg-slate-800">
	<form
		class="space-y-6"
		onsubmit={(e) => {
			e.preventDefault();
			onSubmit();
		}}
	>
		<!-- Header with Icon -->
		<div class="flex items-center gap-4 pb-4 border-b border-slate-700">
			<div class="p-3 bg-blue-500/10 rounded-xl">
				<ClockSolid class="w-7 h-7 text-blue-400" />
			</div>
			<div>
				<h3 class="text-2xl font-bold text-white">
					{editingEntry ? 'Edit Time Entry' : 'Log Time'}
				</h3>
				<p class="text-sm text-slate-400 mt-1">
					{editingEntry ? 'Update your time entry details' : 'Track your work hours'}
				</p>
			</div>
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
				value={formData.clientId}
				onchange={(e) => updateClientId(Number(e.currentTarget.value))}
				required
				items={clientOptions}
				class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
			/>
		</div>

		<!-- Date Selection -->
		<div class="space-y-2">
			<Label for="date" class="text-base font-semibold text-white flex items-center gap-2">
				<CalendarMonthOutline class="w-4 h-4 text-slate-400" />
				Date
			</Label>
			<Input
				id="date"
				type="date"
				value={formData.date}
				oninput={(e) => updateDate(e.currentTarget.value)}
				required
				disabled={dateDisabled}
				class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500 disabled:!opacity-50 disabled:!cursor-not-allowed"
			/>
		</div>

		<!-- Hours Input with Quick Select -->
		<div class="space-y-3">
			<Label for="hours" class="text-base font-semibold text-white flex items-center gap-2">
				<ClockSolid class="w-4 h-4 text-slate-400" />
				Hours
			</Label>

			<!-- Quick Select Buttons -->
			<div class="grid grid-cols-6 gap-2">
				{#each [0.25, 0.5, 1, 2, 4, 8] as quickHours}
					<button
						type="button"
						class="px-3 py-2.5 rounded-lg text-sm font-semibold transition-all {formData.hours ===
						quickHours
							? 'bg-blue-500 text-white shadow-lg shadow-blue-500/30'
							: 'bg-slate-700 text-slate-300 hover:bg-slate-600 border border-slate-600'}"
						onclick={() => updateHours(quickHours)}
					>
						{quickHours}h
					</button>
				{/each}
			</div>

			<!-- Manual Input -->
			<div class="relative">
				<Input
					id="hours"
					type="number"
					value={formData.hours}
					oninput={(e) => updateHours(Number(e.currentTarget.value))}
					placeholder="Enter custom hours"
					min="0.25"
					step="0.25"
					required
					class="!bg-slate-700 !border-slate-600 !text-white !pl-10 focus:!ring-blue-500 focus:!border-blue-500"
				/>
				<div class="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none">
					<ClockSolid class="w-4 h-4 text-slate-400" />
				</div>
			</div>
		</div>

		<!-- Notes/Description -->
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
				value={formData.notes}
				oninput={(e) => updateNotes(e.currentTarget.value)}
				placeholder="What did you work on? Add any relevant details..."
				rows={4}
				class="!bg-slate-700 !border-slate-600 !text-white placeholder:!text-slate-500 focus:!ring-blue-500 focus:!border-blue-500 resize-none !w-full"
			/>
			<p class="text-xs text-slate-500">
				Tip: Detailed notes help you remember what you worked on later
			</p>
		</div>

		<!-- Action Buttons -->
		<div class="flex gap-3 pt-4">
			<Button
				type="submit"
				color="blue"
				class="flex-1 !py-3 !text-base !font-semibold"
				disabled={isSaving}
			>
				{#if isSaving}
					<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
						<circle
							class="opacity-25"
							cx="12"
							cy="12"
							r="10"
							stroke="currentColor"
							stroke-width="4"
						></circle>
						<path
							class="opacity-75"
							fill="currentColor"
							d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
						></path>
					</svg>
					Saving...
				{:else}
					<CheckCircleSolid class="w-4 h-4 mr-2" />
					{editingEntry ? 'Update Entry' : 'Log Time'}
				{/if}
			</Button>
			{#if editingEntry && onDelete}
				<Button
					type="button"
					color="red"
					class="!py-3 px-6 !text-base !font-semibold"
					onclick={onDelete}
					disabled={isSaving}
				>
					Delete
				</Button>
			{/if}
			<Button
				type="button"
				color="alternative"
				class="!py-3 px-6 !text-base !font-semibold"
				onclick={onCancel}
				disabled={isSaving}
			>
				Cancel
			</Button>
		</div>
	</form>
</Modal>

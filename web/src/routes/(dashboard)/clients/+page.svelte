<script lang="ts">
	import { Button, Modal, Label, Input, Textarea, Toast, Spinner } from 'flowbite-svelte';
	import DarkSelect from '$lib/components/DarkSelect.svelte';
	import {
		PlusOutline,
		EditOutline,
		TrashBinOutline,
		CheckCircleSolid
	} from 'flowbite-svelte-icons';
	import { api, type Client, type CreateClientRequest, type SupportedCurrency } from '$lib/api';
	import { invalidate } from '$app/navigation';
	import { getCurrencySymbol, formatCurrencyRate } from '$lib/utils/currency';
	import { authStore } from '$lib/stores';
	import { onMount } from 'svelte';

	let { data } = $props();
	let showModal = $state(false);
	let editingClient: Client | null = $state(null);
	let showToast = $state(false);
	let toastMessage = $state('');
	let isSaving = $state(false);
	let supportedCurrencies = $state<SupportedCurrency[]>([]);

	let formData = $state({
		name: '',
		email: '',
		phone: '',
		company: '',
		address: '',
		hourlyRate: 0,
		currency: 'USD'
	});

	onMount(async () => {
		try {
			supportedCurrencies = await api.getSupportedCurrencies();
		} catch (error) {
			console.error('Failed to load currencies:', error);
		}
	});

	async function loadClients() {
		await invalidate('clients');
	}

	function openAddModal() {
		editingClient = null;
		formData = {
			name: '',
			email: '',
			phone: '',
			company: '',
			address: '',
			hourlyRate: 0,
			currency: $authStore.user?.currency || 'USD'
		};
		showModal = true;
	}

	function openEditModal(client: Client) {
		editingClient = client;
		formData = {
			name: client.name,
			email: client.email,
			phone: client.phone || '',
			company: client.company || '',
			address: client.address || '',
			hourlyRate: client.hourly_rate,
			currency: client.currency
		};
		showModal = true;
	}

	async function handleSubmit() {
		isSaving = true;
		try {
			const data: CreateClientRequest = {
				name: formData.name,
				email: formData.email,
				phone: formData.phone || undefined,
				company: formData.company || undefined,
				address: formData.address || undefined,
				hourly_rate: formData.hourlyRate,
				currency: formData.currency
			};

			if (editingClient) {
				await api.updateClient(editingClient.id, data);
				toastMessage = 'Client updated successfully';
			} else {
				await api.createClient(data);
				toastMessage = 'Client added successfully';
			}

			showToast = true;
			showModal = false;
			await loadClients();
		} catch (error) {
			toastMessage = error instanceof Error ? error.message : 'Operation failed';
			showToast = true;
		} finally {
			isSaving = false;
		}
	}

	async function handleDelete(clientId: number) {
		if (!confirm('Are you sure you want to delete this client?')) return;

		try {
			await api.deleteClient(clientId);
			toastMessage = 'Client deleted successfully';
			showToast = true;
			await loadClients();
		} catch (error) {
			toastMessage = 'Failed to delete client';
			showToast = true;
		}
	}
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex justify-between items-center">
		<div>
			<h1 class="text-3xl font-bold text-white">Clients</h1>
			<p class="text-gray-400 mt-1">Manage your clients and their information</p>
		</div>
		<Button color="primary" onclick={openAddModal}>
			<PlusOutline class="w-5 h-5 mr-2" />
			Add Client
		</Button>
	</div>

	<!-- Clients List -->
	{#await data.clients}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 animate-pulse">
			{#each Array(6) as _, i}
				<div class="bg-slate-800 rounded-xl border border-slate-700 p-6">
					<div class="flex justify-between items-start mb-4">
						<div class="flex-1">
							<div class="h-6 bg-slate-700 rounded w-32 mb-2"></div>
							<div class="h-4 bg-slate-700/60 rounded w-24"></div>
						</div>
						<div class="flex gap-2">
							<div class="w-9 h-9 bg-slate-700 rounded"></div>
							<div class="w-9 h-9 bg-slate-700 rounded"></div>
						</div>
					</div>

					<div class="space-y-2">
						<div class="flex items-center">
							<div class="w-4 h-4 bg-slate-700 rounded mr-2"></div>
							<div class="h-4 bg-slate-700/60 rounded w-40"></div>
						</div>
						<div class="flex items-center">
							<div class="w-4 h-4 bg-slate-700 rounded mr-2"></div>
							<div class="h-4 bg-slate-700/60 rounded w-24"></div>
						</div>
						<div class="flex items-center">
							<div class="w-4 h-4 bg-slate-700 rounded mr-2"></div>
							<div class="h-4 bg-slate-700/60 rounded w-28"></div>
						</div>
						<div class="flex items-center">
							<div class="w-4 h-4 bg-slate-700 rounded mr-2"></div>
							<div class="h-4 bg-slate-700/60 rounded w-36"></div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{:then clients}
		{#if clients.length === 0}
			<div class="bg-slate-800 rounded-xl border border-slate-700 p-12 text-center">
				<div class="max-w-md mx-auto">
					<div
						class="w-16 h-16 bg-slate-700 rounded-full flex items-center justify-center mx-auto mb-4"
					>
						<svg
							class="w-8 h-8 text-slate-400"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
							></path>
						</svg>
					</div>
					<h3 class="text-xl font-semibold text-white mb-2">No clients yet</h3>
					<p class="text-gray-400 mb-6">Get started by adding your first client</p>
					<Button color="primary" onclick={openAddModal}>
						<PlusOutline class="w-5 h-5 mr-2" />
						Add Client
					</Button>
				</div>
			</div>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each clients as client (client.id)}
					<div
						class="bg-slate-800 rounded-xl border border-slate-700 p-6 hover:border-primary-600 transition-colors"
					>
						<div class="flex justify-between items-start mb-4">
							<div class="flex-1">
								<h3 class="text-lg font-semibold text-white mb-1">{client.name}</h3>
								{#if client.company}
									<p class="text-sm text-gray-400">{client.company}</p>
								{/if}
							</div>
							<div class="flex gap-2">
								<button
									onclick={() => openEditModal(client)}
									class="p-2 text-gray-400 hover:text-primary-500 transition-colors"
									aria-label="Edit client"
								>
									<EditOutline class="w-5 h-5" />
								</button>
								<button
									onclick={() => handleDelete(client.id)}
									class="p-2 text-gray-400 hover:text-red-500 transition-colors"
									aria-label="Delete client"
								>
									<TrashBinOutline class="w-5 h-5" />
								</button>
							</div>
						</div>

						<div class="space-y-2">
							<div class="flex items-center text-sm text-gray-300">
								<svg
									class="w-4 h-4 mr-2 text-gray-400"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
									></path>
								</svg>
								{client.email}
							</div>

							<div class="flex items-center text-sm font-medium text-emerald-400">
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
									></path>
								</svg>
								{formatCurrencyRate(client.hourly_rate, client.currency)}
							</div>

							{#if client.phone}
								<div class="flex items-center text-sm text-gray-300">
									<svg
										class="w-4 h-4 mr-2 text-gray-400"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
										></path>
									</svg>
									{client.phone}
								</div>
							{/if}

							{#if client.address}
								<div class="flex items-center text-sm text-gray-300">
									<svg
										class="w-4 h-4 mr-2 text-gray-400"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
										></path>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
										></path>
									</svg>
									{client.address}
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	{:catch error}
		<div class="bg-red-900/50 border border-red-700 text-red-200 px-4 py-3 rounded-lg">
			Failed to load clients: {error.message}
		</div>
	{/await}
</div>

<!-- Add/Edit Modal -->
<Modal bind:open={showModal} size="md" autoclose={false} class="dark:bg-slate-800">
	<form
		onsubmit={(e) => {
			e.preventDefault();
			handleSubmit();
		}}
		class="space-y-6"
	>
		<!-- Header with Icon -->
		<div class="flex items-center gap-4 pb-4 border-b border-slate-700">
			<div class="p-3 bg-blue-500/10 rounded-xl">
				<svg class="w-7 h-7 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
					></path>
				</svg>
			</div>
			<div>
				<h3 class="text-2xl font-bold text-white">
					{editingClient ? 'Edit Client' : 'Add New Client'}
				</h3>
				<p class="text-sm text-slate-400 mt-1">
					{editingClient ? 'Update client information' : 'Add a new client to your workspace'}
				</p>
			</div>
		</div>

		<div class="space-y-2">
			<Label for="name" class="text-base font-semibold text-white flex items-center gap-2">
				<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
					/>
				</svg>
				Name
			</Label>
			<Input
				id="name"
				bind:value={formData.name}
				required
				placeholder="Enter client name"
				class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
			/>
		</div>

		<div class="space-y-2">
			<Label for="email" class="text-base font-semibold text-white flex items-center gap-2">
				<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
					/>
				</svg>
				Email
			</Label>
			<Input
				id="email"
				type="email"
				bind:value={formData.email}
				required
				placeholder="client@example.com"
				class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
			/>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<div class="space-y-2">
				<Label for="phone" class="text-base font-semibold text-white flex items-center gap-2">
					<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
						/>
					</svg>
					Phone
					<span class="text-xs text-slate-500 font-normal">(optional)</span>
				</Label>
				<Input
					id="phone"
					bind:value={formData.phone}
					placeholder="+1 234 567 8900"
					class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
				/>
			</div>

			<div class="space-y-2">
				<Label for="company" class="text-base font-semibold text-white flex items-center gap-2">
					<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
						/>
					</svg>
					Company
					<span class="text-xs text-slate-500 font-normal">(optional)</span>
				</Label>
				<Input
					id="company"
					bind:value={formData.company}
					placeholder="Company name"
					class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
				/>
			</div>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<div class="space-y-2">
				<Label for="hourlyRate" class="text-base font-semibold text-white flex items-center gap-2">
					<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					Hourly Rate
				</Label>
				<Input
					id="hourlyRate"
					type="number"
					bind:value={formData.hourlyRate}
					min="0"
					step="0.01"
					placeholder="0.00"
					class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
				/>
			</div>

			<div class="space-y-2">
				<Label for="currency" class="text-base font-semibold text-white flex items-center gap-2">
					<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					Currency
				</Label>
				<DarkSelect
					id="currency"
					bind:value={formData.currency}
					items={supportedCurrencies.map((c) => ({ value: c.code, name: `${c.code} - ${c.name}` }))}
					class="!bg-slate-700 !border-slate-600 !text-white focus:!ring-blue-500 focus:!border-blue-500"
				/>
			</div>
		</div>

		<div class="space-y-2">
			<Label for="address" class="text-base font-semibold text-white flex items-center gap-2">
				<svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
					/>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
					/>
				</svg>
				Address
				<span class="text-xs text-slate-500 font-normal">(optional)</span>
			</Label>
			<Textarea
				id="address"
				bind:value={formData.address}
				rows="3"
				placeholder="Enter client address..."
				class="!bg-slate-700 !border-slate-600 !text-white placeholder:!text-slate-500 focus:!ring-blue-500 focus:!border-blue-500 resize-none !w-full"
			/>
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
					<Spinner size="4" class="mr-2" />
					Saving...
				{:else}
					<CheckCircleSolid class="w-4 h-4 mr-2" />
					{editingClient ? 'Update Client' : 'Add Client'}
				{/if}
			</Button>
			<Button
				type="button"
				color="alternative"
				class="!py-3 px-6 !text-base !font-semibold"
				onclick={() => (showModal = false)}
				disabled={isSaving}
			>
				Cancel
			</Button>
		</div>
	</form>
</Modal>

<!-- Toast Notification -->
{#if showToast}
	<Toast color="green" position="top-right" class="fixed top-4 right-4 z-50" bind:open={showToast}>
		{#snippet icon()}
			<CheckCircleSolid class="w-5 h-5" />
		{/snippet}
		{toastMessage}
	</Toast>
{/if}

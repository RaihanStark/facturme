<script lang="ts">
	import { Alert, Button, Modal } from 'flowbite-svelte';
	import { ExclamationCircleSolid, InfoCircleSolid } from 'flowbite-svelte-icons';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';

	let showConfirmModal = $state(false);
	let isDeleting = $state(false);
	let error = $state('');
	let hasDemoData = $state(false);
	let isChecking = $state(true);

	onMount(async () => {
		// Check if user has demo data
		try {
			const clients = await api.getClients();
			hasDemoData = clients.some((c) => c.name.includes('🎭') || c.name.includes('(Demo)'));
		} catch (err) {
			console.error('Error checking for demo data:', err);
		} finally {
			isChecking = false;
		}
	});

	async function handleDeleteDemoData() {
		isDeleting = true;
		error = '';

		try {
			// Call server-side endpoint to delete all demo data
			await api.deleteDemoData();

			// Reload the page to refresh all data
			window.location.reload();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete demo data';
			console.error('Error deleting demo data:', err);
		} finally {
			isDeleting = false;
			showConfirmModal = false;
		}
	}
</script>

{#if !isChecking && hasDemoData}
	<Alert color="blue" class="border-0 rounded-none">
		<InfoCircleSolid slot="icon" class="w-5 h-5" />
		<span class="font-medium">Demo mode active!</span>
		You're viewing sample data to explore features.
		<button
			onclick={() => (showConfirmModal = true)}
			class="ml-2 text-blue-800 underline hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
		>
			Delete demo data
		</button>
	</Alert>
{/if}

<!-- Confirmation Modal -->
<Modal bind:open={showConfirmModal} size="sm" autoclose={false}>
	<div class="text-center space-y-4">
		<div class="flex justify-center">
			<div class="bg-red-500/10 p-4 rounded-full">
				<ExclamationCircleSolid class="w-12 h-12 text-red-400" />
			</div>
		</div>

		<div>
			<h3 class="text-xl font-bold text-white mb-2">Delete Demo Data?</h3>
			<p class="text-sm text-slate-400">
				This will permanently delete all demo clients, time entries, and invoices. This action
				cannot be undone.
			</p>
		</div>

		{#if error}
			<div class="p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
				<p class="text-sm text-red-400">{error}</p>
			</div>
		{/if}

		<div class="flex gap-3 justify-center pt-2">
			<Button color="light" onclick={() => (showConfirmModal = false)} disabled={isDeleting}>
				Cancel
			</Button>
			<Button color="red" onclick={handleDeleteDemoData} disabled={isDeleting}>
				{#if isDeleting}
					<svg
						class="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
					>
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
					Deleting...
				{:else}
					Yes, Delete All Demo Data
				{/if}
			</Button>
		</div>
	</div>
</Modal>

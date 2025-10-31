<script lang="ts">
	import { Button, Label } from 'flowbite-svelte';
	import {
		CheckCircleSolid,
		UserAddSolid,
		FileInvoiceSolid,
		RocketSolid,
		GlobeSolid
	} from 'flowbite-svelte-icons';
	import { api } from '$lib/api';
	import { authStore } from '$lib/stores';
	import { goto } from '$app/navigation';
	import DarkInput from './DarkInput.svelte';
	import DarkSelect from './DarkSelect.svelte';

	interface OnboardingClient {
		name: string;
		email: string;
		company: string;
		hourly_rate: number;
	}

	interface LocaleSettings {
		currency: string;
	}

	let { onComplete = () => {} }: { onComplete?: () => void } = $props();

	let currentStep = $state(0);
	let isLoading = $state(false);
	let error = $state('');
	let skipFirstClient = $state(false);
	let loadDemoData = $state(false);
	let clientData: OnboardingClient = $state({
		name: '',
		email: '',
		company: '',
		hourly_rate: 50
	});

	let localeSettings: LocaleSettings = $state({
		currency: 'USD'
	});

	let steps = $state([
		{ label: 'Welcome' },
		{ label: 'Currency' },
		{ label: 'First Client' },
		{ label: 'Demo Data' },
		{ label: 'Complete' }
	]);

	function nextStep() {
		if (currentStep < steps.length - 1) {
			error = '';
			currentStep++;
		}
	}

	function prevStep() {
		if (currentStep > 0) {
			error = '';
			// Reset skipFirstClient if going back to the client step (now step 2)
			if (currentStep === 3 && skipFirstClient) {
				skipFirstClient = false;
			}
			currentStep--;
		}
	}

	async function handleComplete() {
		isLoading = true;
		error = '';

		try {
			// Step 1: Create the first client (if not skipped)
			if (!skipFirstClient) {
				await api.createClient({
					name: clientData.name,
					email: clientData.email,
					company: clientData.company || undefined,
					hourly_rate: clientData.hourly_rate,
					currency: localeSettings.currency
				});
			}

			// Step 2: Complete onboarding with locale settings
			const updatedUser = await api.completeOnboarding({
				currency: localeSettings.currency
			});

			// Step 3: Update auth store with new user data
			authStore.setUser(updatedUser);

			// Step 4: Generate demo data if requested (server-side)
			if (loadDemoData) {
				await api.generateDemoData();
			}

			// Step 5: Navigate to dashboard
			onComplete();
			goto('/');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to complete onboarding';
			console.error('Onboarding error:', err);
		} finally {
			isLoading = false;
		}
	}

	function skipClient() {
		skipFirstClient = true;
		nextStep();
	}

	function canProceedFromStep(step: number): boolean {
		switch (step) {
			case 0:
				return true; // Welcome screen, always can proceed
			case 1:
				return localeSettings.currency !== ''; // Currency step
			case 2:
				// Can skip or fill in client data
				return (
					skipFirstClient ||
					(clientData.name.trim() !== '' &&
						clientData.email.trim() !== '' &&
						clientData.hourly_rate > 0)
				);
			case 3:
				return true; // Demo data step, always can proceed
			default:
				return true;
		}
	}

	// Common currencies
	const currencies = [
		{ value: 'USD', name: 'USD - US Dollar ($)', symbol: '$' },
		{ value: 'EUR', name: 'EUR - Euro (€)', symbol: '€' },
		{ value: 'GBP', name: 'GBP - British Pound (£)', symbol: '£' },
		{ value: 'JPY', name: 'JPY - Japanese Yen (¥)', symbol: '¥' },
		{ value: 'AUD', name: 'AUD - Australian Dollar (A$)', symbol: 'A$' },
		{ value: 'CAD', name: 'CAD - Canadian Dollar (C$)', symbol: 'C$' },
		{ value: 'CHF', name: 'CHF - Swiss Franc (CHF)', symbol: 'CHF' },
		{ value: 'CNY', name: 'CNY - Chinese Yuan (¥)', symbol: '¥' },
		{ value: 'SEK', name: 'SEK - Swedish Krona (kr)', symbol: 'kr' },
		{ value: 'NZD', name: 'NZD - New Zealand Dollar (NZ$)', symbol: 'NZ$' },
		{ value: 'IDR', name: 'IDR - Indonesian Rupiah (Rp)', symbol: 'Rp' },
		{ value: 'SGD', name: 'SGD - Singapore Dollar (S$)', symbol: 'S$' },
		{ value: 'INR', name: 'INR - Indian Rupee (₹)', symbol: '₹' }
	];

	// Date format options
	const dateFormats = [
		{ value: 'MM/DD/YYYY', label: 'MM/DD/YYYY (US)', example: '12/31/2025' },
		{ value: 'DD/MM/YYYY', label: 'DD/MM/YYYY (EU)', example: '31/12/2025' },
		{ value: 'YYYY-MM-DD', label: 'YYYY-MM-DD (ISO)', example: '2025-12-31' },
		{ value: 'DD.MM.YYYY', label: 'DD.MM.YYYY', example: '31.12.2025' }
	];
</script>

<div
	class="fixed inset-0 bg-slate-900/95 backdrop-blur-sm z-50 flex items-center justify-center p-4"
>
	<div
		class="max-w-3xl w-full max-h-[90vh] flex flex-col bg-slate-800 rounded-2xl shadow-2xl border border-slate-700"
	>
		<!-- Progress Indicator -->
		<div class="p-8 border-b border-slate-700/50">
			<!-- Step Progress Bar -->
			<div class="relative">
				<!-- Background Progress Line -->
				<div class="absolute top-6 left-0 right-0 h-1 bg-slate-700/50 rounded-full"></div>

				<!-- Active Progress Line -->
				<div
					class="absolute top-6 left-0 h-1 bg-gradient-to-r from-blue-600 to-blue-500 rounded-full transition-all duration-500 ease-out"
					style="width: {(currentStep / (steps.length - 1)) * 100}%"
				></div>

				<!-- Step Circles -->
				<div class="relative flex justify-between">
					{#each steps as step, index}
						<div class="flex flex-col items-center">
							<!-- Circle -->
							<div
								class="relative z-10 flex items-center justify-center w-12 h-12 rounded-full transition-all duration-300 {index <=
								currentStep
									? index === currentStep
										? 'bg-blue-600 shadow-lg shadow-blue-500/50 scale-110'
										: 'bg-blue-600'
									: 'bg-slate-700'}"
							>
								<div
									class="absolute inset-0 rounded-full transition-all duration-300 {index ===
									currentStep
										? 'animate-pulse bg-blue-500/30'
										: ''}"
								></div>
								{#if index < currentStep}
									<CheckCircleSolid class="w-6 h-6 text-white relative z-10" />
								{:else if index === currentStep}
									<div class="relative z-10">
										<div
											class="w-3 h-3 bg-white rounded-full animate-pulse shadow-lg shadow-white/50"
										></div>
									</div>
								{:else}
									<span class="text-sm font-bold text-slate-400 relative z-10">{index + 1}</span>
								{/if}
							</div>

							<!-- Label -->
							<div class="mt-4 text-center">
								<p
									class="text-sm font-medium transition-colors duration-300 {index <= currentStep
										? 'text-white'
										: 'text-slate-500'}"
								>
									{step.label}
								</p>
								{#if index === currentStep}
									<div class="mt-1 flex justify-center">
										<div class="h-1 w-8 bg-blue-500 rounded-full"></div>
									</div>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>

		<!-- Content Area -->
		<div class="p-8 min-h-[200px] overflow-y-auto flex-1">
			{#if currentStep === 0}
				<!-- Step 1: Welcome -->
				<div class="text-center space-y-6">
					<!-- Free Trial Badge -->
					<div class="flex justify-center">
						<div
							class="inline-flex items-center gap-2 px-4 py-2 bg-green-500/10 border border-green-500/30 rounded-full"
						>
							<span class="text-lg">🎉</span>
							<span class="text-sm font-semibold text-green-400"
								>7-Day Free Trial • No Credit Card Required</span
							>
						</div>
					</div>

					<div class="flex justify-center">
						<div class="bg-blue-500/10 p-6 rounded-full">
							<RocketSolid class="w-16 h-16 text-blue-400" />
						</div>
					</div>
					<div>
						<h2 class="text-3xl font-bold text-white mb-3">Welcome to FacturMe!</h2>
						<p class="text-lg text-slate-300">Let's get you set up in just a few quick steps.</p>
					</div>
					<div class="bg-slate-700/50 rounded-lg p-6 text-left space-y-3">
						<div class="flex items-start gap-3">
							<CheckCircleSolid class="w-5 h-5 text-green-400 mt-0.5 flex-shrink-0" />
							<div>
								<p class="text-white font-medium">Track your time efficiently</p>
								<p class="text-sm text-slate-400">
									Log hours with our intuitive calendar interface
								</p>
							</div>
						</div>
						<div class="flex items-start gap-3">
							<CheckCircleSolid class="w-5 h-5 text-green-400 mt-0.5 flex-shrink-0" />
							<div>
								<p class="text-white font-medium">Manage multiple clients</p>
								<p class="text-sm text-slate-400">Keep all your client information organized</p>
							</div>
						</div>
						<div class="flex items-start gap-3">
							<CheckCircleSolid class="w-5 h-5 text-green-400 mt-0.5 flex-shrink-0" />
							<div>
								<p class="text-white font-medium">Generate invoices instantly</p>
								<p class="text-sm text-slate-400">Create professional invoices in seconds</p>
							</div>
						</div>
					</div>
				</div>
			{:else if currentStep === 1}
				<!-- Step 2: Currency Settings -->
				<div class="space-y-6">
					<div class="text-center mb-8">
						<div class="flex justify-center mb-4">
							<div class="bg-purple-500/10 p-4 rounded-full">
								<GlobeSolid class="w-12 h-12 text-purple-400" />
							</div>
						</div>
						<h2 class="text-2xl font-bold text-white mb-2">Main Currency</h2>
						<p class="text-slate-400">Your primary currency for dashboard and analytics</p>
					</div>

					<form class="space-y-5">
						<!-- Currency Selection -->
						<div class="bg-slate-700/50 rounded-lg p-5">
							<Label for="currency" class="text-white mb-3 text-lg flex items-center gap-2">
								<span>💰</span> Main Currency
							</Label>
							<DarkSelect
								id="currency"
								bind:value={localeSettings.currency}
								size="lg"
								items={currencies.map((c) => ({ value: c.value, name: c.name }))}
							/>
							<p class="text-xs text-slate-400 mt-3 leading-relaxed">
								This is your main currency. All client rates, invoices, and time entries will be converted to this currency for dashboard and analytics purposes.
							</p>
						</div>

						<!-- Preview Box -->
						<div class="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
							<p class="text-sm font-semibold text-blue-300 mb-2">Preview</p>
							<div class="space-y-1 text-sm">
								<p class="text-slate-300">
									<span class="text-slate-400">Amount:</span>
									<span class="font-bold text-white ml-2">
										{currencies.find((c) => c.value === localeSettings.currency)?.symbol}1,234.50
									</span>
								</p>
							</div>
						</div>
					</form>
				</div>
			{:else if currentStep === 2}
				<!-- Step 3: Add First Client -->
				<div class="space-y-6">
					<div class="text-center mb-8">
						<div class="flex justify-center mb-4">
							<div class="bg-green-500/10 p-4 rounded-full">
								<UserAddSolid class="w-12 h-12 text-green-400" />
							</div>
						</div>
						<h2 class="text-2xl font-bold text-white mb-2">Add Your First Client</h2>
						<p class="text-slate-400">Start by adding a client you work with</p>
					</div>

					<form class="space-y-4">
						<div>
							<Label for="client-name" class="text-white mb-2">Client Name *</Label>
							<DarkInput
								id="client-name"
								type="text"
								bind:value={clientData.name}
								placeholder="John Doe"
								required
								size="lg"
							/>
						</div>

						<div>
							<Label for="client-email" class="text-white mb-2">Email Address *</Label>
							<DarkInput
								id="client-email"
								type="email"
								bind:value={clientData.email}
								placeholder="john@example.com"
								required
								size="lg"
							/>
						</div>

						<div>
							<Label for="client-company" class="text-white mb-2">Company (Optional)</Label>
							<DarkInput
								id="client-company"
								type="text"
								bind:value={clientData.company}
								placeholder="Acme Corp"
								size="lg"
							/>
						</div>

						<div>
							<Label for="client-rate" class="text-white mb-2"
								>Hourly Rate ({currencies.find((c) => c.value === localeSettings.currency)?.symbol})
								*</Label
							>
							<DarkInput
								id="client-rate"
								type="number"
								bind:value={clientData.hourly_rate}
								placeholder="50"
								min="1"
								step="5"
								required
								size="lg"
							/>
						</div>
					</form>
				</div>
			{:else if currentStep === 3}
				<!-- Step 4: Demo Data -->
				<div class="space-y-6">
					<div class="text-center mb-8">
						<div class="flex justify-center mb-4">
							<div class="bg-purple-500/10 p-4 rounded-full">
								<span class="text-5xl">🎭</span>
							</div>
						</div>
						<h2 class="text-2xl font-bold text-white mb-2">Explore with Demo Data?</h2>
						<p class="text-slate-400">We can create sample data to help you explore features</p>
					</div>

					<!-- Toggle Card -->
					<div class="bg-slate-700/50 rounded-lg p-6 space-y-4">
						<label class="flex items-start gap-4 cursor-pointer group">
							<input
								type="checkbox"
								bind:checked={loadDemoData}
								class="mt-1 w-5 h-5 rounded border-gray-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-gray-800"
							/>
							<div class="flex-1">
								<div
									class="text-lg font-semibold text-white group-hover:text-blue-400 transition-colors"
								>
									Load sample data
								</div>
								<p class="text-sm text-slate-400 mt-1">
									Get started quickly with pre-populated clients, time entries, and invoices
								</p>
							</div>
						</label>

						{#if loadDemoData}
							<div
								class="ml-9 pl-4 border-l-2 border-blue-500/30 space-y-3 animate-in fade-in duration-300"
							>
								<div class="text-sm text-slate-300">
									<div class="flex items-center gap-2 mb-2">
										<span class="text-green-400">✓</span>
										<span class="font-medium">We'll create:</span>
									</div>
									<ul class="space-y-2 ml-6">
										<li class="flex items-start gap-2">
											<span class="text-blue-400 mt-1">•</span>
											<span
												><strong class="text-white">3 demo clients</strong> - Various business types</span
											>
										</li>
										<li class="flex items-start gap-2">
											<span class="text-blue-400 mt-1">•</span>
											<span
												><strong class="text-white">50 time entries</strong> - Sample work from the past
												2 months</span
											>
										</li>
										<li class="flex items-start gap-2">
											<span class="text-blue-400 mt-1">•</span>
											<span
												><strong class="text-white">5 invoices</strong> - Mix of draft, sent, and paid
												statuses</span
											>
										</li>
									</ul>
								</div>

								<div class="bg-blue-500/10 border border-blue-500/20 rounded p-3">
									<p class="text-xs text-blue-300 flex items-start gap-2">
										<span class="text-base">ℹ️</span>
										<span
											>Demo data is marked with a 🎭 badge and can be easily removed from settings
											anytime.</span
										>
									</p>
								</div>
							</div>
						{/if}
					</div>

					{#if !loadDemoData}
						<div class="bg-slate-700/30 border border-slate-600/50 rounded-lg p-4">
							<p class="text-sm text-slate-400 text-center">
								No problem! You can start fresh and add your own data as you go.
							</p>
						</div>
					{/if}
				</div>
			{:else if currentStep === 4}
				<!-- Step 5: Complete -->
				<div class="text-center space-y-6">
					<div class="flex justify-center">
						<div class="bg-green-500/10 p-6 rounded-full animate-pulse">
							<CheckCircleSolid class="w-16 h-16 text-green-400" />
						</div>
					</div>
					<div>
						<h2 class="text-3xl font-bold text-white mb-3">You're All Set!</h2>
						<p class="text-lg text-slate-300">
							Your workspace is ready. Let's start tracking time!
						</p>
					</div>

					<div class="bg-slate-700/50 rounded-lg p-6 text-left space-y-4">
						<h3 class="text-lg font-semibold text-white">What you've set up:</h3>
						<div class="space-y-3">
							{#if !skipFirstClient}
								<div class="flex items-center justify-between p-3 bg-slate-800 rounded">
									<span class="text-slate-300">First Client</span>
									<span class="text-white font-medium">{clientData.name}</span>
								</div>
								<div class="flex items-center justify-between p-3 bg-slate-800 rounded">
									<span class="text-slate-300">Client Rate</span>
									<span class="text-white font-medium">
										{currencies.find((c) => c.value === localeSettings.currency)
											?.symbol}{clientData.hourly_rate}/hour
									</span>
								</div>
							{/if}
							<div class="flex items-center justify-between p-3 bg-slate-800 rounded">
								<span class="text-slate-300">Currency</span>
								<span class="text-white font-medium">
									{currencies.find((c) => c.value === localeSettings.currency)?.name}
								</span>
							</div>
							{#if loadDemoData}
								<div
									class="flex items-center justify-between p-3 bg-purple-500/20 border border-purple-500/30 rounded"
								>
									<span class="text-slate-300">Demo Data</span>
									<span class="text-purple-300 font-medium flex items-center gap-2">
										<span>🎭</span> Enabled
									</span>
								</div>
							{/if}
						</div>
					</div>

					<div class="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
						<p class="text-sm text-blue-300">
							🎉 <strong>Next steps:</strong>
							{#if loadDemoData}
								Explore the demo data on your dashboard, then head to the calendar to log real time
								entries!
							{:else}
								Head to the calendar to log your first time entry, or add more clients from the
								Clients page.
							{/if}
						</p>
					</div>
				</div>
			{/if}
		</div>

		<!-- Navigation Footer -->
		<div class="p-6 border-t border-slate-700">
			{#if error}
				<div class="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
					<p class="text-sm text-red-400">{error}</p>
				</div>
			{/if}

			<div class="flex justify-between items-center">
				<div>
					{#if currentStep > 0}
						<Button
							size="lg"
							onclick={prevStep}
							disabled={isLoading}
							class="bg-slate-700 hover:bg-slate-600 text-white border-slate-600"
						>
							Back
						</Button>
					{:else}
						<div class="text-sm text-slate-500">Step {currentStep + 1} of {steps.length}</div>
					{/if}
				</div>

				<div class="flex items-center gap-3">
					{#if currentStep === 0}
						<div class="text-sm text-slate-400 mr-2">Takes less than 2 minutes</div>
					{/if}

					{#if currentStep === 2 && !skipFirstClient}
						<!-- Skip button for first client step -->
						<Button
							size="lg"
							onclick={skipClient}
							disabled={isLoading}
							class="bg-slate-700 hover:bg-slate-600 text-white border-slate-600"
						>
							Skip for now
						</Button>
					{/if}

					{#if currentStep < steps.length - 1}
						<Button
							color="blue"
							size="lg"
							onclick={nextStep}
							disabled={!canProceedFromStep(currentStep) || isLoading}
							class="bg-blue-600 hover:bg-blue-700 text-white"
						>
							{currentStep === 0 ? "Let's Get Started" : 'Continue'}
						</Button>
					{:else}
						<Button
							size="lg"
							onclick={handleComplete}
							disabled={isLoading}
							class="bg-green-600 hover:bg-green-700 text-white"
						>
							{#if isLoading}
								<svg
									class="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
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
								Setting up...
							{:else}
								Start Using FacturMe
							{/if}
						</Button>
					{/if}
				</div>
			</div>
		</div>
	</div>
</div>

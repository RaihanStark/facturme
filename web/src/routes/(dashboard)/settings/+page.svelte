<script lang="ts">
	import { Button, Input, Label, Select } from 'flowbite-svelte';
	import { api } from '$lib/api';
	import { authStore } from '$lib/stores';
	import { LockSolid, CheckCircleSolid, ExclamationCircleSolid, UserSolid } from 'flowbite-svelte-icons';

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let error = $state('');
	let success = $state('');
	let validationError = $state('');
	let selectedCurrency = $state($authStore.user?.currency || 'USD');
	let isCurrencyLoading = $state(false);
	let currencySuccess = $state('');
	let currencyError = $state('');
	let isResendingVerification = $state(false);
	let verificationSuccess = $state('');
	let verificationError = $state('');

	const currencies = [
		{ value: 'USD', name: 'USD - US Dollar ($)' },
		{ value: 'EUR', name: 'EUR - Euro (€)' },
		{ value: 'GBP', name: 'GBP - British Pound (£)' },
		{ value: 'JPY', name: 'JPY - Japanese Yen (¥)' },
		{ value: 'AUD', name: 'AUD - Australian Dollar (A$)' },
		{ value: 'CAD', name: 'CAD - Canadian Dollar (C$)' },
		{ value: 'CHF', name: 'CHF - Swiss Franc (CHF)' },
		{ value: 'CNY', name: 'CNY - Chinese Yuan (¥)' },
		{ value: 'SEK', name: 'SEK - Swedish Krona (kr)' },
		{ value: 'NZD', name: 'NZD - New Zealand Dollar (NZ$)' },
		{ value: 'IDR', name: 'IDR - Indonesian Rupiah (Rp)' },
		{ value: 'SGD', name: 'SGD - Singapore Dollar (S$)' },
		{ value: 'INR', name: 'INR - Indian Rupee (₹)' }
	];

	function validatePasswords() {
		validationError = '';

		if (newPassword.length < 8) {
			validationError = 'New password must be at least 8 characters long';
			return false;
		}

		if (newPassword !== confirmPassword) {
			validationError = 'Passwords do not match';
			return false;
		}

		if (currentPassword === newPassword) {
			validationError = 'New password must be different from current password';
			return false;
		}

		return true;
	}

	async function handleChangePassword(e: Event) {
		e.preventDefault();

		if (!validatePasswords()) {
			return;
		}

		isLoading = true;
		error = '';
		success = '';

		try {
			const response = await api.changePassword(currentPassword, newPassword);
			success = response.message;
			// Clear form
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
			// Clear success message after 5 seconds
			setTimeout(() => {
				success = '';
			}, 5000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to change password. Please try again.';
		} finally {
			isLoading = false;
		}
	}

	async function handleCurrencyChange() {
		if (selectedCurrency === $authStore.user?.currency) return;

		isCurrencyLoading = true;
		currencyError = '';
		currencySuccess = '';

		try {
			const updatedUser = await api.updateCurrency(selectedCurrency);
			authStore.setUser(updatedUser);
			currencySuccess = 'Currency updated successfully';
			setTimeout(() => {
				currencySuccess = '';
			}, 5000);
		} catch (err) {
			currencyError = err instanceof Error ? err.message : 'Failed to update currency';
			selectedCurrency = $authStore.user?.currency || 'USD';
		} finally {
			isCurrencyLoading = false;
		}
	}

	async function handleResendVerification() {
		isResendingVerification = true;
		verificationError = '';
		verificationSuccess = '';

		try {
			await api.resendVerificationEmail();
			verificationSuccess = 'Verification email sent! Please check your inbox.';
			setTimeout(() => {
				verificationSuccess = '';
			}, 5000);
		} catch (err) {
			verificationError = err instanceof Error ? err.message : 'Failed to send verification email';
		} finally {
			isResendingVerification = false;
		}
	}
</script>

<svelte:head>
	<title>FacturMe - Settings</title>
</svelte:head>

<div class="max-w-7xl">
	<!-- Page Header -->
	<div class="mb-8">
		<h1 class="text-3xl font-bold text-white mb-2">Settings</h1>
		<p class="text-slate-400">Manage your account settings and preferences</p>
	</div>

	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Account Information -->
		<div class="lg:col-span-1">
			<div class="bg-slate-800 rounded-xl border border-slate-700 shadow-xl p-6">
				<div class="flex items-center gap-3 mb-6">
					<div class="rounded-lg bg-blue-500/20 p-3">
						<UserSolid class="w-6 h-6 text-blue-400" />
					</div>
					<h2 class="text-xl font-semibold text-white">Account Information</h2>
				</div>

				<div class="space-y-5">
					<div class="pb-5 border-b border-slate-700">
						<p class="text-xs font-medium text-slate-400 uppercase tracking-wider mb-2">Name</p>
						<p class="text-white font-medium text-lg">{$authStore.user?.name}</p>
					</div>
					<div class="pb-5 border-b border-slate-700">
						<p class="text-xs font-medium text-slate-400 uppercase tracking-wider mb-2">Email</p>
						<div class="flex items-center gap-2 flex-wrap">
							<p class="text-white font-medium break-all">{$authStore.user?.email}</p>
							{#if $authStore.user?.email_verified}
								<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-green-900/50 border border-green-700 text-green-300 text-xs font-medium">
									<CheckCircleSolid class="w-3.5 h-3.5" />
									Verified
								</span>
							{:else}
								<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-yellow-900/50 border border-yellow-700 text-yellow-300 text-xs font-medium">
									<ExclamationCircleSolid class="w-3.5 h-3.5" />
									Not Verified
								</span>
							{/if}
						</div>
						{#if !$authStore.user?.email_verified}
							<div class="mt-4">
								<!-- Success Message -->
								{#if verificationSuccess}
									<div class="bg-green-900/50 border border-green-700 text-green-200 px-4 py-3 rounded-lg mb-3 flex items-start gap-2 text-sm">
										<CheckCircleSolid class="w-4 h-4 flex-shrink-0 mt-0.5" />
										<p>{verificationSuccess}</p>
									</div>
								{/if}

								<!-- Error Message -->
								{#if verificationError}
									<div class="bg-red-900/50 border border-red-700 text-red-200 px-4 py-3 rounded-lg mb-3 flex items-start gap-2 text-sm">
										<ExclamationCircleSolid class="w-4 h-4 flex-shrink-0 mt-0.5" />
										<p>{verificationError}</p>
									</div>
								{/if}

								<p class="text-sm text-slate-400 mb-3">
									Please verify your email address to ensure account security and enable password recovery.
								</p>
								<Button
									size="sm"
									color="alternative"
									onclick={handleResendVerification}
									disabled={isResendingVerification}
									class="bg-slate-700 hover:bg-slate-600 border-slate-600 text-white"
								>
									{#if isResendingVerification}
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
										Sending...
									{:else}
										Resend Verification Email
									{/if}
								</Button>
							</div>
						{/if}
					</div>
				</div>
			</div>
			<!-- Currency Preferences -->
			<div class="bg-slate-800 rounded-xl border border-slate-700 shadow-xl p-6 mt-6">
				<div class="flex items-center gap-3 mb-6">
					<div class="rounded-lg bg-green-500/20 p-3">
						<svg class="w-6 h-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
					</div>
					<div>
						<h2 class="text-xl font-semibold text-white">Main Currency</h2>
						<p class="text-sm text-slate-400 mt-1">Your primary currency for dashboard and analytics</p>
					</div>
				</div>

				<!-- Success Message -->
				{#if currencySuccess}
					<div
						class="bg-green-900/50 border border-green-700 text-green-200 px-5 py-4 rounded-xl mb-6 flex items-start gap-3"
					>
						<CheckCircleSolid class="w-5 h-5 flex-shrink-0 mt-0.5" />
						<p class="font-medium">{currencySuccess}</p>
					</div>
				{/if}

				<!-- Error Message -->
				{#if currencyError}
					<div
						class="bg-red-900/50 border border-red-700 text-red-200 px-5 py-4 rounded-xl mb-6 flex items-start gap-3"
					>
						<ExclamationCircleSolid class="w-5 h-5 flex-shrink-0 mt-0.5" />
						<p class="font-medium">{currencyError}</p>
					</div>
				{/if}

				<div class="space-y-4">
					<div>
						<Label for="currency" class="mb-2 text-white font-medium">Select Currency</Label>
						<Select
							id="currency"
							bind:value={selectedCurrency}
							onchange={handleCurrencyChange}
							disabled={isCurrencyLoading}
							class="bg-slate-700/50 border-slate-600 text-white focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
						>
							{#each currencies as currency}
								<option value={currency.value}>{currency.name}</option>
							{/each}
						</Select>
					</div>
					<p class="text-xs text-slate-400 flex items-start gap-1.5 leading-relaxed">
						<svg class="w-4 h-4 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
						<span>This is your main currency. All client rates, invoices, and time entries will be converted to this currency for dashboard and analytics purposes.</span>
					</p>
				</div>
			</div>
		</div>

		<!-- Change Password -->
		<div class="lg:col-span-2">
			<div class="bg-slate-800 rounded-xl border border-slate-700 shadow-xl p-6">
				<div class="flex items-center gap-3 mb-6">
					<div class="rounded-lg bg-primary-500/20 p-3">
						<LockSolid class="w-6 h-6 text-primary-400" />
					</div>
					<div>
						<h2 class="text-xl font-semibold text-white">Change Password</h2>
						<p class="text-sm text-slate-400 mt-1">
							Update your password to keep your account secure
						</p>
					</div>
				</div>

				<!-- Success Message -->
				{#if success}
					<div
						class="bg-green-900/50 border border-green-700 text-green-200 px-5 py-4 rounded-xl mb-6 flex items-start gap-3"
					>
						<CheckCircleSolid class="w-5 h-5 flex-shrink-0 mt-0.5" />
						<p class="font-medium">{success}</p>
					</div>
				{/if}

				<!-- Error Messages -->
				{#if error}
					<div
						class="bg-red-900/50 border border-red-700 text-red-200 px-5 py-4 rounded-xl mb-6 flex items-start gap-3"
					>
						<ExclamationCircleSolid class="w-5 h-5 flex-shrink-0 mt-0.5" />
						<p class="font-medium">{error}</p>
					</div>
				{/if}

				{#if validationError}
					<div
						class="bg-yellow-900/50 border border-yellow-700 text-yellow-200 px-5 py-4 rounded-xl mb-6 flex items-start gap-3"
					>
						<ExclamationCircleSolid class="w-5 h-5 flex-shrink-0 mt-0.5" />
						<p class="font-medium">{validationError}</p>
					</div>
				{/if}

				<!-- Change Password Form -->
				<form onsubmit={handleChangePassword} class="space-y-6">
					<!-- Current Password -->
					<div>
						<Label for="currentPassword" class="mb-2 text-white font-medium"
							>Current Password</Label
						>
						<Input
							id="currentPassword"
							type="password"
							bind:value={currentPassword}
							placeholder="Enter your current password"
							required
							autocomplete="current-password"
							class="bg-slate-700/50 border-slate-600 text-white placeholder-slate-400 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 h-12"
							oninput={() => {
								error = '';
								validationError = '';
							}}
						/>
					</div>

					<div class="border-t border-slate-700 pt-6">
						<!-- New Password -->
						<div class="mb-5">
							<Label for="newPassword" class="mb-2 text-white font-medium">New Password</Label>
							<Input
								id="newPassword"
								type="password"
								bind:value={newPassword}
								placeholder="Enter your new password"
								required
								autocomplete="new-password"
								class="bg-slate-700/50 border-slate-600 text-white placeholder-slate-400 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 h-12"
								oninput={() => {
									error = '';
									validationError = '';
								}}
							/>
							<p class="mt-2 text-xs text-slate-400 flex items-center gap-1">
								<svg
									class="w-4 h-4"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
									/>
								</svg>
								Must be at least 8 characters long
							</p>
						</div>

						<!-- Confirm New Password -->
						<div>
							<Label for="confirmPassword" class="mb-2 text-white font-medium"
								>Confirm New Password</Label
							>
							<Input
								id="confirmPassword"
								type="password"
								bind:value={confirmPassword}
								placeholder="Confirm your new password"
								required
								autocomplete="new-password"
								class="bg-slate-700/50 border-slate-600 text-white placeholder-slate-400 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 h-12"
								oninput={() => {
									error = '';
									validationError = '';
								}}
							/>
						</div>
					</div>

					<!-- Submit Button -->
					<div class="flex items-center gap-3 pt-4">
						<Button type="submit" color="primary" size="lg" class="px-8" disabled={isLoading}>
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
								Changing Password...
							{:else}
								Change Password
							{/if}
						</Button>
					</div>
				</form>
			</div>
		</div>
	</div>
</div>

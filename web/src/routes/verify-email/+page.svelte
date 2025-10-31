<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { authStore } from '$lib/stores';
	import logo from '$lib/assets/logo-white.png';
	import { onMount } from 'svelte';
	import {
		EnvelopeSolid,
		CheckCircleSolid,
		ExclamationCircleSolid,
		PaperPlaneOutline
	} from 'flowbite-svelte-icons';

	let isLoading = $state(false);
	let error = $state('');
	let success = $state('');
	let isResending = $state(false);
	let resendSuccess = $state('');
	let resendError = $state('');
	let canResend = $state(false);
	let resendCooldown = $state(60);

	// Check URL params for token and auto-verify
	onMount(() => {
		const urlParams = new URLSearchParams(window.location.search);
		const urlToken = urlParams.get('token');
		if (urlToken) {
			handleVerify(urlToken);
		}

		// Enable resend button after 60 seconds
		const timer = setTimeout(() => {
			canResend = true;
		}, 60000);

		// Countdown timer
		const countdown = setInterval(() => {
			resendCooldown = resendCooldown - 1;
			if (resendCooldown <= 0) {
				clearInterval(countdown);
			}
		}, 1000);

		return () => {
			clearTimeout(timer);
			clearInterval(countdown);
		};
	});

	async function handleVerify(token: string) {
		isLoading = true;
		error = '';
		success = '';

		try {
			const response = await api.verifyEmail({ token: token.trim() });
			authStore.setUser(response);
			success = 'Email verified successfully! Redirecting...';
			// Redirect to dashboard/onboarding after a short delay
			setTimeout(() => {
				goto('/');
			}, 1500);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Verification failed. Please try again.';
			isLoading = false;
		}
	}

	async function handleResend() {
		isResending = true;
		resendError = '';
		resendSuccess = '';

		try {
			const response = await api.resendVerificationEmail();
			resendSuccess = response.message;
			// Reset cooldown
			canResend = false;
			resendCooldown = 60;
			setTimeout(() => {
				canResend = true;
			}, 60000);
		} catch (err) {
			resendError = err instanceof Error ? err.message : 'Failed to resend email. Please try again.';
		} finally {
			isResending = false;
		}
	}

	function handleLogout() {
		api.logout();
		authStore.clearAuth();
		goto('/login');
	}
</script>

<svelte:head>
	<title>FacturMe - Verify Email</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 via-slate-900 to-slate-800 px-4 py-12">
	<div class="w-full max-w-lg">
		<!-- Logo/Brand Section -->
		<div class="text-center mb-8">
			<div class="flex justify-center mb-6">
				<img src={logo} alt="FacturMe" class="h-16" />
			</div>
		</div>

		<!-- Verification Card -->
		<div class="bg-slate-800 rounded-2xl border border-slate-700 shadow-2xl overflow-hidden">
			{#if success}
				<!-- Success State -->
				<div class="p-8 text-center">
					<div class="flex justify-center mb-6">
						<div class="rounded-full bg-green-500/20 p-4">
							<CheckCircleSolid class="w-16 h-16 text-green-400" />
						</div>
					</div>
					<h2 class="text-2xl font-bold text-white mb-3">Email Verified!</h2>
					<p class="text-slate-300 mb-6">{success}</p>
					<div class="animate-pulse">
						<div class="h-1 bg-gradient-to-r from-primary-600 to-primary-400 rounded-full"></div>
					</div>
				</div>
			{:else if isLoading}
				<!-- Loading State -->
				<div class="p-8 text-center">
					<div class="flex justify-center mb-6">
						<div class="rounded-full bg-primary-500/20 p-4">
							<svg
								class="animate-spin h-16 w-16 text-primary-400"
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
						</div>
					</div>
					<h2 class="text-2xl font-bold text-white mb-3">Verifying Your Email...</h2>
					<p class="text-slate-300">Please wait a moment</p>
				</div>
			{:else if error}
				<!-- Error State -->
				<div class="p-8">
					<div class="text-center mb-6">
						<div class="flex justify-center mb-4">
							<div class="rounded-full bg-red-500/20 p-4">
								<ExclamationCircleSolid class="w-12 h-12 text-red-400" />
							</div>
						</div>
						<h1 class="text-3xl font-bold text-white mb-2">Verification Failed</h1>
						<p class="text-slate-400">We couldn't verify your email address</p>
					</div>

					<!-- Error Message -->
					<div class="bg-red-900/30 border border-red-700/50 rounded-lg p-4 mb-6">
						<p class="text-red-200 text-sm">{error}</p>
					</div>

					<!-- Help Text -->
					<div class="bg-slate-700/50 rounded-lg p-4 mb-6">
						<p class="text-slate-300 text-sm leading-relaxed mb-3">
							This could happen if:
						</p>
						<ul class="text-slate-300 text-sm space-y-2 list-disc list-inside">
							<li>The verification link has expired (links expire after 24 hours)</li>
							<li>The link has already been used</li>
							<li>The verification token is invalid</li>
						</ul>
					</div>

					<!-- Action Buttons -->
					<div class="space-y-3">
						<button
							type="button"
							onclick={() => goto('/register')}
							class="w-full px-4 py-3 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
						>
							Register Again
						</button>
						<button
							type="button"
							onclick={handleLogout}
							class="w-full px-4 py-3 bg-slate-700 hover:bg-slate-600 text-white font-medium rounded-lg transition-colors"
						>
							Sign Out
						</button>
					</div>
				</div>
			{:else}
				<!-- Waiting for Email Click -->
				<div class="p-8">
					<!-- Icon and Title -->
					<div class="text-center mb-6">
						<div class="flex justify-center mb-4">
							<div class="rounded-full bg-primary-500/20 p-4">
								<EnvelopeSolid class="w-12 h-12 text-primary-400" />
							</div>
						</div>
						<h1 class="text-3xl font-bold text-white mb-2">Check Your Email</h1>
						<p class="text-slate-400">We've sent a verification link to</p>
						<p class="text-primary-400 font-semibold mt-1 break-all px-4">
							{$authStore.user?.email}
						</p>
					</div>

					<!-- Instructions -->
					<div class="bg-slate-700/50 rounded-lg p-4 mb-6">
						<p class="text-slate-300 text-sm leading-relaxed">
							Click the <strong>Verify Email Address</strong> button in the email to activate your
							account. The link will expire in 24 hours.
						</p>
					</div>

					<!-- Resend Success Message -->
					{#if resendSuccess}
						<div class="bg-green-900/30 border border-green-700/50 rounded-lg p-4 mb-6 flex items-start gap-3">
							<CheckCircleSolid class="w-5 h-5 text-green-400 flex-shrink-0 mt-0.5" />
							<p class="text-green-200 text-sm">{resendSuccess}</p>
						</div>
					{/if}

					<!-- Resend Error Message -->
					{#if resendError}
						<div class="bg-red-900/30 border border-red-700/50 rounded-lg p-4 mb-6 flex items-start gap-3">
							<ExclamationCircleSolid class="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
							<p class="text-red-200 text-sm">{resendError}</p>
						</div>
					{/if}

					<!-- Tips -->
					<div class="space-y-3 mb-6">
						<div class="flex items-start gap-3 text-sm">
							<div class="flex-shrink-0 w-6 h-6 rounded-full bg-primary-500/20 flex items-center justify-center mt-0.5">
								<span class="text-primary-400 text-xs font-bold">1</span>
							</div>
							<p class="text-slate-300">
								Check your <strong>inbox</strong> for an email from FacturMe
							</p>
						</div>
						<div class="flex items-start gap-3 text-sm">
							<div class="flex-shrink-0 w-6 h-6 rounded-full bg-primary-500/20 flex items-center justify-center mt-0.5">
								<span class="text-primary-400 text-xs font-bold">2</span>
							</div>
							<p class="text-slate-300">
								Can't find it? Check your <strong>spam or junk folder</strong>
							</p>
						</div>
						<div class="flex items-start gap-3 text-sm">
							<div class="flex-shrink-0 w-6 h-6 rounded-full bg-primary-500/20 flex items-center justify-center mt-0.5">
								<span class="text-primary-400 text-xs font-bold">3</span>
							</div>
							<p class="text-slate-300">
								Click the verification button in the email
							</p>
						</div>
					</div>

					<!-- Resend Button -->
					<div class="mb-6">
						<button
							type="button"
							onclick={handleResend}
							disabled={!canResend || isResending}
							class="w-full px-4 py-3 bg-slate-700 hover:bg-slate-600 text-white font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
						>
							{#if isResending}
								<svg
									class="animate-spin h-5 w-5"
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
							{:else if !canResend}
								<PaperPlaneOutline class="w-5 h-5" />
								Resend Email ({resendCooldown}s)
							{:else}
								<PaperPlaneOutline class="w-5 h-5" />
								Resend Verification Email
							{/if}
						</button>
					</div>
				</div>

				<!-- Footer Section -->
				<div class="bg-slate-900/50 border-t border-slate-700 px-8 py-4">
					<div class="flex flex-col sm:flex-row items-center justify-between gap-3 text-sm">
						<p class="text-slate-400">Wrong email address?</p>
						<button
							type="button"
							onclick={handleLogout}
							class="text-primary-400 hover:text-primary-300 font-medium transition-colors"
						>
							Sign out and try again
						</button>
					</div>
				</div>
			{/if}
		</div>

		<!-- Help Text -->
		{#if !success && !isLoading && !error}
			<div class="text-center mt-6">
				<p class="text-slate-500 text-sm">
					Still having trouble? Contact support at
					<a href="mailto:support@yourdomain.com" class="text-primary-400 hover:text-primary-300">
						support@yourdomain.com
					</a>
				</p>
			</div>
		{/if}
	</div>
</div>

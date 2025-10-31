<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import logo from '$lib/assets/logo-white.png';
	import { onMount } from 'svelte';
	import { Button, Input, Label } from 'flowbite-svelte';
	import { CheckCircleSolid, ExclamationCircleSolid, LockSolid } from 'flowbite-svelte-icons';

	let token = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let error = $state('');
	let success = $state(false);
	let validationError = $state('');

	// Get token from URL on mount
	onMount(() => {
		const urlParams = new URLSearchParams(window.location.search);
		const urlToken = urlParams.get('token');
		if (!urlToken) {
			error = 'Invalid or missing reset token. Please request a new password reset link.';
		} else {
			token = urlToken;
		}
	});

	function validatePasswords() {
		validationError = '';

		if (password.length < 8) {
			validationError = 'Password must be at least 8 characters long';
			return false;
		}

		if (password !== confirmPassword) {
			validationError = 'Passwords do not match';
			return false;
		}

		return true;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!validatePasswords()) {
			return;
		}

		isLoading = true;
		error = '';

		try {
			await api.resetPassword(token, password);
			success = true;
			// Redirect to login after 2 seconds
			setTimeout(() => {
				goto('/login');
			}, 2000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to reset password. Please try again.';
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>FacturMe - Reset Password</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-slate-900 px-4">
	<div class="w-full max-w-md">
		<!-- Logo/Brand Section -->
		<div class="text-center mb-8">
			<div class="flex justify-center mb-6">
				<img src={logo} alt="FacturMe" class="h-16" />
			</div>
			<h1 class="text-3xl font-bold text-white mb-2">Set New Password</h1>
			<p class="text-slate-400">Choose a strong password for your account</p>
		</div>

		<!-- Reset Password Card -->
		<div class="bg-slate-800 rounded-xl border border-slate-700 p-8">
			{#if success}
				<!-- Success State -->
				<div class="text-center space-y-4">
					<div class="flex justify-center">
						<div
							class="w-16 h-16 rounded-full bg-green-900/50 border border-green-700 flex items-center justify-center"
						>
							<CheckCircleSolid class="w-8 h-8 text-green-400" />
						</div>
					</div>
					<h2 class="text-xl font-semibold text-white">Password Reset Successful!</h2>
					<p class="text-slate-400">
						Your password has been updated successfully. Redirecting you to login...
					</p>
					<div class="animate-pulse pt-4">
						<div class="h-1 bg-gradient-to-r from-primary-600 to-primary-400 rounded-full"></div>
					</div>
				</div>
			{:else if error && !token}
				<!-- Invalid Token State -->
				<div class="text-center space-y-4">
					<div class="flex justify-center">
						<div
							class="w-16 h-16 rounded-full bg-red-900/50 border border-red-700 flex items-center justify-center"
						>
							<ExclamationCircleSolid class="w-8 h-8 text-red-400" />
						</div>
					</div>
					<h2 class="text-xl font-semibold text-white">Invalid Reset Link</h2>
					<p class="text-slate-400">{error}</p>
					<div class="pt-4 space-y-3">
						<Button
							color="primary"
							class="w-full"
							onclick={() => goto('/forgot-password')}
						>
							Request New Reset Link
						</Button>
						<Button
							color="alternative"
							class="w-full"
							onclick={() => goto('/login')}
						>
							Back to Login
						</Button>
					</div>
				</div>
			{:else}
				<!-- Form State -->
				<form onsubmit={handleSubmit} class="space-y-6">
					<!-- Error Messages -->
					{#if error}
						<div class="bg-red-900/50 border border-red-700 text-red-200 px-4 py-3 rounded-lg">
							{error}
						</div>
					{/if}

					{#if validationError}
						<div class="bg-yellow-900/50 border border-yellow-700 text-yellow-200 px-4 py-3 rounded-lg">
							{validationError}
						</div>
					{/if}

					<!-- Password Field -->
					<div>
						<Label for="password" class="mb-2 text-white">New Password</Label>
						<Input
							id="password"
							type="password"
							bind:value={password}
							placeholder="••••••••"
							required
							class="bg-slate-700 border-slate-600 text-white placeholder-slate-400"
							oninput={() => (validationError = '')}
						/>
						<p class="mt-2 text-sm text-slate-400">
							Must be at least 8 characters long
						</p>
					</div>

					<!-- Confirm Password Field -->
					<div>
						<Label for="confirmPassword" class="mb-2 text-white">Confirm New Password</Label>
						<Input
							id="confirmPassword"
							type="password"
							bind:value={confirmPassword}
							placeholder="••••••••"
							required
							class="bg-slate-700 border-slate-600 text-white placeholder-slate-400"
							oninput={() => (validationError = '')}
						/>
					</div>

					<!-- Submit Button -->
					<Button type="submit" color="primary" class="w-full" disabled={isLoading}>
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
							Resetting Password...
						{:else}
							<LockSolid class="w-5 h-5 mr-2 inline" />
							Reset Password
						{/if}
					</Button>

					<!-- Back to Login Link -->
					<div class="text-center">
						<a
							href="/login"
							class="text-slate-400 hover:text-white font-medium"
						>
							Back to Login
						</a>
					</div>
				</form>
			{/if}
		</div>
	</div>
</div>

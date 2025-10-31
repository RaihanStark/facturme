<script lang="ts">
	import { Button, Input, Label } from 'flowbite-svelte';
	import { api } from '$lib/api';
	import logo from '$lib/assets/logo-white.png';
	import { EnvelopeSolid } from 'flowbite-svelte-icons';

	let email = $state('');
	let isLoading = $state(false);
	let error = $state('');
	let success = $state(false);

	// Temporary: Disable forgot password while email system is in verification
	const forgotPasswordEnabled = false;

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isLoading = true;
		error = '';
		success = false;

		try {
			await api.forgotPassword(email);
			success = true;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to send reset email. Please try again.';
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>FacturMe - Forgot Password</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-slate-900 px-4">
	<div class="w-full max-w-md">
		<!-- Logo/Brand Section -->
		<div class="text-center mb-8">
			<div class="flex justify-center mb-6">
				<img src={logo} alt="FacturMe" class="h-16" />
			</div>
			<h1 class="text-3xl font-bold text-white mb-2">Reset Your Password</h1>
			<p class="text-slate-400">We'll send you a link to reset your password</p>
		</div>

		<!-- Forgot Password Card -->
		<div class="bg-slate-800 rounded-xl border border-slate-700 p-8">
			{#if !forgotPasswordEnabled}
				<!-- Forgot password temporarily disabled message -->
				<div class="bg-blue-900/30 border border-blue-700/50 rounded-lg p-4 text-center">
					<svg
						class="w-8 h-8 text-blue-400 mx-auto mb-3"
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
					<p class="text-sm text-slate-300 mb-2">
						Password reset via email is currently being set up.
					</p>
					<p class="text-xs text-slate-400 mb-4">
						Please contact support if you need to reset your password.
					</p>
					<a
						href="/login"
						class="inline-flex items-center gap-2 text-primary-500 hover:text-primary-400 font-medium"
					>
						← Back to Login
					</a>
				</div>
			{:else if success}
				<!-- Success State -->
				<div class="text-center space-y-4">
					<div class="flex justify-center">
						<div
							class="w-16 h-16 rounded-full bg-green-900/50 border border-green-700 flex items-center justify-center"
						>
							<EnvelopeSolid class="w-8 h-8 text-green-400" />
						</div>
					</div>
					<h2 class="text-xl font-semibold text-white">Check Your Email</h2>
					<p class="text-slate-400">
						We've sent a password reset link to <span class="text-white font-medium"
							>{email}</span
						>. The link will expire in 1 hour.
					</p>
					<div class="pt-4">
						<a
							href="/login"
							class="inline-flex items-center gap-2 text-primary-500 hover:text-primary-400 font-medium"
						>
							← Back to Login
						</a>
					</div>
				</div>
			{:else}
				<!-- Form State -->
				<form onsubmit={handleSubmit} class="space-y-6">
					<!-- Error Message -->
					{#if error}
						<div class="bg-red-900/50 border border-red-700 text-red-200 px-4 py-3 rounded-lg">
							{error}
						</div>
					{/if}

					<!-- Email Field -->
					<div>
						<Label for="email" class="mb-2 text-white">Email Address</Label>
						<Input
							id="email"
							type="email"
							bind:value={email}
							placeholder="you@example.com"
							required
							class="bg-slate-700 border-slate-600 text-white placeholder-slate-400"
						/>
						<p class="mt-2 text-sm text-slate-400">
							Enter the email address associated with your account
						</p>
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
							Sending Reset Link...
						{:else}
							Send Reset Link
						{/if}
					</Button>

					<!-- Back to Login Link -->
					<div class="text-center">
						<a
							href="/login"
							class="inline-flex items-center gap-2 text-slate-400 hover:text-white font-medium"
						>
							← Back to Login
						</a>
					</div>
				</form>
			{/if}
		</div>
	</div>
</div>

<script lang="ts">
	import { Button, Input, Label } from 'flowbite-svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { authStore } from '$lib/stores';
	import logo from '$lib/assets/logo-white.png';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let error = $state('');

	// Temporary: Disable email registration while email system is in verification
	const emailRegistrationEnabled = false;

	async function handleRegister(e: Event) {
		e.preventDefault();
		isLoading = true;
		error = '';

		// Validate passwords match
		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			isLoading = false;
			return;
		}

		// Validate password length
		if (password.length < 6) {
			error = 'Password must be at least 6 characters';
			isLoading = false;
			return;
		}

		try {
			const response = await api.register({ name, email, password });
			authStore.setAuth(response.user, response.token);

			// Success - redirect to dashboard
			goto('/');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Registration failed. Please try again.';
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>FacturMe - Register</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-slate-900 px-4">
	<div class="w-full max-w-md">
		<!-- Logo/Brand Section -->
		<div class="text-center mb-8">
			<div class="flex justify-center mb-6">
				<img src={logo} alt="FacturMe" class="h-16" />
			</div>
			<h1 class="text-3xl font-bold text-white mb-2">Create Account</h1>
			<p class="text-slate-400">Sign up to get started with FacturMe</p>
		</div>

		<!-- Register Card -->
		<div class="bg-slate-800 rounded-xl border border-slate-700 p-8">
			<form onsubmit={handleRegister} class="space-y-6">
				<!-- Error Message -->
				{#if error}
					<div class="bg-red-900/50 border border-red-700 text-red-200 px-4 py-3 rounded-lg">
						{error}
					</div>
				{/if}

				<!-- Name Field -->
				<div>
					<Label for="name" class="mb-2 text-white">Full Name</Label>
					<Input
						id="name"
						type="text"
						bind:value={name}
						placeholder="John Doe"
						required
						class="bg-slate-700 border-slate-600 text-white placeholder-slate-400"
					/>
				</div>

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
				</div>

				<!-- Password Field -->
				<div>
					<Label for="password" class="mb-2 text-white">Password</Label>
					<Input
						id="password"
						type="password"
						bind:value={password}
						placeholder="••••••••"
						required
						minlength={6}
						class="bg-slate-700 border-slate-600 text-white placeholder-slate-400"
					/>
					<p class="mt-1 text-xs text-slate-400">Minimum 6 characters</p>
				</div>

				<!-- Confirm Password Field -->
				<div>
					<Label for="confirmPassword" class="mb-2 text-white">Confirm Password</Label>
					<Input
						id="confirmPassword"
						type="password"
						bind:value={confirmPassword}
						placeholder="••••••••"
						required
						minlength={6}
						class="bg-slate-700 border-slate-600 text-white placeholder-slate-400"
					/>
				</div>

				<!-- Submit Button -->
				<Button
					type="submit"
					color="primary"
					class="w-full"
					disabled={isLoading}
				>
					{#if isLoading}
						<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
						Creating account...
					{:else}
						Create Account
					{/if}
				</Button>

				<!-- Login Link -->
				<div class="text-center">
					<p class="text-slate-400">
						Already have an account?
						<a href="/login" class="text-primary-500 hover:text-primary-400 font-medium">
							Sign in
						</a>
					</p>
				</div>
			</form>
		</div>
	</div>
</div>

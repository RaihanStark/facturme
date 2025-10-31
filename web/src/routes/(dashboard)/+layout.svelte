<script lang="ts">
	import '../../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import logo from '$lib/assets/logo-white.png';
	import {
		Sidebar,
		SidebarGroup,
		SidebarItem,
		Navbar,
		NavBrand,
		NavHamburger,
		Avatar,
		Dropdown,
		DropdownItem,
		DropdownHeader,
		DropdownGroup,
		Toast
	} from 'flowbite-svelte';
	import {
		ChartPieSolid,
		ClockSolid,
		UsersGroupSolid,
		FileInvoiceSolid,
		CalendarMonthSolid,
		CogSolid,
		ExclamationCircleSolid
	} from 'flowbite-svelte-icons';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import DemoBanner from '$lib/components/DemoBanner.svelte';
	import { TourGuideClient } from '@sjmc11/tourguidejs';

	let { children } = $props();
	let sidebarOpen = $state(false);
	let activeUrl = $state($page.url.pathname);
	let showToast = $state(false);
	let toastMessage = $state('');

	let tour: TourGuideClient | null = null;

	$effect(() => {
		// Only start tour if user has completed onboarding but not the tour
		if ($authStore.user?.onboarding_completed && !$authStore.user?.tour_completed) {
			// Use a small delay to ensure DOM is ready
			const timer = setTimeout(() => {
				// Initialize tour
				tour = new TourGuideClient({
					exitOnClickOutside: false,
					steps: [
						{
							title: 'Welcome to FacturMe',
							content: "Let's take a quick look at your dashboard.",
							order: 1,
							beforeEnter: () => {
								// Ensure we're on the dashboard page
								if ($page.url.pathname !== '/') {
									goto('/');
								}
							}
						},
						{
							title: 'Your Dashboard',
							content:
								'View your work hours, revenue, and invoices at a glance. Use the date filters to see different time periods.',
							order: 2,
							target: '#dashboard-content'
						},
						{
							title: 'Explore More Features',
							content:
								'Check out the Calendar view to visualize your time entries and plan your work schedule!',
							order: 3,
							target: '#calendar-link',
							afterLeave: () => {
								return goto('/calendar');
							}
						},
						{
							title: 'Calendar View',
							content: 'Visualize your work hours across the month with color-coded days.',
							order: 4
						},
						{
							title: 'Visual Time Tracking',
							content:
								'The calendar heatmap shows your work patterns at a glance. Track consistency, spot gaps, and visualize your productivity over the past 4 weeks.',
							order: 5,
							target: '#calendar-grid'
						},
						{
							title: 'Color-Coded Productivity',
							content:
								'The color legend helps you track your work patterns: Red for days with no hours, yellow for light work (<5h), green for productive days (≥5h), and gray for future dates.',
							order: 6,
							target: '#calendar-legend'
						},
						{
							title: 'Quick Add/Edit Time',
							content: 'Each day is clickable for quick time entry management.',
							order: 7,
							target: 'div.space-y-2 > div:nth-child(1) > button:nth-child(1)'
						}
					]
				});

				// Mark tour as completed when it finishes
				tour.onFinish(() => {
					api.completeTour().then((updatedUser) => {
						authStore.setUser(updatedUser);
					});
				});

				tour.start();
			}, 500);

			return () => {
				clearTimeout(timer);
				if (tour) {
					tour.exit();
					// Remove DOM elements added by the tour
					const tourElements = document.querySelectorAll('.tg-dialog, .tg-dialog-overlay');
					tourElements.forEach((el) => el.remove());
					tour = null;
				}
			};
		}
	});

	$effect(() => {
		activeUrl = $page.url.pathname;
	});

	// Consolidated redirect logic to prevent multiple rapid navigations
	$effect(() => {
		if (!$authStore.user) return;

		const currentPath = $page.url.pathname;

		// Redirect to onboarding if not completed
		if (!$authStore.user.onboarding_completed && currentPath !== '/onboarding') {
			goto('/onboarding');
			return;
		}
	});

	const toggleSidebar = () => {
		sidebarOpen = !sidebarOpen;
	};

	function handleLogout() {
		api.logout();
		authStore.clearAuth();
		goto('/login');
	}

	const spanClass = 'flex-1 ms-3 whitespace-nowrap';
	const activeClass =
		'flex items-center p-2 text-base font-normal text-white bg-primary-600 dark:bg-primary-700 rounded-lg dark:text-white hover:bg-primary-700 dark:hover:bg-primary-600';
	const nonActiveClass =
		'flex items-center p-2 text-base font-normal text-gray-900 rounded-lg dark:text-white hover:bg-gray-100 dark:hover:bg-gray-700';
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>FacturMe - Time & Invoice Management</title>
	<script>
		// Set dark mode by default
		document.documentElement.classList.add('dark');
	</script>
</svelte:head>

<div class="dark min-h-screen bg-gradient-to-br from-slate-900 via-slate-900 to-slate-800">
	<!-- Top Navbar -->
	<Navbar class="fixed top-0 border-b border-gray-600 bg-gray-800 z-50 py-3" fluid={true}>
		<NavBrand href="/">
			<img src={logo} alt="FacturMe" class="h-12" />
		</NavBrand>
		<div class="flex items-center md:order-2 gap-3">
			<Avatar id="avatar-menu" class="cursor-pointer" size="md" />
			<NavHamburger onclick={toggleSidebar} class="lg:hidden" />
		</div>
		<Dropdown placement="bottom" triggeredBy="#avatar-menu">
			<DropdownHeader>
				<span class="block text-sm text-gray-500 dark:text-gray-400">Signed in as</span>
				{#if $authStore.user?.name}
					<span class="block truncate text-sm">
						{$authStore.user.name}
					</span>
				{/if}
				<span class="block truncate text-sm font-semibold">
					{$authStore.user?.email || 'user@example.com'}
				</span>
			</DropdownHeader>
			<DropdownGroup>
				<DropdownItem class="w-full text-left" onclick={handleLogout}>Sign out</DropdownItem>
			</DropdownGroup>
		</Dropdown>
	</Navbar>

	<div class="flex">
		<!-- Sidebar -->
		<Sidebar
			{activeUrl}
			backdrop={false}
			classes={{ nonactive: nonActiveClass, active: activeClass }}
			class={`fixed top-0 mt-[73px] left-0 z-40 h-[calc(100vh-73px)] w-64 transition-transform bg-slate-800 ${
				sidebarOpen ? 'translate-x-0' : '-translate-x-full'
			} lg:translate-x-0`}
		>
			<SidebarGroup class="space-y-2 pb-48">
				<SidebarItem label="Dashboard" {spanClass} href="/">
					{#snippet icon()}
						<ChartPieSolid
							class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					{/snippet}
				</SidebarItem>
				<SidebarItem label="Calendar" {spanClass} href="/calendar" id="calendar-link">
					{#snippet icon()}
						<CalendarMonthSolid
							class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					{/snippet}
				</SidebarItem>
				<SidebarItem label="Time Tracking" {spanClass} href="/time">
					{#snippet icon()}
						<ClockSolid
							class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					{/snippet}
				</SidebarItem>
				<SidebarItem label="Invoices" {spanClass} href="/invoices">
					{#snippet icon()}
						<FileInvoiceSolid
							class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					{/snippet}
				</SidebarItem>
				<SidebarItem label="Clients" {spanClass} href="/clients">
					{#snippet icon()}
						<UsersGroupSolid
							class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					{/snippet}
				</SidebarItem>
				<SidebarItem label="Settings" {spanClass} href="/settings">
					{#snippet icon()}
						<CogSolid
							class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					{/snippet}
				</SidebarItem>
			</SidebarGroup>
		</Sidebar>

		<!-- Main Content -->
		<main class="flex-1 lg:ml-64 !pt-22">
			<!-- Demo Banner -->
			<DemoBanner />

			<div class="p-4 md:p-6">
				{@render children?.()}
			</div>
		</main>
	</div>

	<!-- Toast Notification -->
	{#if showToast}
		<Toast color="red" position="top-right" class="fixed top-4 right-4 z-50" bind:open={showToast}>
			{#snippet icon()}
				<ExclamationCircleSolid class="w-5 h-5" />
			{/snippet}
			{toastMessage}
		</Toast>
	{/if}
</div>

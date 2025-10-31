<script lang="ts">
	import Onboarding from '$lib/components/Onboarding.svelte';
	import { goto } from '$app/navigation';

	function handleOnboardingComplete() {
		// In real implementation, this would save the onboarding data
		// and mark the user as onboarded in the database
		goto('/');
	}
</script>

<Onboarding onComplete={handleOnboardingComplete} />

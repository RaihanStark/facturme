<script lang="ts">
	import { Button } from 'flowbite-svelte';

	interface Props {
		title: string;
		href?: string;
		buttonText?: string;
		children?: import('svelte').Snippet;
	}

	let { title, href, buttonText = 'View All', children }: Props = $props();
</script>

<div class="bg-slate-800 rounded-xl border border-slate-700 shadow-lg">
	<div class="px-6 py-4 border-b border-slate-700">
		<div class="flex items-center justify-between">
			<h2 class="text-lg font-semibold text-white">{title}</h2>
			{#if href}
				<Button {href} size="xs" color="light">{buttonText}</Button>
			{/if}
		</div>
	</div>
	<div class="p-6">
		{@render children?.()}
	</div>
</div>

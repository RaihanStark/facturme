<script lang="ts">
	interface Props {
		icon: import('svelte').Snippet;
		title: string;
		description?: string;
	}

	let { icon, title, description }: Props = $props();
</script>

<div class="text-center py-12">
	<div
		class="w-16 h-16 bg-slate-700/50 rounded-full flex items-center justify-center mx-auto mb-4"
	>
		{@render icon?.()}
	</div>
	<p class="text-slate-400 font-medium">{title}</p>
	{#if description}
		<p class="text-sm text-slate-500 mt-1">{description}</p>
	{/if}
</div>

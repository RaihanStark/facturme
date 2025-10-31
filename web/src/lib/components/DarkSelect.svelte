<script lang="ts">
	import { Select } from 'flowbite-svelte';

	interface SelectItem {
		value: string;
		name: string;
	}

	interface Props {
		id?: string;
		value?: string;
		items?: SelectItem[];
		size?: 'sm' | 'md' | 'lg';
		disabled?: boolean;
		class?: string;
	}

	let {
		id,
		value = $bindable(),
		items = [],
		size = 'md',
		disabled = false,
		class: customClass = ''
	}: Props = $props();

	const baseClasses =
		'bg-slate-700 border-slate-600 text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 [&>option]:bg-slate-700 [&>option]:text-white';
	const combinedClasses = `${baseClasses} ${customClass}`.trim();
</script>

<Select {id} bind:value {items} {size} {disabled} selectClass={combinedClasses} />

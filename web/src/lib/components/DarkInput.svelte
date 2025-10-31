<script lang="ts">
	import { Input } from 'flowbite-svelte';
	import type { InputType } from 'flowbite-svelte';

	interface Props {
		id?: string;
		type?: InputType;
		value?: string | number;
		placeholder?: string;
		required?: boolean;
		disabled?: boolean;
		size?: 'sm' | 'md' | 'lg';
		min?: string | number;
		max?: string | number;
		step?: string | number;
		class?: string;
	}

	let {
		id,
		type = 'text',
		value = $bindable(),
		placeholder,
		required = false,
		disabled = false,
		size = 'md',
		min,
		max,
		step,
		class: customClass = ''
	}: Props = $props();

	const baseClasses =
		'bg-slate-700 border-slate-600 text-white placeholder-slate-400 focus:ring-blue-500 focus:border-blue-500';
	const combinedClasses = `${baseClasses} ${customClass}`.trim();
</script>

<Input
	{id}
	{type}
	bind:value
	{placeholder}
	{required}
	{disabled}
	{size}
	{min}
	{max}
	{step}
	class={combinedClasses}
/>

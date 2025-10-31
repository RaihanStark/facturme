<script lang="ts">
	import '../app.css';
	import '@sjmc11/tourguidejs/src/scss/tour.scss';
	let { children } = $props();
</script>

<!-- Main Content -->
{@render children?.()}

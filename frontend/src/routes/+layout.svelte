<script lang="ts">
  import '../app.css';
  import AppShell from '@/components/studio/AppShell.svelte';

  let { children } = $props();
</script>

<AppShell>
  {@render children()}
</AppShell>

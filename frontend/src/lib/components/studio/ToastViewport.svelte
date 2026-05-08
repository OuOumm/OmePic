<script lang="ts">
  import { CheckCircle, Info, XCircle } from 'lucide-svelte';
  import { toasts } from '@/stores/toast.svelte';
</script>

<div class="fixed right-4 top-4 z-[80] grid gap-3">
  {#each toasts.items as toast (toast.id)}
    <div
      class="studio-panel flex max-w-sm items-start gap-3 px-4 py-3 text-sm animate-in"
      data-tone={toast.tone}
    >
      {#if toast.tone === 'success'}
        <CheckCircle class="mt-0.5 size-4 text-emerald-700" />
      {:else if toast.tone === 'error'}
        <XCircle class="mt-0.5 size-4 text-red-700" />
      {:else}
        <Info class="mt-0.5 size-4 text-sky-700" />
      {/if}
      <span>{toast.message}</span>
    </div>
  {/each}
</div>

<style>
  .animate-in {
    animation: toast-in 180ms ease-out both;
  }

  @keyframes toast-in {
    from {
      opacity: 0;
      transform: translate(12px, -8px) rotate(1deg);
    }
    to {
      opacity: 1;
      transform: translate(0, 0) rotate(-0.2deg);
    }
  }
</style>


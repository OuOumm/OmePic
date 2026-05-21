<script lang="ts">
  import { page } from '$app/state';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { ArrowLeft, Home, RefreshCw } from 'lucide-svelte';

  const status = $derived(page.status || 500);
  const isNotFound = $derived(status === 404);
  const isForbidden = $derived(status === 403);
  const titleKey = $derived(
    isNotFound ? 'error.notFoundTitle' : isForbidden ? 'error.forbiddenTitle' : 'error.genericTitle'
  );
  const messageKey = $derived(
    isNotFound ? 'error.notFoundMessage' : isForbidden ? 'error.forbiddenMessage' : 'error.genericMessage'
  );
  const recoveryKey = $derived(
    isNotFound ? 'error.notFoundHint' : isForbidden ? 'error.forbiddenHint' : 'error.genericHint'
  );
  const accentVar = $derived(
    isNotFound ? 'var(--marker-blue)' : isForbidden ? 'var(--marker-yellow)' : 'var(--danger)'
  );

  function handleSecondaryAction() {
    if (isNotFound || isForbidden) {
      window.history.back();
      return;
    }
    window.location.reload();
  }
</script>

<svelte:head><title>{t(preferences.language, titleKey)} · OmePic</title></svelte:head>

<div class="grid min-h-[calc(100dvh-8rem)] place-items-center px-4 py-6 sm:px-6">
  <section class="studio-panel sketch-enter relative w-full max-w-3xl overflow-hidden p-6 sm:p-8">
    <div class="pointer-events-none absolute inset-x-6 top-5 flex items-center justify-between text-[10px] font-black uppercase tracking-[0.32em] text-[hsl(var(--ink-muted))] sm:inset-x-8">
      <span>{t(preferences.language, 'error.eyebrow')}</span>
      <span>{status}</span>
    </div>

    <div class="absolute -right-8 top-12 h-24 w-24 rotate-[14deg] rounded-full border-[3px] border-dashed border-[hsl(var(--ink)/0.2)] bg-[hsl(var(--paper-deep)/0.45)]" aria-hidden="true"></div>
    <div class="absolute left-6 top-16 h-3 w-24 rotate-[-3deg] rounded-full blur-[1px] sm:left-8" style={`background:hsl(${accentVar} / 0.58)`}></div>

    <div class="relative space-y-5 pt-10">
      <p class="inline-flex w-fit rotate-[-1deg] border-[3px] ink-line bg-[hsl(var(--paper))] px-3 py-1 text-sm font-black shadow-[4px_4px_0_hsl(var(--ink))]">
        {status} · {t(preferences.language, 'error.statusLabel')}
      </p>
      <div class="space-y-3">
        <h1 class="max-w-2xl text-3xl font-black leading-tight sm:text-5xl">
          {t(preferences.language, titleKey)}
        </h1>
        <p class="max-w-2xl text-sm font-bold leading-7 text-[hsl(var(--ink-muted))] sm:text-base">
          {t(preferences.language, messageKey)}
        </p>
      </div>

      <div class="blueprint-grid grid gap-3 border-[3px] border-dashed ink-line bg-[hsl(var(--paper-deep)/0.42)] p-4">
        <span class="tape-label w-fit rotate-[1deg]" style={`background:hsl(${accentVar} / 0.72)`}>
          {t(preferences.language, 'error.quickRecovery')}
        </span>
        <p class="pt-3 text-sm font-bold leading-6 text-[hsl(var(--ink-muted))]">
          {t(preferences.language, recoveryKey)}
        </p>
      </div>

      <div class="flex flex-wrap justify-end gap-3 pt-1">
        <a href="/" class="studio-button inline-flex" data-tone="primary">
          <Home class="size-4" />
          {t(preferences.language, 'error.backHome')}
        </a>
        <button class="studio-button inline-flex" type="button" onclick={handleSecondaryAction}>
          {#if isNotFound || isForbidden}
            <ArrowLeft class="size-4" />
            {t(preferences.language, 'error.goBack')}
          {:else}
            <RefreshCw class="size-4" />
            {t(preferences.language, 'error.tryAgain')}
          {/if}
        </button>
      </div>
    </div>
  </section>
</div>

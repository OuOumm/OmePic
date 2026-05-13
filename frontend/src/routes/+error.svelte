<script lang="ts">
  import { page } from '$app/state';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { AlertTriangle, Home } from 'lucide-svelte';
</script>

<svelte:head><title>{t(preferences.language, 'common.error')} · OmePic</title></svelte:head>

<div class="grid min-h-[calc(100dvh-8rem)] place-items-center px-4">
  <div class="studio-panel max-w-lg p-8 text-center sketch-enter">
    <AlertTriangle class="mx-auto mb-4 size-12 text-[hsl(var(--danger))]" aria-hidden="true" />
    <h1 class="text-4xl font-black">{t(preferences.language, 'common.error')}</h1>
    <p class="mt-4 text-lg font-bold text-[hsl(var(--ink-muted))]">
      {page.status === 404
        ? t(preferences.language, 'history.noMatches')
        : t(preferences.language, 'common.error')}
    </p>
    <a href="/" class="studio-button mt-8 inline-flex" data-tone="primary">
      <Home class="size-4" />
      {t(preferences.language, 'nav.upload')}
    </a>
  </div>
</div>

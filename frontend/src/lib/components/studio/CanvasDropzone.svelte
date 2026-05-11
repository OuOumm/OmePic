<script lang="ts">
  import { Clipboard, Image as ImageIcon, Link2, Zap } from 'lucide-svelte';
  import type { Language } from '@/types';
  import { t } from '@/i18n';

  export let language: Language;
  export let disabled = false;
  export let dragging = false;
  export let subtitle: string | null = null;
  export let allowedTypes = '';
  export let onFiles: (files: File[]) => void;

  let input: HTMLInputElement;

  function openPicker() {
    if (!disabled) input?.click();
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    if (disabled) return;
    dragging = false;
    const files = Array.from(event.dataTransfer?.files ?? []).filter((file) => file.type.startsWith('image/'));
    if (files.length) onFiles(files);
  }

  function handleChange() {
    const files = Array.from(input.files ?? []);
    if (files.length) onFiles(files);
    input.value = '';
  }

  const allowedTypesId = 'upload-allowed-types';
</script>

<div
  class="group relative flex min-h-[390px] w-full overflow-hidden border-[3px] border-dashed ink-line bg-[hsl(var(--paper)/0.72)] px-6 text-left transition-transform duration-200 {dragging ? '-rotate-1 scale-[1.01] bg-[hsl(var(--marker-yellow)/0.28)]' : 'hover:-rotate-[0.35deg]'} {disabled ? 'cursor-not-allowed opacity-60' : ''}"
  role="group"
  aria-describedby={allowedTypes ? allowedTypesId : undefined}
  ondragover={(event) => event.preventDefault()}
  ondrop={handleDrop}
>
  <svg class="absolute inset-0 h-full w-full opacity-60" viewBox="0 0 900 420" preserveAspectRatio="none" aria-hidden="true">
    <path class="ink-draw" d="M47 72 C187 18, 272 77, 392 48 C518 18, 624 42, 837 24" stroke="hsl(var(--ink))" stroke-width="3" stroke-linecap="round" fill="none" />
    <path class="ink-draw" style="animation-delay:.22s" d="M92 343 C216 278, 345 381, 486 315 C612 256, 709 339, 850 282" stroke="hsl(var(--marker-blue))" stroke-width="10" stroke-linecap="round" fill="none" />
  </svg>

  <div class="relative z-10 grid w-full gap-8 py-8 lg:grid-cols-[1fr_260px] lg:items-center">
    <div>
      <div class="tape-label mb-6 rotate-[-4deg]">{t(language, 'upload.title')}</div>
      <h1 class="max-w-4xl text-5xl font-black leading-[0.88] tracking-tight md:text-7xl lg:text-8xl">
        {t(language, 'upload.dropTitle')}
      </h1>
      <p class="mt-6 max-w-2xl text-base font-bold text-[hsl(var(--ink-muted))] md:text-lg">
        {subtitle ?? t(language, 'upload.dropSubtitle')}
      </p>
      {#if allowedTypes}
        <p id={allowedTypesId} class="mt-3 text-sm font-bold text-[hsl(var(--ink-muted))]">{t(language, 'upload.allowedTypes', { types: allowedTypes })}</p>
      {/if}
      <button class="studio-button mt-8" data-tone="primary" type="button" disabled={disabled} onclick={openPicker} aria-describedby={allowedTypes ? allowedTypesId : undefined}><ImageIcon class="size-4" aria-hidden="true" />{t(language, 'upload.select')}</button>
    </div>

    <div class="relative hidden min-h-72 lg:block">
      <div class="absolute left-4 top-2 rotate-[-7deg] border-2 ink-line bg-[hsl(var(--marker-pink))] px-4 py-3 font-black shadow-[4px_4px_0_hsl(var(--ink))]">
        <Clipboard class="mb-2 size-7" />{t(language, 'upload.sourcePaste')}
      </div>
      <div class="absolute right-2 top-24 rotate-[5deg] border-2 ink-line bg-[hsl(var(--marker-blue))] px-4 py-3 font-black shadow-[4px_4px_0_hsl(var(--ink))]">
        <Link2 class="mb-2 size-7" />URL
      </div>
      <div class="absolute bottom-3 left-10 rotate-[-2deg] border-2 ink-line bg-[hsl(var(--marker-green))] px-4 py-3 font-black shadow-[4px_4px_0_hsl(var(--ink))]">
        <Zap class="mb-2 size-7" />{t(language, 'upload.sourceHost')}
      </div>
    </div>
  </div>

  <input bind:this={input} class="sr-only" type="file" accept="image/*" multiple onchange={handleChange} aria-label={t(language, 'upload.select')} aria-describedby={allowedTypes ? allowedTypesId : undefined} />
</div>

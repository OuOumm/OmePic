<script lang="ts">
  import { UploadCloud } from 'lucide-svelte';
  import type { Language } from '@/types';
  import { t } from '@/i18n';

  export let language: Language;
  export let disabled = false;
  export let dragging = false;
  export let onFiles: (files: File[]) => void;

  let input: HTMLInputElement;

  function openPicker() {
    if (!disabled) input?.click();
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    if (disabled) return;
    const files = Array.from(event.dataTransfer?.files ?? []).filter((file) => file.type.startsWith('image/'));
    if (files.length) onFiles(files);
  }

  function handleChange() {
    const files = Array.from(input.files ?? []);
    if (files.length) onFiles(files);
    input.value = '';
  }
</script>

<button
  type="button"
  class="relative flex min-h-[310px] w-full flex-col items-center justify-center border-[3px] border-dashed ink-line bg-[hsl(var(--paper)/0.72)] px-6 text-center transition-transform duration-200 {dragging ? '-rotate-1 scale-[1.01] bg-[hsl(var(--marker-yellow)/0.28)]' : 'hover:-rotate-[0.35deg]'} disabled:cursor-not-allowed disabled:opacity-60"
  disabled={disabled}
  onclick={openPicker}
  ondragover={(event) => event.preventDefault()}
  ondrop={handleDrop}
>
  <div class="absolute left-5 top-5 rotate-[-6deg] bg-[hsl(var(--marker-pink))] px-3 py-1 text-xs font-black uppercase shadow-[3px_3px_0_hsl(var(--ink))]">
    {t(language, 'upload.title')}
  </div>
  <UploadCloud class="mb-5 size-16" strokeWidth={1.8} />
  <h1 class="max-w-3xl text-4xl font-black leading-none tracking-tight md:text-6xl">
    {t(language, 'upload.dropTitle')}
  </h1>
  <p class="mt-4 max-w-xl text-base font-semibold text-[hsl(var(--ink-muted))]">
    {t(language, 'upload.dropSubtitle')}
  </p>
  <span class="studio-button mt-7" data-tone="primary">{t(language, 'upload.select')}</span>
  <input bind:this={input} class="sr-only" type="file" accept="image/*" multiple onchange={handleChange} />
</button>


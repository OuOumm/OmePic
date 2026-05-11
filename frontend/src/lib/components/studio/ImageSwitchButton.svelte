<script lang="ts">
  import { t } from '@/i18n';
  import type { Language } from '@/types';

  type Props = {
    direction: 'previous' | 'next';
    language: Language;
    disabled?: boolean;
    onclick: () => void;
  };

  let { direction, language, disabled = false, onclick }: Props = $props();

  const isPrevious = $derived(direction === 'previous');
  const label = $derived(t(language, isPrevious ? 'common.previousImage' : 'common.nextImage'));
  const arrowPath = $derived(isPrevious ? 'M45 16 L22 36 L45 56' : 'M27 16 L50 36 L27 56');
</script>

<button
  class="group absolute top-1/2 z-10 grid size-12 -translate-y-1/2 place-items-center border-[3px] ink-line bg-[hsl(var(--paper)/0.94)] shadow-[4px_4px_0_hsl(var(--ink)/0.72)] touch-manipulation transition-[opacity,transform,background-color] duration-150 hover:scale-105 hover:bg-[hsl(var(--marker-yellow))] focus-visible:bg-[hsl(var(--marker-yellow))] sm:size-14 {isPrevious ? 'left-6 sm:left-10' : 'right-6 sm:right-10'} disabled:pointer-events-none disabled:opacity-30"
  type="button"
  {disabled}
  aria-label={label}
  onclick={onclick}
>
  <svg class="size-9 sm:size-10" viewBox="0 0 72 72" aria-hidden="true">
    <path d="M10 36 C22 30, 50 30, 62 36" fill="none" stroke="hsl(var(--marker-yellow))" stroke-width="13" stroke-linecap="round" opacity="0.92" />
    <path d={arrowPath} fill="none" stroke="hsl(var(--ink))" stroke-width="8" stroke-linecap="round" stroke-linejoin="round" />
  </svg>
</button>

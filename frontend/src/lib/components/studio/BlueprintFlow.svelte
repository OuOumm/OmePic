<script lang="ts">
  import type { Language } from '@/types';
  import { t } from '@/i18n';

  type Step = string | { id: string; label: string };

  type Props = {
    language?: Language;
    label?: string;
    steps?: Step[];
    activeIndex?: number;
  };

  let { language = 'en', label, steps = [], activeIndex = 0 }: Props = $props();
  const heading = $derived(label ?? t(language, 'studio.pipeline'));

  function stepKey(step: Step) {
    return typeof step === 'string' ? step : step.id;
  }

  function stepLabel(step: Step) {
    return typeof step === 'string' ? step : step.label;
  }
</script>

<div class="blueprint-grid content-auto border-[3px] ink-line p-5 sketch-enter">
  <div class="mb-4 flex items-center justify-between gap-3">
    <h2 class="text-xl font-black">{heading}</h2>
    <span class="tape-label rotate-2">{t(language, 'studio.liveSketch')}</span>
  </div>
  <div class="relative grid gap-5 md:grid-cols-{steps.length || 1}">
    <svg class="absolute left-0 top-1/2 hidden h-10 w-full -translate-y-1/2 md:block" viewBox="0 0 900 40" preserveAspectRatio="none" aria-hidden="true">
      <path class="ink-draw" d="M20 20 C210 4, 335 38, 450 18 S680 14, 880 20" stroke="hsl(var(--ink))" stroke-width="3" stroke-dasharray="9 8" fill="none" stroke-linecap="round" />
    </svg>
    {#each steps as step, index (stepKey(step))}
      <div class="relative z-10 flex items-center gap-3 bg-[hsl(var(--paper)/0.72)] p-2">
        <span class="grid size-9 place-items-center rounded-full border-2 ink-line font-black {index <= activeIndex ? 'bg-[hsl(var(--marker-green))]' : 'bg-[hsl(var(--paper))]'}">{index + 1}</span>
        <span class="font-black">{stepLabel(step)}</span>
      </div>
    {/each}
  </div>
</div>

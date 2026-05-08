<script lang="ts">
  import { Sparkles } from 'lucide-svelte';

  type Props = {
    title: string;
    subtitle?: string;
    eyebrow?: string;
    tone?: 'yellow' | 'blue' | 'pink' | 'green';
  };

  let { title, subtitle = '', eyebrow = '', tone = 'yellow' }: Props = $props();

  const toneVar = $derived(
    tone === 'blue'
      ? 'var(--marker-blue)'
      : tone === 'pink'
        ? 'var(--marker-pink)'
        : tone === 'green'
          ? 'var(--marker-green)'
          : 'var(--marker-yellow)',
  );
</script>

<section class="relative overflow-hidden border-b-[3px] ink-line pb-6 sketch-enter">
  <svg class="pointer-events-none absolute right-0 top-2 hidden h-24 w-72 opacity-70 md:block" viewBox="0 0 300 90" fill="none" aria-hidden="true">
    <path class="ink-draw" d="M6 72 C42 18, 96 86, 138 37 S229 14, 292 54" stroke="hsl(var(--ink))" stroke-width="3" stroke-linecap="round" />
    <path class="ink-draw" style="animation-delay:.18s" d="M42 62 C82 49, 156 58, 236 31" stroke={`hsl(${toneVar} / .95)`} stroke-width="10" stroke-linecap="round" />
  </svg>
  {#if eyebrow}
    <p class="tape-label mb-3 rotate-[-1deg]" style={`background:hsl(${toneVar})`}>
      <Sparkles class="mr-1 inline size-3" />{eyebrow}
    </p>
  {/if}
  <h1 class="max-w-5xl text-4xl font-black leading-none tracking-tight md:text-6xl lg:text-7xl">{title}</h1>
  {#if subtitle}
    <p class="mt-4 max-w-3xl text-base font-semibold text-[hsl(var(--ink-muted))] md:text-lg">{subtitle}</p>
  {/if}
</section>

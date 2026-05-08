<script lang="ts">
  export let label = 'Pipeline';
  export let steps: string[] = [];
  export let activeIndex = 0;
</script>

<div class="blueprint-grid border-[3px] ink-line p-5 sketch-enter">
  <div class="mb-4 flex items-center justify-between gap-3">
    <h2 class="text-xl font-black">{label}</h2>
    <span class="tape-label rotate-2">live sketch</span>
  </div>
  <div class="relative grid gap-5 md:grid-cols-{steps.length || 1}">
    <svg class="absolute left-0 top-1/2 hidden h-10 w-full -translate-y-1/2 md:block" viewBox="0 0 900 40" preserveAspectRatio="none" aria-hidden="true">
      <path class="ink-draw" d="M20 20 C210 4, 335 38, 450 18 S680 14, 880 20" stroke="hsl(var(--ink))" stroke-width="3" stroke-dasharray="9 8" fill="none" stroke-linecap="round" />
    </svg>
    {#each steps as step, index (step)}
      <div class="relative z-10 flex items-center gap-3 bg-[hsl(var(--paper)/0.72)] p-2">
        <span class="grid size-9 place-items-center rounded-full border-2 ink-line font-black {index <= activeIndex ? 'bg-[hsl(var(--marker-green))]' : 'bg-[hsl(var(--paper))]'}">{index + 1}</span>
        <span class="font-black">{step}</span>
      </div>
    {/each}
  </div>
</div>

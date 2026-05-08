<script lang="ts">
  export let title: string;
  export let subtitle = '';
  export let eyebrow = '';
</script>

<section class="border-b-[3px] ink-line pb-5">
  {#if eyebrow}
    <p class="mb-2 inline-block rotate-[-1deg] bg-[hsl(var(--marker-yellow))] px-2 py-1 text-xs font-black uppercase shadow-[3px_3px_0_hsl(var(--ink))]">{eyebrow}</p>
  {/if}
  <h1 class="text-4xl font-black leading-none tracking-tight md:text-6xl">{title}</h1>
  {#if subtitle}
    <p class="mt-3 max-w-3xl text-base font-semibold text-[hsl(var(--ink-muted))] md:text-lg">{subtitle}</p>
  {/if}
</section>


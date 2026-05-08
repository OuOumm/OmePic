<script lang="ts">
  type Props = {
    label: string;
    value: string | number;
    detail?: string;
    tone?: 'yellow' | 'blue' | 'pink' | 'green';
  };

  let { label, value, detail = '', tone = 'yellow' }: Props = $props();

  const background = $derived(
    tone === 'blue'
      ? 'hsl(var(--marker-blue) / 0.28)'
      : tone === 'pink'
        ? 'hsl(var(--marker-pink) / 0.28)'
        : tone === 'green'
          ? 'hsl(var(--marker-green) / 0.28)'
          : 'hsl(var(--marker-yellow) / 0.28)',
  );
</script>

<div class="relative border-y-[3px] ink-line py-4 sketch-enter" style={`background:${background}`}>
  <p class="text-xs font-black uppercase tracking-[0.12em] text-[hsl(var(--ink-muted))]">{label}</p>
  <p class="mt-1 text-3xl font-black leading-none md:text-4xl">{value}</p>
  {#if detail}<p class="mt-2 text-xs font-bold text-[hsl(var(--ink-muted))]">{detail}</p>{/if}
</div>

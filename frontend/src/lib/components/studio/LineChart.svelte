<script lang="ts">
  type Point = {
    id: string | number;
    value: number;
    tone?: 'green' | 'pink' | 'blue';
    label?: string;
    timestamp?: string;
  };

  type MetaItem = { label: string; value: string };

  type Props = {
    points: Point[];
    emptyLabel: string;
    ariaLabel: string;
    minValue?: number;
    maxValue?: number;
    heightClass?: string;
    metadata?: MetaItem[];
  };

  let { points, emptyLabel, ariaLabel, minValue, maxValue, heightClass = 'h-64', metadata }: Props = $props();

  // Layout
  const W = 360;
  const H = 180;
  const ML = 44; // margin left (Y-axis labels)
  const MR = 12;
  let MT = $derived(metadata && metadata.length ? 30 : 14);
  const MB = 28; // margin bottom (X-axis labels)
  let plotW = $derived(W - ML - MR);
  let plotH = $derived(H - MT - MB);

  const activePoints = $derived(points.filter((p) => Number.isFinite(p.value)));

  const range = $derived.by(() => {
    const values = activePoints.map((p) => p.value);
    const lo = minValue ?? Math.min(...values, 0);
    const hi = maxValue ?? Math.max(...values, 1);
    return hi <= lo ? { min: lo, max: lo + 1 } : { min: lo, max: hi };
  });

  const avg = $derived.by(() => {
    if (!activePoints.length) return 0;
    return activePoints.reduce((s, p) => s + p.value, 0) / activePoints.length;
  });

  function x(i: number): number {
    if (activePoints.length <= 1) return ML + plotW / 2;
    return ML + (i / (activePoints.length - 1)) * plotW;
  }

  function y(v: number): number {
    const ratio = (v - range.min) / (range.max - range.min);
    return MT + plotH - ratio * plotH;
  }

  // Y-axis ticks: 4 evenly spaced
  const yTicks = $derived.by(() => {
    const count = 4;
    return Array.from({ length: count + 1 }, (_, i) => {
      const v = range.min + (i / count) * (range.max - range.min);
      return { value: v, y: y(v), label: `${Math.round(v)} ms` };
    });
  });

  // X-axis ticks: up to 5 time labels
  const xTicks = $derived.by(() => {
    if (activePoints.length < 2) return [];
    const count = Math.min(5, activePoints.length);
    return Array.from({ length: count }, (_, i) => {
      const idx = Math.round((i / (count - 1)) * (activePoints.length - 1));
      const p = activePoints[idx];
      return { x: x(idx), label: p.timestamp ? shortTime(p.timestamp) : '' };
    });
  });

  function shortTime(ts: string): string {
    try {
      const d = new Date(ts);
      const hh = String(d.getHours()).padStart(2, '0');
      const mm = String(d.getMinutes()).padStart(2, '0');
      return `${hh}:${mm}`;
    } catch {
      return '';
    }
  }

  function linePath(): string {
    return activePoints.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(i).toFixed(1)},${y(p.value).toFixed(1)}`).join(' ');
  }

  function areaPath(): string {
    if (!activePoints.length) return '';
    const top = activePoints.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(i).toFixed(1)},${y(p.value).toFixed(1)}`).join(' ');
    return `${top} L${x(activePoints.length - 1).toFixed(1)},${(MT + plotH).toFixed(1)} L${x(0).toFixed(1)},${(MT + plotH).toFixed(1)} Z`;
  }

  // Failure points (pink dots)
  const failures = $derived(
    activePoints
      .map((p, i) => ({ ...p, i, cx: x(i), cy: y(p.value) }))
      .filter((p) => p.tone === 'pink')
  );
</script>

{#if activePoints.length < 2}
  <div class={`flex items-center justify-center border-2 ink-line bg-[hsl(var(--paper-deep)/0.76)] text-xs font-bold text-[hsl(var(--ink-muted))] ${heightClass}`}>
    {emptyLabel}
  </div>
{:else}
  <svg viewBox={`0 0 ${W} ${H}`} class={`w-full ${heightClass}`} role="img" aria-label={ariaLabel}>
    <!-- Y-axis grid lines + labels -->
    {#each yTicks as tick, i (i)}
      <line x1={ML} y1={tick.y} x2={W - MR} y2={tick.y} stroke="currentColor" stroke-opacity={i === 0 ? 0.2 : 0.08} stroke-width="0.8" vector-effect="non-scaling-stroke" />
      <text x={ML - 6} y={tick.y + 3.5} text-anchor="end" class="fill-[hsl(var(--ink-muted))] text-[8px] font-bold" style="font-family: ui-monospace, monospace">{tick.label}</text>
    {/each}

    <!-- Area fill -->
    <path d={areaPath()} fill="url(#lc-area-grad)" />

    <!-- Average reference line -->
    <line x1={ML} y1={y(avg)} x2={W - MR} y2={y(avg)} stroke="hsl(var(--marker-blue))" stroke-opacity="0.45" stroke-width="1" stroke-dasharray="4 3" vector-effect="non-scaling-stroke" />
    <text x={W - MR} y={y(avg) - 4} text-anchor="end" class="fill-[hsl(var(--marker-blue))] text-[7px] font-black" style="font-family: ui-monospace, monospace">avg {Math.round(avg)} ms</text>

    <!-- Main line -->
    <path d={linePath()} fill="none" stroke="hsl(var(--ink))" stroke-width="1.8" stroke-linejoin="round" stroke-linecap="round" vector-effect="non-scaling-stroke" />

    <!-- Failure dots -->
    {#each failures as f (f.id)}
      <circle cx={f.cx} cy={f.cy} r="4" fill="hsl(var(--marker-pink))" stroke="hsl(var(--ink))" stroke-width="1.2" vector-effect="non-scaling-stroke">
        {#if f.label}<title>{f.label}</title>{/if}
      </circle>
    {/each}

    <!-- Metadata annotations -->
    {#if metadata && metadata.length}
      <g transform="translate({MR}, 4)">
        {#each metadata as item, i (i)}
          {@const offsetX = i * 76}
          <text x={offsetX} y={0} class="fill-[hsl(var(--ink-muted))] text-[5.5px] font-bold" style="font-family: ui-monospace, monospace">{item.label}</text>
          <text x={offsetX} y={8} class="fill-[hsl(var(--ink))] text-[7px] font-black" style="font-family: ui-monospace, monospace">{item.value}</text>
        {/each}
      </g>
    {/if}

    <!-- X-axis labels -->
    {#each xTicks as tick, i (i)}
      <text x={tick.x} y={H - 6} text-anchor="middle" class="fill-[hsl(var(--ink-muted))] text-[8px] font-bold" style="font-family: ui-monospace, monospace">{tick.label}</text>
    {/each}

    <defs>
      <linearGradient id="lc-area-grad" x1="0" x2="0" y1="0" y2="1">
        <stop offset="0%" stop-color="hsl(var(--marker-blue))" stop-opacity="0.25" />
        <stop offset="100%" stop-color="hsl(var(--marker-blue))" stop-opacity="0.03" />
      </linearGradient>
    </defs>
  </svg>
{/if}

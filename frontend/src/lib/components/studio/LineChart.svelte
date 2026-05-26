<script lang="ts">
  type Point = {
    id: string | number;
    value: number;
    tone?: 'green' | 'pink' | 'blue';
    label?: string;
  };

  type Props = {
    points: Point[];
    emptyLabel: string;
    ariaLabel: string;
    minValue?: number;
    maxValue?: number;
    heightClass?: string;
  };

  let { points, emptyLabel, ariaLabel, minValue, maxValue, heightClass = 'h-32' }: Props = $props();

  const chartWidth = 100;
  const topPadding = 10;
  const bottomPadding = 88;
  const activePoints = $derived(points.filter((point) => Number.isFinite(point.value)));
  const range = $derived.by(() => {
    const values = activePoints.map((point) => point.value);
    const min = minValue ?? Math.min(...values, 0);
    const max = maxValue ?? Math.max(...values, 1);
    return max <= min ? { min, max: min + 1 } : { min, max };
  });

  function pointX(index: number) {
    if (activePoints.length <= 1) return chartWidth / 2;
    return (index / (activePoints.length - 1)) * chartWidth;
  }

  function pointY(value: number) {
    const ratio = (value - range.min) / (range.max - range.min);
    return bottomPadding - ratio * (bottomPadding - topPadding);
  }

  function pathPoints() {
    return activePoints.map((point, index) => `${pointX(index).toFixed(1)},${pointY(point.value).toFixed(1)}`).join(' ');
  }

  function fillColor(tone: Point['tone']) {
    if (tone === 'pink') return 'hsl(var(--marker-pink))';
    if (tone === 'blue') return 'hsl(var(--marker-blue))';
    return 'hsl(var(--marker-green))';
  }
</script>

<svg viewBox="0 0 100 96" class={`${heightClass} w-full overflow-visible border-2 ink-line bg-[hsl(var(--paper-deep))]`} role="img" aria-label={ariaLabel}>
  <line x1="0" y1="88" x2="100" y2="88" stroke="currentColor" stroke-opacity="0.3" stroke-width="1" />
  <line x1="0" y1="49" x2="100" y2="49" stroke="currentColor" stroke-opacity="0.12" stroke-width="1" />
  <line x1="0" y1="10" x2="100" y2="10" stroke="currentColor" stroke-opacity="0.16" stroke-width="1" />
  {#if activePoints.length}
    <polyline points={pathPoints()} fill="none" stroke="currentColor" stroke-width="3" stroke-linejoin="round" stroke-linecap="round" />
    {#each activePoints as point, index (point.id)}
      <circle cx={pointX(index)} cy={pointY(point.value)} r="3.2" fill={fillColor(point.tone)} stroke="currentColor" stroke-width="1.5">
        {#if point.label}<title>{point.label}</title>{/if}
      </circle>
    {/each}
  {:else}
    <text x="50" y="52" text-anchor="middle" class="fill-current text-[8px] font-black">{emptyLabel}</text>
  {/if}
</svg>

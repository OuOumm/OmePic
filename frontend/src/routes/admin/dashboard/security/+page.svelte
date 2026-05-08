<script lang="ts">
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { adminGetAbuseOverview, adminGetIPBans } from '@/api';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { formatBytes } from '@/utils';
  import type { AdminAbuseOverview, AdminIPBan } from '@/types';

  let overview = $state<AdminAbuseOverview | null>(null);
  let bans = $state<AdminIPBan[]>([]);

  async function load() {
    if (!preferences.adminToken) return;
    [overview, bans] = await Promise.all([adminGetAbuseOverview(preferences.adminToken), adminGetIPBans(preferences.adminToken)]);
  }

  $effect(() => { load(); });
</script>

<svelte:head><title>{t(preferences.language, 'admin.abuseTitle')} · OmePic</title></svelte:head>

<div class="space-y-6">
  <PageTitle eyebrow="Security sketch" title={t(preferences.language, 'admin.abuseTitle')} subtitle={t(preferences.language, 'admin.abuseDescription')} />
  {#if overview}
    <div class="grid gap-4 md:grid-cols-3">
      <div class="border-y-[3px] ink-line py-4"><p class="text-xs uppercase">Uploads</p><p class="text-3xl font-black">{overview.upload_count}</p></div>
      <div class="border-y-[3px] ink-line py-4"><p class="text-xs uppercase">Size</p><p class="text-3xl font-black">{formatBytes(overview.upload_size)}</p></div>
      <div class="border-y-[3px] ink-line py-4"><p class="text-xs uppercase">Bans</p><p class="text-3xl font-black">{overview.active_ip_ban_count}</p></div>
    </div>
    <div class="grid gap-6 lg:grid-cols-2">
      <section>
        <h2 class="border-b-[3px] ink-line pb-2 text-2xl font-black">Top IPs</h2>
        {#each overview.top_ips as item (item.ip_address)}
          <div class="studio-table-row grid grid-cols-[1fr_100px_120px] gap-4 py-3 text-sm"><span class="font-black">{item.ip_address_masked}</span><span>{item.upload_count}</span><span>{formatBytes(item.total_size)}</span></div>
        {/each}
      </section>
      <section>
        <h2 class="border-b-[3px] ink-line pb-2 text-2xl font-black">Banned IPs</h2>
        {#each bans as ban (ban.id)}
          <div class="studio-table-row py-3 text-sm"><p class="font-black">{ban.ip_address_masked}</p><p class="text-[hsl(var(--ink-muted))]">{ban.reason}</p></div>
        {/each}
      </section>
    </div>
  {/if}
</div>


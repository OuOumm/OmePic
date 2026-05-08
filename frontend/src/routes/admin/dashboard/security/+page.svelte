<script lang="ts">
  import { Ban, ShieldCheck, Trash2 } from 'lucide-svelte';
  import IPDetailPanel from '@/components/studio/IPDetailPanel.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { adminCreateIPBan, adminDeleteIPBan, adminDeleteIPBanImages, adminGetAbuseOverview, adminGetIPBans } from '@/api';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import { formatBytes } from '@/utils';
  import type { AdminAbuseOverview, AdminIPBan } from '@/types';

  let overview = $state<AdminAbuseOverview | null>(null);
  let bans = $state<AdminIPBan[]>([]);
  let reason = $state('manual review');
  let durationHours = $state(24);
  let activeIp = $state<string | null>(null);

  async function load() {
    if (!preferences.adminToken) return;
    [overview, bans] = await Promise.all([adminGetAbuseOverview(preferences.adminToken), adminGetIPBans(preferences.adminToken)]);
  }

  async function banIp(ip: string) {
    if (!preferences.adminToken) return;
    await adminCreateIPBan(preferences.adminToken, { ip_address: ip, duration_hours: durationHours, reason });
    toast.success(t(preferences.language, 'common.success'));
    await load();
  }

  async function unban(id: number) {
    if (!preferences.adminToken) return;
    await adminDeleteIPBan(preferences.adminToken, id);
    toast.success(t(preferences.language, 'common.success'));
    await load();
  }

  async function purgeImages(id: number) {
    if (!preferences.adminToken || !confirm('Delete all images for this ban?')) return;
    const result = await adminDeleteIPBanImages(preferences.adminToken, id);
    toast.success(`Deleted ${result.deleted_count}`);
    await load();
  }

  $effect(() => { load(); });
</script>

<svelte:head><title>{t(preferences.language, 'admin.abuseTitle')} · OmePic</title></svelte:head>

<div class="space-y-8">
  <PageTitle eyebrow="Security sketch" title={t(preferences.language, 'admin.abuseTitle')} subtitle={t(preferences.language, 'admin.abuseDescription')} tone="pink" />
  <section class="paper-strip grid gap-4 py-4 md:grid-cols-[1fr_140px]">
    <label class="grid gap-2 text-sm font-black">
      Ban reason
      <input class="studio-input" bind:value={reason} />
    </label>
    <label class="grid gap-2 text-sm font-black">
      Hours
      <input class="studio-input" type="number" min="1" bind:value={durationHours} />
    </label>
  </section>
  {#if overview}
    <div class="grid gap-4 md:grid-cols-3">
      <div class="border-y-[3px] ink-line py-4"><p class="text-xs uppercase">Uploads</p><p class="text-3xl font-black">{overview.upload_count}</p></div>
      <div class="border-y-[3px] ink-line py-4"><p class="text-xs uppercase">Size</p><p class="text-3xl font-black">{formatBytes(overview.upload_size)}</p></div>
      <div class="border-y-[3px] ink-line py-4"><p class="text-xs uppercase">Bans</p><p class="text-3xl font-black">{overview.active_ip_ban_count}</p></div>
    </div>
    <div class="grid gap-8 lg:grid-cols-2">
      <section>
        <h2 class="border-b-[3px] ink-line pb-2 text-2xl font-black">Top IPs</h2>
        {#each overview.top_ips as item (item.ip_address)}
          <div class="studio-table-row grid gap-3 py-3 text-sm md:grid-cols-[1fr_80px_110px_110px] md:items-center">
            <span class="font-black">{item.ip_address_masked}</span>
            <span>{item.upload_count}</span>
            <span>{formatBytes(item.total_size)}</span>
            <button class="studio-button p-2 text-xs" data-tone={item.is_banned ? 'green' : 'danger'} type="button" disabled={item.is_banned} onclick={() => banIp(item.ip_address)}>
              {item.is_banned ? 'Banned' : 'Ban'}
            </button>
          </div>
        {/each}
      </section>
      <section>
        <h2 class="border-b-[3px] ink-line pb-2 text-2xl font-black">Banned IPs</h2>
        {#if bans.length === 0}
          <div class="grid min-h-32 place-items-center border-[3px] border-dashed ink-line"><p class="flex items-center gap-2 font-black"><ShieldCheck class="size-5" />No active bans</p></div>
        {/if}
        {#each bans as ban (ban.id)}
          <div class="studio-table-row grid gap-3 py-3 text-sm md:grid-cols-[1fr_auto] md:items-start">
            <div><p class="font-black">{ban.ip_address_masked}</p><p class="text-[hsl(var(--ink-muted))]">{ban.reason}</p><p class="text-xs text-[hsl(var(--ink-muted))]">{ban.expires_at ?? 'never expires'}</p></div>
            <div class="flex gap-2">
              <button class="studio-button p-2" type="button" onclick={() => unban(ban.id)} aria-label="unban"><Ban class="size-4" /></button>
              <button class="studio-button p-2" data-tone="danger" type="button" onclick={() => purgeImages(ban.id)} aria-label="purge"><Trash2 class="size-4" /></button>
            </div>
          </div>
        {/each}
      </section>
    </div>
  {/if}
  <IPDetailPanel ip={activeIp} {reason} {durationHours} onClose={() => (activeIp = null)} onChanged={load} />
</div>

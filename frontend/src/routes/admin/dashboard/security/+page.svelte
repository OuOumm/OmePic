<script lang="ts">
  import { Ban, ShieldCheck, Trash2 } from 'lucide-svelte';
  import BanIPDialog from '@/components/studio/BanIPDialog.svelte';
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
  let activeIp = $state<string | null>(null);
  let banTarget = $state<{ ip: string; label?: string } | null>(null);
  let banning = $state(false);

  const topIps = $derived(Array.isArray(overview?.top_ips) ? overview.top_ips : []);
  const safeBans = $derived(Array.isArray(bans) ? bans : []);

  async function load() {
    if (!preferences.adminToken) return;
    const [nextOverview, nextBans] = await Promise.all([adminGetAbuseOverview(preferences.adminToken), adminGetIPBans(preferences.adminToken)]);
    overview = nextOverview ? { ...nextOverview, top_ips: Array.isArray(nextOverview.top_ips) ? nextOverview.top_ips : [] } : null;
    bans = Array.isArray(nextBans) ? nextBans : [];
  }

  async function banIp(input: { ip: string; reason: string; durationHours: number | null }) {
    if (!preferences.adminToken) return;
    banning = true;
    try {
      await adminCreateIPBan(preferences.adminToken, { ip_address: input.ip, duration_hours: input.durationHours, reason: input.reason });
      toast.success(t(preferences.language, 'common.success'));
      banTarget = null;
      await load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      banning = false;
    }
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
  {#if overview}
    <div class="grid gap-4 md:grid-cols-3">
      <div class="border-y-[3px] ink-line py-4"><p class="text-xs uppercase">Uploads</p><p class="text-3xl font-black">{overview.upload_count}</p></div>
      <div class="border-y-[3px] ink-line py-4"><p class="text-xs uppercase">Size</p><p class="text-3xl font-black">{formatBytes(overview.upload_size)}</p></div>
      <div class="border-y-[3px] ink-line py-4"><p class="text-xs uppercase">Bans</p><p class="text-3xl font-black">{overview.active_ip_ban_count}</p></div>
    </div>
    <div class="grid gap-8 lg:grid-cols-2">
      <section>
        <div class="mb-3 flex items-center justify-between border-b-[3px] ink-line pb-2">
          <h2 class="text-2xl font-black">Top IPs</h2>
          <span class="text-xs font-black uppercase text-[hsl(var(--ink-muted))]">Click IP for detail</span>
        </div>
        {#each topIps as item (item.ip_address)}
          <div class="studio-table-row grid gap-3 py-3 text-sm md:grid-cols-[1fr_80px_110px_110px] md:items-center">
            <button class="text-left font-black hover:marker-highlight" type="button" onclick={() => (activeIp = item.ip_address)}>{item.ip_address_masked}</button>
            <span>{item.upload_count}</span>
            <span>{formatBytes(item.total_size)}</span>
            <button class="studio-button p-2 text-xs" data-tone={item.is_banned ? 'green' : 'danger'} type="button" disabled={item.is_banned} onclick={() => (banTarget = { ip: item.ip_address, label: item.ip_address_masked })}>
              {item.is_banned ? 'Banned' : 'Ban'}
            </button>
          </div>
        {/each}
      </section>
      <section>
        <div class="mb-3 flex items-center justify-between border-b-[3px] ink-line pb-2">
          <h2 class="text-2xl font-black">Banned IPs</h2>
          <span class="text-xs font-black uppercase text-[hsl(var(--ink-muted))]">Click IP for detail</span>
        </div>
        {#if safeBans.length === 0}
          <div class="grid min-h-32 place-items-center border-[3px] border-dashed ink-line"><p class="flex items-center gap-2 font-black"><ShieldCheck class="size-5" />No active bans</p></div>
        {/if}
        {#each safeBans as ban (ban.id)}
          <div class="studio-table-row grid gap-3 py-3 text-sm md:grid-cols-[1fr_auto] md:items-start">
            <div><button class="font-black hover:marker-highlight" type="button" onclick={() => (activeIp = ban.ip_address)}>{ban.ip_address_masked}</button><p class="text-[hsl(var(--ink-muted))]">{ban.reason}</p><p class="text-xs text-[hsl(var(--ink-muted))]">{ban.expires_at ?? 'never expires'}</p></div>
            <div class="flex gap-2">
              <button class="studio-button p-2" type="button" onclick={() => unban(ban.id)} aria-label="unban"><Ban class="size-4" /></button>
              <button class="studio-button p-2" data-tone="danger" type="button" onclick={() => purgeImages(ban.id)} aria-label="purge"><Trash2 class="size-4" /></button>
            </div>
          </div>
        {/each}
      </section>
    </div>
  {/if}
  <IPDetailPanel ip={activeIp} onClose={() => (activeIp = null)} onChanged={load} onBan={(target) => (banTarget = target)} />
  <BanIPDialog target={banTarget} busy={banning} onClose={() => (banTarget = null)} onConfirm={banIp} />
</div>

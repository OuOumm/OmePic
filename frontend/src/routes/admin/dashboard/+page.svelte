<script lang="ts">
  import { Loader2, Lock } from 'lucide-svelte';
  import MetricStrip from '@/components/studio/MetricStrip.svelte';
  import PageTitle from '@/components/studio/PageTitle.svelte';
  import { adminGetStatus, adminGetSystemSettings, adminLogin } from '@/api';
  import { t } from '@/i18n';
  import { preferences, setAdminToken } from '@/stores/preferences.svelte';
  import { formatBytes } from '@/utils';
  import type { AdminStatus, AdminSystemSettings } from '@/types';

  let password = $state('');
  let loading = $state(false);
  let status = $state<AdminStatus | null>(null);
  let system = $state<AdminSystemSettings | null>(null);
  let error = $state<string | null>(null);

  async function login() {
    loading = true;
    error = null;
    try {
      const token = await adminLogin(password);
      setAdminToken(token);
      await loadData();
    } catch (err) {
      error = err instanceof Error ? err.message : t(preferences.language, 'admin.loginError');
    } finally {
      loading = false;
    }
  }

  async function loadData() {
    if (!preferences.adminToken) return;
    loading = true;
    try {
      [status, system] = await Promise.all([adminGetStatus(preferences.adminToken), adminGetSystemSettings(preferences.adminToken)]);
    } catch (err) {
      error = err instanceof Error ? err.message : t(preferences.language, 'common.error');
    } finally {
      loading = false;
    }
  }

  $effect(() => { loadData(); });
</script>

<svelte:head><title>{t(preferences.language, 'admin.statusTitle')} · OmePic</title></svelte:head>

{#if !preferences.adminToken}
  <div class="mx-auto max-w-md pt-12">
    <div class="studio-panel p-6 rotate-[-0.5deg]">
      <Lock class="mb-4 size-9" />
      <h1 class="text-3xl font-black">{t(preferences.language, 'admin.login')}</h1>
      <label class="mt-5 grid gap-2 font-black">
        {t(preferences.language, 'admin.password')}
        <input class="studio-input" type="password" bind:value={password} onkeydown={(event) => event.key === 'Enter' && login()} />
      </label>
      {#if error}<p class="mt-3 text-sm font-bold text-[hsl(var(--danger))]">{error}</p>{/if}
      <button class="studio-button mt-5 w-full" data-tone="primary" type="button" onclick={login} disabled={loading || !password}>
        {#if loading}<Loader2 class="size-4 animate-spin" />{/if}
        {t(preferences.language, 'admin.loginBtn')}
      </button>
    </div>
  </div>
{:else}
  <div class="space-y-7">
    <PageTitle eyebrow="Blueprint" title={t(preferences.language, 'admin.statusTitle')} subtitle={t(preferences.language, 'admin.statusDescription')} />
    {#if loading}<p class="font-black">{t(preferences.language, 'common.loading')}</p>{/if}
    {#if error}<div class="studio-panel p-4 text-[hsl(var(--danger))]">{error}</div>{/if}
    {#if status}
      <div class="grid gap-4 md:grid-cols-4">
        <MetricStrip label={t(preferences.language, 'admin.totalImages')} value={status.total_images} tone="yellow" />
        <MetricStrip label={t(preferences.language, 'admin.totalSize')} value={formatBytes(status.total_storage_size)} tone="blue" />
        <MetricStrip label={t(preferences.language, 'admin.todayUploads')} value={status.today_uploads} tone="green" />
        <MetricStrip label={t(preferences.language, 'admin.uniqueTokens')} value={status.unique_tokens} tone="pink" />
      </div>
    {/if}
    {#if system}
      <section class="blueprint-grid border-[3px] ink-line p-5">
        <h2 class="text-2xl font-black">Runtime map</h2>
        <div class="mt-4 grid gap-3 md:grid-cols-2">
          <p><b>HTTP</b> · {system.readonly.environment.http_addr}</p>
          <p><b>DB</b> · {system.readonly.environment.database_path}</p>
          <p><b>Redis</b> · {system.readonly.environment.redis_configured ? t(preferences.language, 'common.enabled') : t(preferences.language, 'common.disabled')}</p>
          <p><b>Health</b> · {system.readonly.service.health}</p>
        </div>
      </section>
    {/if}
  </div>
{/if}


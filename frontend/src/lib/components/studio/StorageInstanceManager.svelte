<script lang="ts">
  import { Activity, Edit3, Plus, RefreshCw, Save, Trash2, X } from 'lucide-svelte';
  import {
    adminCheckAllStorageHealth,
    adminCheckStorageHealth,
    adminCreateStorageInstance,
    adminDeleteStorageInstance,
    adminGetStorageHealth,
    adminGetStorageHealthHistory,
    adminSetDefaultStorage,
    adminUpdateStorageInstance,
  } from '@/api';
  import { attachAccessibleDialog } from '@/actions/accessible-dialog';
  import ConfirmDialog from './ConfirmDialog.svelte';
  import LineChart from './LineChart.svelte';
  import { t } from '@/i18n';
  import PageTitle from './PageTitle.svelte';
  import { preferences } from '@/stores/preferences.svelte';
  import { formatDate, isAbortError } from '@/utils';
  import { runAsyncAction, toastApiError } from '@/ui-errors';
  import type { AdminConfig, AdminStorageHealthCheck, StorageInstance } from '@/types';

  type Props = {
    config: AdminConfig;
    onChange: (config: AdminConfig) => void;
  };

  const healthTrendHours = 24;
  const healthAutoCheckIntervalMs = 5 * 60 * 1000;

  const blank: StorageInstance = {
    storage_key: '',
    name: '',
    is_default: false,
    storage_backend: 'local',
    local_storage_path: '',
    s3_endpoint: '',
    s3_region: '',
    s3_bucket: '',
    s3_access_key: '',
    s3_secret_key: '',
    s3_use_ssl: true,
    s3_force_path_style: false,
    webdav_url: '',
    webdav_user: '',
    webdav_pass: '',
  };

  let { config, onChange }: Props = $props();
  let form = $state<StorageInstance>({ ...blank });
  let editingKey = $state<string | null>(null);
  let editorOpen = $state(false);
  let deleteTarget = $state<StorageInstance | null>(null);
  let busyKey = $state('');
  let saving = $state(false);
  let healthChecks = $state.raw<AdminStorageHealthCheck[]>([]);
  let checkingStorageKey = $state<string | null>(null);
  let healthDetail = $state<AdminStorageHealthCheck | null>(null);
  let healthHistory = $state<Record<string, AdminStorageHealthCheck[]>>({});

  const healthByKey = $derived.by(() => {
    const byKey: Record<string, AdminStorageHealthCheck> = {};
    for (const check of healthChecks) byKey[check.storage_key] = check;
    return byKey;
  });

  function closeEditor() {
    editingKey = null;
    form = { ...blank };
    editorOpen = false;
  }

  function startCreate() {
    editingKey = null;
    form = { ...blank };
    editorOpen = true;
  }

  function startEdit(instance: StorageInstance) {
    editingKey = instance.storage_key;
    form = { ...blank, ...instance, s3_secret_key: '', webdav_pass: '' };
    editorOpen = true;
  }

  function payload() {
    const base: Partial<StorageInstance> = {
      storage_key: form.storage_key.trim(),
      name: form.name.trim(),
      is_default: form.is_default,
      storage_backend: form.storage_backend,
    };
    if (form.storage_backend === 'local') {
      base.local_storage_path = form.local_storage_path?.trim();
    }
    if (form.storage_backend === 's3') {
      base.s3_endpoint = form.s3_endpoint?.trim();
      base.s3_region = form.s3_region?.trim();
      base.s3_bucket = form.s3_bucket?.trim();
      base.s3_access_key = form.s3_access_key?.trim();
      if (!editingKey || form.s3_secret_key?.trim()) base.s3_secret_key = form.s3_secret_key?.trim();
      base.s3_use_ssl = form.s3_use_ssl;
      base.s3_force_path_style = form.s3_force_path_style;
    }
    if (form.storage_backend === 'webdav') {
      base.webdav_url = form.webdav_url?.trim();
      base.webdav_user = form.webdav_user?.trim();
      if (!editingKey || form.webdav_pass?.trim()) base.webdav_pass = form.webdav_pass?.trim();
    }
    return base;
  }

  function applyHealthChecks(next: AdminStorageHealthCheck[]) {
    healthChecks = next;
    if (healthDetail) healthDetail = next.find((item) => item.storage_key === healthDetail?.storage_key) ?? healthDetail;
  }

  function healthHistorySince() {
    return new Date(Date.now() - healthTrendHours * 60 * 60 * 1000).toISOString();
  }

  async function loadHealthHistory(storageKey: string, signal?: AbortSignal) {
    const token = preferences.adminToken;
    if (!token) return;
    const history = await adminGetStorageHealthHistory(token, storageKey, healthHistorySince(), signal);
    if (signal?.aborted) return;
    healthHistory = { ...healthHistory, [storageKey]: Array.isArray(history) ? history : [] };
  }

  async function loadAllHealthHistory(signal?: AbortSignal) {
    const token = preferences.adminToken;
    if (!token) return;
    const entries = await Promise.all(config.storage_configs.map(async (item) => {
      const history = await adminGetStorageHealthHistory(token, item.storage_key, healthHistorySince(), signal);
      return [item.storage_key, Array.isArray(history) ? history : []] as const;
    }));
    if (signal?.aborted) return;
    healthHistory = { ...healthHistory, ...Object.fromEntries(entries) };
  }

  async function loadHealth(signal?: AbortSignal) {
    const token = preferences.adminToken;
    if (!token) return;
    try {
      const checks = await adminGetStorageHealth(token, signal);
      if (signal?.aborted) return;
      applyHealthChecks(Array.isArray(checks) ? checks : []);
      await loadAllHealthHistory(signal);
    } catch (err) {
      if (isAbortError(err)) return;
      toastApiError(err, preferences.language);
    }
  }

  async function checkOneStorage(storageKey: string) {
    const token = preferences.adminToken;
    if (!token) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (checkingStorageKey = value ? storageKey : null),
      successMessage: t(preferences.language, 'admin.storageHealthCheckSuccess'),
      action: () => adminCheckStorageHealth(token, storageKey),
      onSuccess: (check) => {
        const byKey: Record<string, AdminStorageHealthCheck> = {};
        for (const item of healthChecks) byKey[item.storage_key] = item;
        byKey[check.storage_key] = check;
        applyHealthChecks(Object.values(byKey));
        healthDetail = check;
        void loadHealthHistory(check.storage_key);
      },
    });
  }

  async function checkAllStorage() {
    const token = preferences.adminToken;
    if (!token) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (checkingStorageKey = value ? '*' : null),
      successMessage: t(preferences.language, 'admin.storageHealthCheckSuccess'),
      action: () => adminCheckAllStorageHealth(token),
      onSuccess: async (checks) => {
        applyHealthChecks(Array.isArray(checks) ? checks : []);
        await loadAllHealthHistory();
      },
    });
  }

  async function save() {
    const token = preferences.adminToken;
    if (!token || !form.name.trim() || (editingKey && !form.storage_key.trim())) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (saving = value),
      successMessage: t(preferences.language, 'common.success'),
      action: () => editingKey
        ? adminUpdateStorageInstance(token, editingKey, payload())
        : adminCreateStorageInstance(token, payload()),
      onSuccess: async (next) => {
        onChange(next);
        closeEditor();
        await loadHealth();
      },
    });
  }

  async function setDefault(storageKey: string) {
    const token = preferences.adminToken;
    if (!token) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (busyKey = value ? storageKey : ''),
      successMessage: t(preferences.language, 'common.success'),
      action: () => adminSetDefaultStorage(token, storageKey),
      onSuccess: onChange,
    });
  }

  async function remove(instance: StorageInstance) {
    const token = preferences.adminToken;
    if (!token || instance.is_default) return;
    await runAsyncAction({
      language: preferences.language,
      setBusy: (value) => (busyKey = value ? instance.storage_key : ''),
      successMessage: t(preferences.language, 'common.success'),
      action: () => adminDeleteStorageInstance(token, instance.storage_key),
      onSuccess: async (next) => {
        onChange(next);
        deleteTarget = null;
        if (editingKey === instance.storage_key) closeEditor();
        await loadHealth();
      },
    });
  }

  function healthLabel(check: AdminStorageHealthCheck | undefined) {
    if (!check || !check.last_check_at) return t(preferences.language, 'admin.storageHealthUnknown');
    return check.status === 'healthy' ? t(preferences.language, 'admin.storageHealthHealthy') : t(preferences.language, 'admin.storageHealthUnhealthy');
  }

  function healthTone(check: AdminStorageHealthCheck | undefined) {
    if (!check || !check.last_check_at) return 'blue';
    return check.status === 'healthy' ? 'green' : 'danger';
  }

  async function autoCheckAllStorage() {
    const token = preferences.adminToken;
    if (!token || checkingStorageKey !== null) return;
    checkingStorageKey = '*';
    try {
      const checks = await adminCheckAllStorageHealth(token);
      applyHealthChecks(Array.isArray(checks) ? checks : []);
      await loadAllHealthHistory();
    } catch {
      // 自动检测静默失败，避免后台页面每 5 分钟弹出错误提示。
    } finally {
      checkingStorageKey = null;
    }
  }

  function trendSamples(storageKey: string) {
    return healthHistory[storageKey] ?? [];
  }

  function trendChartPoints(storageKey: string) {
    return trendSamples(storageKey).map((sample) => ({
      id: sample.id,
      value: Math.max(0, sample.latency_ms),
      tone: sample.status === 'healthy' ? 'green' as const : 'pink' as const,
      label: `${sample.latency_ms} ms`,
    }));
  }


  $effect(() => {
    const controller = new AbortController();
    void loadHealth(controller.signal);
    const timer = window.setInterval(() => void autoCheckAllStorage(), healthAutoCheckIntervalMs);
    return () => {
      controller.abort();
      window.clearInterval(timer);
    };
  });
</script>

<section class="grid min-w-0 gap-6 overflow-hidden">
  <div class="min-w-0">
    <PageTitle eyebrow={t(preferences.language, 'admin.submenuStorage')} title={t(preferences.language, 'admin.storageInstances')} subtitle={t(preferences.language, 'admin.settingsDescription')} tone="blue" />
    <div class="mt-6 mb-4 flex flex-col gap-3 border-b-[3px] ink-line pb-3 sm:flex-row sm:items-center sm:justify-between">
      <p class="text-sm font-bold text-[hsl(var(--ink-muted))]">{t(preferences.language, 'admin.storageHealthInlineHint')}</p>
      <div class="flex flex-col gap-2 sm:flex-row">
        <button class="studio-button p-2" type="button" disabled={checkingStorageKey !== null} onclick={checkAllStorage} title={t(preferences.language, 'admin.storageHealthCheckAll')} aria-label={t(preferences.language, 'admin.storageHealthCheckAll')}><RefreshCw class={`size-4 ${checkingStorageKey === '*' ? 'animate-spin' : ''}`} /></button>
        <button class="studio-button" data-tone="blue" type="button" onclick={startCreate}><Plus class="size-4" />{t(preferences.language, 'admin.storageNew')}</button>
      </div>
    </div>
    <div class="w-full min-w-0 max-w-full touch-pan-x overflow-x-auto overscroll-x-contain [-webkit-overflow-scrolling:touch]">
      <table class="w-full min-w-[820px] border-collapse text-sm">
        <thead>
          <tr class="border-b-[3px] ink-line text-left text-xs font-black uppercase tracking-[0.12em] text-[hsl(var(--ink-muted))]">
            <th class="px-2 py-2" scope="col">{t(preferences.language, 'admin.storageName')}</th>
            <th class="w-[180px] px-2 py-2" scope="col">{t(preferences.language, 'admin.storageKey')}</th>
            <th class="w-[110px] px-2 py-2" scope="col">{t(preferences.language, 'admin.storageBackend')}</th>
            <th class="w-[130px] px-2 py-2" scope="col">{t(preferences.language, 'admin.storageHealthStatus')}</th>
            <th class="w-[230px] px-2 py-2 text-right" scope="col">{t(preferences.language, 'admin.imagesTableActions')}</th>
          </tr>
        </thead>
        <tbody>
          {#each config.storage_configs as item (item.storage_key)}
            {@const health = healthByKey[item.storage_key]}
            <tr class="studio-table-row align-middle">
              <th class="min-w-0 px-2 py-2 text-left font-normal" scope="row"><span class="block truncate font-black">{item.name}</span></th>
              <td class="min-w-0 px-2 py-2"><span class="block truncate text-sm font-semibold text-[hsl(var(--ink-muted))]">{item.storage_key}</span></td>
              <td class="px-2 py-2 font-black uppercase">{item.storage_backend}</td>
              <td class="px-2 py-2">
                <button class="studio-button min-w-20 justify-center px-2.5 py-1.5 text-xs" data-tone={healthTone(health)} type="button" onclick={() => (healthDetail = health ?? { id: 0, storage_key: item.storage_key, status: 'unknown', last_check_at: '', latency_ms: 0, error_message: t(preferences.language, 'admin.storageHealthNotChecked'), consecutive_failures: 0, created_at: '', updated_at: '' })}>
                  <Activity class="size-3.5" />{healthLabel(health)}
                </button>
              </td>
              <td class="px-2 py-2">
                <div class="flex flex-nowrap justify-end gap-2">
                  <button class="studio-button p-2" type="button" disabled={checkingStorageKey !== null} onclick={() => checkOneStorage(item.storage_key)} title={t(preferences.language, 'admin.storageHealthCheckOne')} aria-label={t(preferences.language, 'admin.storageHealthCheckOne')}><RefreshCw class={`size-4 ${checkingStorageKey === item.storage_key ? 'animate-spin' : ''}`} /></button>
                  <button class="studio-button px-2 py-1.5" type="button" onclick={() => startEdit(item)} aria-label={t(preferences.language, 'announcement.edit')}><Edit3 class="size-4" /></button>
                  <button class="studio-button px-2 py-1.5 text-xs" data-tone="green" type="button" disabled={item.is_default || busyKey === item.storage_key} onclick={() => setDefault(item.storage_key)}>{t(preferences.language, 'common.default')}</button>
                  <button class="studio-button px-2 py-1.5" data-tone="danger" type="button" disabled={item.is_default || busyKey === item.storage_key} onclick={() => (deleteTarget = item)} aria-label={t(preferences.language, 'common.delete')}><Trash2 class="size-4" /></button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>

  {#if editorOpen}
    <div class="fixed inset-0 z-50 grid place-items-center p-4" role="dialog" aria-modal="true" aria-labelledby="storage-editor-title" tabindex="-1" {@attach attachAccessibleDialog(() => ({ onClose: closeEditor }))}>
      <button class="absolute inset-0 cursor-default bg-[hsl(var(--ink))]/35" type="button" onclick={closeEditor} aria-label={t(preferences.language, 'common.cancel')}></button>
      <form class="studio-panel relative max-h-[calc(100dvh-3rem)] w-full max-w-2xl overflow-y-auto p-5 rotate-[0.25deg]" onsubmit={(event) => { event.preventDefault(); save(); }}>
        <div class="mb-4 flex items-center justify-between border-b-2 ink-line pb-2">
          <h2 id="storage-editor-title" class="text-2xl font-black">{editingKey ? t(preferences.language, 'admin.storageEdit') : t(preferences.language, 'admin.storageCreate')}</h2>
          <button class="studio-button p-2" type="button" onclick={closeEditor} aria-label={t(preferences.language, 'common.cancel')}><X class="size-4" /></button>
        </div>
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="grid gap-2 text-sm font-black">
            {t(preferences.language, 'admin.storageKey')}
            <input class="studio-input" bind:value={form.storage_key} disabled={!!editingKey} />
          </label>
          <label class="grid gap-2 text-sm font-black">
            {t(preferences.language, 'admin.storageName')}
            <input class="studio-input" bind:value={form.name} />
          </label>
        </div>
        <label class="mt-4 grid gap-2 text-sm font-black">
          {t(preferences.language, 'admin.storageBackend')}
          <select class="studio-input" bind:value={form.storage_backend}>
            <option value="local">local</option>
            <option value="s3">s3</option>
            <option value="webdav">webdav</option>
          </select>
        </label>

        {#if form.storage_backend === 'local'}
          <label class="mt-4 grid gap-2 text-sm font-black">
            {t(preferences.language, 'admin.storageLocalPath')}
            <input class="studio-input" bind:value={form.local_storage_path} />
          </label>
        {:else if form.storage_backend === 's3'}
          <div class="mt-4 grid gap-3">
            <label class="grid gap-2 text-sm font-black">{t(preferences.language, 'admin.storageEndpoint')}<input class="studio-input" bind:value={form.s3_endpoint} /></label>
            <div class="grid gap-3 sm:grid-cols-2">
              <label class="grid gap-2 text-sm font-black">{t(preferences.language, 'admin.storageRegion')}<input class="studio-input" bind:value={form.s3_region} /></label>
              <label class="grid gap-2 text-sm font-black">{t(preferences.language, 'admin.storageBucket')}<input class="studio-input" bind:value={form.s3_bucket} /></label>
            </div>
            <label class="grid gap-2 text-sm font-black">{t(preferences.language, 'admin.storageAccessKey')}<input class="studio-input" bind:value={form.s3_access_key} /></label>
            <label class="grid gap-2 text-sm font-black">{t(preferences.language, 'admin.storageSecretKey')}<input class="studio-input" type="password" autocomplete="new-password" bind:value={form.s3_secret_key} /></label>
            <div class="grid gap-2 text-sm font-black">
              <label class="flex items-center gap-3"><input type="checkbox" bind:checked={form.s3_use_ssl} />{t(preferences.language, 'admin.storageUseSsl')}</label>
              <label class="flex items-center gap-3"><input type="checkbox" bind:checked={form.s3_force_path_style} />{t(preferences.language, 'admin.storageForcePathStyle')}</label>
            </div>
          </div>
        {:else}
          <div class="mt-4 grid gap-3">
            <label class="grid gap-2 text-sm font-black">{t(preferences.language, 'admin.storageWebdavUrl')}<input class="studio-input" bind:value={form.webdav_url} /></label>
            <label class="grid gap-2 text-sm font-black">{t(preferences.language, 'admin.storageUser')}<input class="studio-input" bind:value={form.webdav_user} /></label>
            <label class="grid gap-2 text-sm font-black">{t(preferences.language, 'admin.storagePassword')}<input class="studio-input" type="password" autocomplete="new-password" bind:value={form.webdav_pass} /></label>
          </div>
        {/if}

        <label class="mt-4 flex items-center gap-3 border-y-2 ink-line py-3 font-black">
          <input type="checkbox" bind:checked={form.is_default} />
          {t(preferences.language, 'common.default')}
        </label>
        <button class="studio-button mt-5 w-full" data-tone="primary" type="submit" disabled={saving || !form.name.trim() || Boolean(editingKey && !form.storage_key.trim())}><Save class="size-4" />{t(preferences.language, 'common.save')}</button>
      </form>
    </div>
  {/if}

  {#if healthDetail}
    <div class="fixed inset-0 z-50 grid place-items-center p-4" role="dialog" aria-modal="true" aria-labelledby="storage-health-title" tabindex="-1" {@attach attachAccessibleDialog(() => ({ onClose: () => (healthDetail = null) }))}>
      <button class="absolute inset-0 cursor-default bg-[hsl(var(--ink))]/35" type="button" onclick={() => (healthDetail = null)} aria-label={t(preferences.language, 'common.close')}></button>
      <section class="studio-panel relative max-h-[calc(100dvh-3rem)] w-full max-w-2xl overflow-y-auto p-5 rotate-[-0.25deg]">
        <div class="mb-4 flex items-center justify-between border-b-2 ink-line pb-2">
          <div>
            <p class="tape-label" style={`background:hsl(var(${healthDetail.status === 'healthy' ? '--marker-green' : '--marker-pink'}))`}>{healthLabel(healthDetail)}</p>
            <h2 id="storage-health-title" class="mt-3 text-2xl font-black">{healthDetail.storage_key}</h2>
          </div>
          <button class="studio-button p-2" type="button" onclick={() => (healthDetail = null)} aria-label={t(preferences.language, 'common.close')}><X class="size-4" /></button>
        </div>
        <dl class="grid gap-3 text-sm md:grid-cols-2">
          <div class="border-2 ink-line bg-[hsl(var(--paper-deep))] p-3"><dt class="font-black">{t(preferences.language, 'admin.storageHealthLastCheck')}</dt><dd>{healthDetail.last_check_at ? formatDate(healthDetail.last_check_at, preferences.language) : t(preferences.language, 'admin.storageHealthNotChecked')}</dd></div>
          <div class="border-2 ink-line bg-[hsl(var(--paper-deep))] p-3"><dt class="font-black">{t(preferences.language, 'admin.storageHealthLatency')}</dt><dd>{healthDetail.latency_ms} ms</dd></div>
          <div class="border-2 ink-line bg-[hsl(var(--paper-deep))] p-3"><dt class="font-black">{t(preferences.language, 'admin.storageHealthFailuresLabel')}</dt><dd>{healthDetail.consecutive_failures}</dd></div>
          <div class="border-2 ink-line bg-[hsl(var(--paper-deep))] p-3"><dt class="font-black">{t(preferences.language, 'admin.storageHealthUpdatedAt')}</dt><dd>{healthDetail.updated_at ? formatDate(healthDetail.updated_at, preferences.language) : '—'}</dd></div>
        </dl>
        <div class="mt-4 border-2 ink-line bg-[hsl(var(--paper))] p-4">
          <div class="mb-2 flex items-center justify-between gap-2">
            <h3 class="font-black">{t(preferences.language, 'admin.storageHealthTrend')}</h3>
            <span class="text-xs font-bold text-[hsl(var(--ink-muted))]">{t(preferences.language, 'admin.storageHealthSessionTrend')}</span>
          </div>
          <LineChart
            points={trendChartPoints(healthDetail.storage_key)}
            emptyLabel={t(preferences.language, 'admin.storageHealthNoTrend')}
            ariaLabel={t(preferences.language, 'admin.storageHealthTrend')}
          />
        </div>
        {#if healthDetail.error_message}
          <div class="mt-4 border-2 ink-line bg-[hsl(var(--marker-pink))] p-3 text-[hsl(var(--marker-ink))]"><h3 class="font-black">{t(preferences.language, 'admin.storageHealthError')}</h3><p class="mt-2 break-words text-sm font-bold">{healthDetail.error_message}</p></div>
        {/if}
        <div class="mt-5 flex justify-end">
          <button class="studio-button p-2" type="button" disabled={checkingStorageKey !== null || !healthDetail?.storage_key} onclick={() => healthDetail?.storage_key && checkOneStorage(healthDetail.storage_key)} title={t(preferences.language, 'admin.storageHealthCheckOne')} aria-label={t(preferences.language, 'admin.storageHealthCheckOne')}><RefreshCw class={`size-4 ${checkingStorageKey === healthDetail?.storage_key ? 'animate-spin' : ''}`} /></button>
        </div>
      </section>
    </div>
  {/if}

  <ConfirmDialog
    open={deleteTarget !== null}
    title={`${t(preferences.language, 'common.delete')} ${deleteTarget?.name ?? ''}?`}
    description={deleteTarget?.storage_key ?? ''}
    confirmLabel={t(preferences.language, 'common.delete')}
    cancelLabel={t(preferences.language, 'common.cancel')}
    busy={Boolean(deleteTarget && busyKey === deleteTarget.storage_key)}
    onClose={() => (deleteTarget = null)}
    onConfirm={() => deleteTarget && remove(deleteTarget)}
  />
</section>

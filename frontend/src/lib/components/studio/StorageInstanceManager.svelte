<script lang="ts">
  import { Edit3, Plus, Save, Trash2, X } from 'lucide-svelte';
  import {
    adminCreateStorageInstance,
    adminDeleteStorageInstance,
    adminSetDefaultStorage,
    adminUpdateStorageInstance,
  } from '@/api';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import type { AdminConfig, StorageInstance } from '@/types';

  type Props = {
    config: AdminConfig;
    onChange: (config: AdminConfig) => void;
  };

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
  let busyKey = $state('');
  let saving = $state(false);

  function startCreate() {
    editingKey = null;
    form = { ...blank };
  }

  function startEdit(instance: StorageInstance) {
    editingKey = instance.storage_key;
    form = { ...blank, ...instance };
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
      base.s3_secret_key = form.s3_secret_key?.trim();
      base.s3_use_ssl = form.s3_use_ssl;
      base.s3_force_path_style = form.s3_force_path_style;
    }
    if (form.storage_backend === 'webdav') {
      base.webdav_url = form.webdav_url?.trim();
      base.webdav_user = form.webdav_user?.trim();
      base.webdav_pass = form.webdav_pass?.trim();
    }
    return base;
  }

  async function save() {
    if (!preferences.adminToken || !form.storage_key.trim() || !form.name.trim()) return;
    saving = true;
    try {
      const next = editingKey
        ? await adminUpdateStorageInstance(preferences.adminToken, editingKey, payload())
        : await adminCreateStorageInstance(preferences.adminToken, payload());
      onChange(next);
      toast.success(t(preferences.language, 'common.success'));
      startCreate();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      saving = false;
    }
  }

  async function setDefault(storageKey: string) {
    if (!preferences.adminToken) return;
    busyKey = storageKey;
    try {
      onChange(await adminSetDefaultStorage(preferences.adminToken, storageKey));
      toast.success(t(preferences.language, 'common.success'));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      busyKey = '';
    }
  }

  async function remove(instance: StorageInstance) {
    if (!preferences.adminToken || instance.is_default || !confirm(`${t(preferences.language, 'common.delete')} ${instance.name}?`)) return;
    busyKey = instance.storage_key;
    try {
      onChange(await adminDeleteStorageInstance(preferences.adminToken, instance.storage_key));
      toast.success(t(preferences.language, 'common.success'));
      if (editingKey === instance.storage_key) startCreate();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      busyKey = '';
    }
  }
</script>

<section class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_390px]">
  <div>
    <div class="mb-4 flex items-end justify-between border-b-[3px] ink-line pb-3">
      <div>
        <span class="tape-label rotate-[-2deg]">storage</span>
        <h2 class="mt-3 text-3xl font-black">Storage instances</h2>
      </div>
      <button class="studio-button" data-tone="blue" type="button" onclick={startCreate}><Plus class="size-4" />New</button>
    </div>
    <div class="grid gap-2">
      {#each config.storage_configs as item (item.storage_key)}
        <article class="studio-table-row grid gap-3 py-4 md:grid-cols-[1fr_120px_220px] md:items-center">
          <div>
            <div class="flex flex-wrap items-center gap-2">
              <h3 class="text-xl font-black">{item.name}</h3>
              {#if item.is_default}<span class="tape-label rotate-1" style="background:hsl(var(--marker-green))">{t(preferences.language, 'common.default')}</span>{/if}
            </div>
            <p class="text-sm font-semibold text-[hsl(var(--ink-muted))]">{item.storage_key}</p>
          </div>
          <div class="font-black uppercase">{item.storage_backend}</div>
          <div class="flex flex-wrap gap-2 md:justify-end">
            <button class="studio-button p-2" type="button" onclick={() => startEdit(item)} aria-label="edit"><Edit3 class="size-4" /></button>
            <button class="studio-button p-2 text-xs" data-tone="green" type="button" disabled={item.is_default || busyKey === item.storage_key} onclick={() => setDefault(item.storage_key)}>{item.is_default ? 'Default' : 'Default'}</button>
            <button class="studio-button p-2" data-tone="danger" type="button" disabled={item.is_default || busyKey === item.storage_key} onclick={() => remove(item)} aria-label="delete"><Trash2 class="size-4" /></button>
          </div>
        </article>
      {/each}
    </div>
  </div>

  <form class="studio-panel h-fit p-5 rotate-[0.25deg]" onsubmit={(event) => { event.preventDefault(); save(); }}>
    <div class="mb-4 flex items-center justify-between border-b-2 ink-line pb-2">
      <h2 class="text-2xl font-black">{editingKey ? 'Edit storage' : 'Create storage'}</h2>
      {#if editingKey}<button class="studio-button p-2" type="button" onclick={startCreate} aria-label="cancel"><X class="size-4" /></button>{/if}
    </div>
    <div class="grid gap-3 sm:grid-cols-2">
      <label class="grid gap-2 text-sm font-black">
        Key
        <input class="studio-input" bind:value={form.storage_key} disabled={!!editingKey} />
      </label>
      <label class="grid gap-2 text-sm font-black">
        Name
        <input class="studio-input" bind:value={form.name} />
      </label>
    </div>
    <label class="mt-4 grid gap-2 text-sm font-black">
      Backend
      <select class="studio-input" bind:value={form.storage_backend}>
        <option value="local">local</option>
        <option value="s3">s3</option>
        <option value="webdav">webdav</option>
      </select>
    </label>

    {#if form.storage_backend === 'local'}
      <label class="mt-4 grid gap-2 text-sm font-black">
        Local path
        <input class="studio-input" bind:value={form.local_storage_path} />
      </label>
    {:else if form.storage_backend === 's3'}
      <div class="mt-4 grid gap-3">
        <label class="grid gap-2 text-sm font-black">Endpoint<input class="studio-input" bind:value={form.s3_endpoint} /></label>
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="grid gap-2 text-sm font-black">Region<input class="studio-input" bind:value={form.s3_region} /></label>
          <label class="grid gap-2 text-sm font-black">Bucket<input class="studio-input" bind:value={form.s3_bucket} /></label>
        </div>
        <label class="grid gap-2 text-sm font-black">Access key<input class="studio-input" bind:value={form.s3_access_key} /></label>
        <label class="grid gap-2 text-sm font-black">Secret key<input class="studio-input" type="password" bind:value={form.s3_secret_key} /></label>
        <div class="grid gap-2 text-sm font-black">
          <label class="flex items-center gap-3"><input type="checkbox" bind:checked={form.s3_use_ssl} />Use SSL</label>
          <label class="flex items-center gap-3"><input type="checkbox" bind:checked={form.s3_force_path_style} />Force path style</label>
        </div>
      </div>
    {:else}
      <div class="mt-4 grid gap-3">
        <label class="grid gap-2 text-sm font-black">WebDAV URL<input class="studio-input" bind:value={form.webdav_url} /></label>
        <label class="grid gap-2 text-sm font-black">User<input class="studio-input" bind:value={form.webdav_user} /></label>
        <label class="grid gap-2 text-sm font-black">Password<input class="studio-input" type="password" bind:value={form.webdav_pass} /></label>
      </div>
    {/if}

    <label class="mt-4 flex items-center gap-3 border-y-2 ink-line py-3 font-black">
      <input type="checkbox" bind:checked={form.is_default} />
      {t(preferences.language, 'common.default')}
    </label>
    <button class="studio-button mt-5 w-full" data-tone="primary" type="submit" disabled={saving || !form.storage_key.trim() || !form.name.trim()}><Save class="size-4" />{t(preferences.language, 'common.save')}</button>
  </form>
</section>

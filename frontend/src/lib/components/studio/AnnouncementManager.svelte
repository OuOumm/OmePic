<script lang="ts">
  import { Archive, Megaphone, Pencil, Plus, Save, Trash2, X } from 'lucide-svelte';
  import {
    adminArchiveAnnouncement,
    adminCreateAnnouncement,
    adminDeleteAnnouncement,
    adminGetAnnouncements,
    adminUpdateAnnouncement,
  } from '@/api';
  import { t } from '@/i18n';
  import { preferences } from '@/stores/preferences.svelte';
  import { toast } from '@/stores/toast.svelte';
  import type { Announcement, AnnouncementInput } from '@/types';

  const blank: AnnouncementInput = {
    title: '',
    content: '',
    status: 'draft',
    priority: 'normal',
    starts_at: null,
    ends_at: null,
    sort_order: 0,
  };

  let announcements = $state<Announcement[]>([]);
  let form = $state<AnnouncementInput>({ ...blank });
  let editingId = $state<number | null>(null);
  let loading = $state(false);
  let saving = $state(false);

  async function load() {
    if (!preferences.adminToken) return;
    loading = true;
    try {
      announcements = await adminGetAnnouncements(preferences.adminToken);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      loading = false;
    }
  }

  function edit(item: Announcement) {
    editingId = item.id;
    form = {
      title: item.title,
      content: item.content,
      status: item.status ?? 'published',
      priority: item.priority,
      starts_at: item.starts_at,
      ends_at: item.ends_at,
      sort_order: item.sort_order ?? 0,
    };
  }

  function reset() {
    editingId = null;
    form = { ...blank };
  }

  async function save() {
    if (!preferences.adminToken || !form.title.trim() || !form.content.trim()) return;
    saving = true;
    try {
      if (editingId) {
        await adminUpdateAnnouncement(preferences.adminToken, editingId, form);
      } else {
        await adminCreateAnnouncement(preferences.adminToken, form);
      }
      toast.success(t(preferences.language, 'common.success'));
      reset();
      await load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(preferences.language, 'common.error'));
    } finally {
      saving = false;
    }
  }

  async function remove(item: Announcement) {
    if (!preferences.adminToken || !confirm(`${t(preferences.language, 'common.delete')} ${item.title}?`)) return;
    await adminDeleteAnnouncement(preferences.adminToken, item.id);
    toast.success(t(preferences.language, 'common.success'));
    await load();
  }

  async function archive(item: Announcement) {
    if (!preferences.adminToken) return;
    await adminArchiveAnnouncement(preferences.adminToken, item.id);
    toast.success(t(preferences.language, 'common.success'));
    await load();
  }

  $effect(() => { load(); });
</script>

<section class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
  <div>
    <div class="mb-4 flex items-end justify-between border-b-[3px] ink-line pb-3">
      <div>
        <span class="tape-label rotate-[-2deg]">announcements</span>
        <h2 class="mt-3 text-3xl font-black">Public notices</h2>
      </div>
      <button class="studio-button" data-tone="blue" type="button" onclick={reset}><Plus class="size-4" />New</button>
    </div>

    {#if loading}
      <p class="font-black">{t(preferences.language, 'common.loading')}</p>
    {:else if announcements.length === 0}
      <div class="grid min-h-44 place-items-center border-[3px] border-dashed ink-line text-center">
        <div><Megaphone class="mx-auto mb-3 size-8" /><p class="font-black">No announcements yet</p></div>
      </div>
    {:else}
      <div class="grid gap-3">
        {#each announcements as item (item.id)}
          <article class="studio-table-row grid gap-3 py-4 md:grid-cols-[1fr_auto] md:items-start">
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="text-xl font-black">{item.title}</h3>
                <span class="tape-label rotate-1">{item.status ?? 'published'}</span>
                <span class="tape-label rotate-[-1deg]" style="background:hsl(var(--marker-pink))">{item.priority}</span>
              </div>
              <p class="mt-2 max-w-3xl whitespace-pre-wrap text-sm font-semibold text-[hsl(var(--ink-muted))]">{item.content}</p>
            </div>
            <div class="flex gap-2">
              <button class="studio-button p-2" type="button" onclick={() => edit(item)} aria-label="edit"><Pencil class="size-4" /></button>
              <button class="studio-button p-2" type="button" onclick={() => archive(item)} aria-label="archive"><Archive class="size-4" /></button>
              <button class="studio-button p-2" data-tone="danger" type="button" onclick={() => remove(item)} aria-label="delete"><Trash2 class="size-4" /></button>
            </div>
          </article>
        {/each}
      </div>
    {/if}
  </div>

  <form class="studio-panel h-fit p-5 rotate-[0.35deg]" onsubmit={(event) => { event.preventDefault(); save(); }}>
    <div class="mb-4 flex items-center justify-between border-b-2 ink-line pb-2">
      <h2 class="text-2xl font-black">{editingId ? 'Edit notice' : 'Draft notice'}</h2>
      {#if editingId}<button class="studio-button p-2" type="button" onclick={reset} aria-label="cancel"><X class="size-4" /></button>{/if}
    </div>
    <label class="grid gap-2 text-sm font-black">
      Title
      <input class="studio-input" bind:value={form.title} />
    </label>
    <label class="mt-4 grid gap-2 text-sm font-black">
      Content
      <textarea class="studio-input min-h-32" bind:value={form.content}></textarea>
    </label>
    <div class="mt-4 grid gap-3 sm:grid-cols-2">
      <label class="grid gap-2 text-sm font-black">
        Status
        <select class="studio-input" bind:value={form.status}>
          {#each ['draft', 'published', 'archived'] as status (status)}
            <option value={status}>{status}</option>
          {/each}
        </select>
      </label>
      <label class="grid gap-2 text-sm font-black">
        Priority
        <select class="studio-input" bind:value={form.priority}>
          {#each ['normal', 'important', 'urgent'] as priority (priority)}
            <option value={priority}>{priority}</option>
          {/each}
        </select>
      </label>
    </div>
    <div class="mt-4 grid gap-3 sm:grid-cols-2">
      <label class="grid gap-2 text-sm font-black">
        Starts
        <input class="studio-input" type="datetime-local" value={form.starts_at ?? ''} onchange={(event) => (form.starts_at = event.currentTarget.value || null)} />
      </label>
      <label class="grid gap-2 text-sm font-black">
        Ends
        <input class="studio-input" type="datetime-local" value={form.ends_at ?? ''} onchange={(event) => (form.ends_at = event.currentTarget.value || null)} />
      </label>
    </div>
    <label class="mt-4 grid gap-2 text-sm font-black">
      Sort order
      <input class="studio-input" type="number" bind:value={form.sort_order} />
    </label>
    <button class="studio-button mt-5 w-full" data-tone="primary" type="submit" disabled={saving || !form.title.trim() || !form.content.trim()}>
      <Save class="size-4" />{editingId ? t(preferences.language, 'common.save') : 'Publish draft'}
    </button>
  </form>
</section>

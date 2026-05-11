<script lang="ts">
  import { Archive, Eye, Megaphone, Pencil, Plus, Save, Trash2, X } from 'lucide-svelte';
  import {
    adminArchiveAnnouncement,
    adminCreateAnnouncement,
    adminDeleteAnnouncement,
    adminGetAnnouncements,
    adminUpdateAnnouncement,
  } from '@/api';
  import { accessibleDialog } from '@/actions/accessible-dialog';
  import ConfirmDialog from './ConfirmDialog.svelte';
  import { t } from '@/i18n';
  import MarkdownContent from './MarkdownContent.svelte';
  import PageTitle from './PageTitle.svelte';
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

  const statusOptions = [
    { value: 'draft', labelKey: 'announcement.statusDraft' },
    { value: 'published', labelKey: 'announcement.statusPublished' },
    { value: 'archived', labelKey: 'announcement.statusArchived' }
  ];
  const priorityOptions = [
    { value: 'normal', labelKey: 'announcement.priorityNormal' },
    { value: 'important', labelKey: 'announcement.priorityImportant' },
    { value: 'urgent', labelKey: 'announcement.priorityUrgent' }
  ];

  function labelFor(options: { value: string; labelKey: string }[], value: string) {
    return t(preferences.language, options.find((option) => option.value === value)?.labelKey ?? value);
  }

  let announcements = $state<Announcement[]>([]);
  let form = $state<AnnouncementInput>({ ...blank });
  let editingId = $state<number | null>(null);
  let loading = $state(false);
  let saving = $state(false);
  let editorOpen = $state(false);
  let viewingAnnouncement = $state<Announcement | null>(null);
  let deleteTarget = $state<Announcement | null>(null);

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

  function openCreate() {
    editingId = null;
    form = { ...blank };
    editorOpen = true;
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
    editorOpen = true;
  }

  function reset() {
    editingId = null;
    form = { ...blank };
    editorOpen = false;
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
    if (!preferences.adminToken) return;
    await adminDeleteAnnouncement(preferences.adminToken, item.id);
    deleteTarget = null;
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

<section>
  <div>
    <PageTitle eyebrow={t(preferences.language, 'admin.submenuAnnouncements')} title={t(preferences.language, 'announcement.managerTitle')} subtitle={t(preferences.language, 'announcement.managerSubtitle')} tone="pink" />
    <div class="mt-6 mb-4 flex justify-end border-b-[3px] ink-line pb-3">
      <button class="studio-button" data-tone="blue" type="button" onclick={openCreate}><Plus class="size-4" />{t(preferences.language, 'announcement.new')}</button>
    </div>

    {#if loading}
      <p class="font-black">{t(preferences.language, 'common.loading')}</p>
    {:else if announcements.length === 0}
      <div class="grid min-h-44 place-items-center border-[3px] border-dashed ink-line text-center">
        <div><Megaphone class="mx-auto mb-3 size-8" /><p class="font-black">{t(preferences.language, 'announcement.empty')}</p></div>
      </div>
    {:else}
      <div class="w-full min-w-0 max-w-full overflow-x-auto">
        <table class="w-full min-w-[820px] border-collapse text-sm">
          <thead>
            <tr class="border-b-[3px] ink-line text-left text-xs font-black uppercase tracking-[0.12em] text-[hsl(var(--ink-muted))]">
              <th class="w-[220px] px-2 py-2" scope="col">{t(preferences.language, 'announcement.title')}</th>
              <th class="w-[110px] px-2 py-2" scope="col">{t(preferences.language, 'announcement.status')}</th>
              <th class="w-[110px] px-2 py-2" scope="col">{t(preferences.language, 'announcement.priority')}</th>
              <th class="px-2 py-2" scope="col">{t(preferences.language, 'announcement.content')}</th>
              <th class="w-[190px] px-2 py-2 text-right" scope="col">{t(preferences.language, 'admin.imagesTableActions')}</th>
            </tr>
          </thead>
          <tbody>
            {#each announcements as item (item.id)}
              <tr class="studio-table-row align-top">
                <th class="min-w-0 px-2 py-4 text-left font-normal" scope="row"><span class="block truncate text-xl font-black">{item.title}</span></th>
                <td class="px-2 py-4"><span class="tape-label rotate-1">{labelFor(statusOptions, item.status ?? 'published')}</span></td>
                <td class="px-2 py-4"><span class="tape-label rotate-[-1deg]" style="background:hsl(var(--marker-pink))">{labelFor(priorityOptions, item.priority)}</span></td>
                <td class="min-w-0 px-2 py-4">
                  <button class="max-w-3xl text-left text-sm font-semibold text-[hsl(var(--ink-muted))]" type="button" onclick={() => (viewingAnnouncement = item)} aria-label={t(preferences.language, 'announcement.viewDetail')}>
                    <span class="line-clamp-2 whitespace-pre-wrap">{item.content}</span>
                  </button>
                </td>
                <td class="px-2 py-4">
                  <div class="flex justify-end gap-2">
                    <button class="studio-button p-2" type="button" onclick={() => (viewingAnnouncement = item)} aria-label={t(preferences.language, 'announcement.viewDetail')}><Eye class="size-4" /></button>
                    <button class="studio-button p-2" type="button" onclick={() => edit(item)} aria-label={t(preferences.language, 'announcement.edit')}><Pencil class="size-4" /></button>
                    <button class="studio-button p-2" type="button" onclick={() => archive(item)} aria-label={t(preferences.language, 'announcement.archive')}><Archive class="size-4" /></button>
                    <button class="studio-button p-2" data-tone="danger" type="button" onclick={() => (deleteTarget = item)} aria-label={t(preferences.language, 'common.delete')}><Trash2 class="size-4" /></button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>

  {#if viewingAnnouncement}
    <div class="fixed inset-0 z-50 grid place-items-center p-4" role="dialog" aria-modal="true" aria-labelledby="announcement-detail-title" tabindex="-1" use:accessibleDialog={{ onClose: () => (viewingAnnouncement = null) }}>
      <button class="absolute inset-0 cursor-default bg-[hsl(var(--ink))]/35 backdrop-blur-[2px]" type="button" aria-label={t(preferences.language, 'common.cancel')} onclick={() => (viewingAnnouncement = null)}></button>
      <div class="studio-panel relative grid max-h-[calc(100dvh-3rem)] w-full max-w-3xl grid-rows-[auto_minmax(0,1fr)] overflow-hidden p-5 rotate-[0.25deg]">
        <div class="mb-4 flex items-start justify-between gap-3 border-b-[3px] ink-line pb-3">
          <div class="min-w-0">
            <span class="tape-label rotate-[-2deg]" style="background:hsl(var(--marker-pink))">{labelFor(priorityOptions, viewingAnnouncement.priority)}</span>
            <h2 id="announcement-detail-title" class="mt-3 break-all text-3xl font-black">{viewingAnnouncement.title}</h2>
          </div>
          <button class="studio-button p-2" type="button" onclick={() => (viewingAnnouncement = null)} aria-label={t(preferences.language, 'common.cancel')}><X class="size-4" /></button>
        </div>
        <div class="min-h-0 overflow-y-auto pr-1">
          <MarkdownContent content={viewingAnnouncement.content} />
        </div>
      </div>
    </div>
  {/if}

  {#if editorOpen}
    <div class="fixed inset-0 z-50 grid place-items-center bg-[hsl(var(--ink)/0.45)] p-4" role="dialog" aria-modal="true" aria-labelledby="announcement-editor-title" tabindex="-1" use:accessibleDialog={{ onClose: reset }}>
      <button class="absolute inset-0 cursor-default" type="button" aria-label={t(preferences.language, 'common.cancel')} onclick={reset}></button>
      <form class="studio-panel relative max-h-[calc(100dvh-3rem)] w-full max-w-2xl overflow-y-auto p-5 rotate-[0.35deg]" onsubmit={(event) => { event.preventDefault(); save(); }}>
        <div class="mb-4 flex items-center justify-between border-b-2 ink-line pb-2">
          <h2 id="announcement-editor-title" class="text-2xl font-black">{editingId ? t(preferences.language, 'announcement.editNotice') : t(preferences.language, 'announcement.draftNotice')}</h2>
          <button class="studio-button p-2" type="button" onclick={reset} aria-label={t(preferences.language, 'common.cancel')}><X class="size-4" /></button>
        </div>
        <label class="grid gap-2 text-sm font-black">
          {t(preferences.language, 'announcement.title')}
          <input class="studio-input" bind:value={form.title} />
        </label>
        <label class="mt-4 grid gap-2 text-sm font-black">
          {t(preferences.language, 'announcement.content')}
          <textarea class="studio-input min-h-32" bind:value={form.content}></textarea>
        </label>
        <div class="mt-4 grid gap-3 sm:grid-cols-2">
          <label class="grid min-w-0 gap-2 text-sm font-black">
            {t(preferences.language, 'announcement.status')}
            <select class="studio-input w-full min-w-0" bind:value={form.status}>
              {#each statusOptions as status (status.value)}
                <option value={status.value}>{t(preferences.language, status.labelKey)}</option>
              {/each}
            </select>
          </label>
          <label class="grid min-w-0 gap-2 text-sm font-black">
            {t(preferences.language, 'announcement.priority')}
            <select class="studio-input w-full min-w-0" bind:value={form.priority}>
              {#each priorityOptions as priority (priority.value)}
                <option value={priority.value}>{t(preferences.language, priority.labelKey)}</option>
              {/each}
            </select>
          </label>
        </div>
        <div class="mt-4 grid gap-3 sm:grid-cols-2">
          <label class="grid min-w-0 gap-2 text-sm font-black">
            {t(preferences.language, 'announcement.starts')}
            <input class="studio-input w-full min-w-0" type="datetime-local" value={form.starts_at ?? ''} onchange={(event) => (form.starts_at = event.currentTarget.value || null)} />
          </label>
          <label class="grid min-w-0 gap-2 text-sm font-black">
            {t(preferences.language, 'announcement.ends')}
            <input class="studio-input w-full min-w-0" type="datetime-local" value={form.ends_at ?? ''} onchange={(event) => (form.ends_at = event.currentTarget.value || null)} />
          </label>
        </div>
        <label class="mt-4 grid gap-2 text-sm font-black">
          {t(preferences.language, 'announcement.sortOrder')}
          <input class="studio-input" type="number" bind:value={form.sort_order} />
        </label>
        <button class="studio-button mt-5 w-full" data-tone="primary" type="submit" disabled={saving || !form.title.trim() || !form.content.trim()}>
          <Save class="size-4" />{editingId ? t(preferences.language, 'common.save') : t(preferences.language, 'announcement.publishDraft')}
        </button>
      </form>
    </div>
  {/if}
  <ConfirmDialog
    open={deleteTarget !== null}
    title={t(preferences.language, 'announcement.deleteConfirm', { title: deleteTarget?.title ?? '' })}
    description={deleteTarget?.content ?? ''}
    confirmLabel={t(preferences.language, 'common.delete')}
    cancelLabel={t(preferences.language, 'common.cancel')}
    onClose={() => (deleteTarget = null)}
    onConfirm={() => deleteTarget && remove(deleteTarget)}
  />
</section>

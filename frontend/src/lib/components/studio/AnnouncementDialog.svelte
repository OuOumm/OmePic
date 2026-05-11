<script lang="ts">
  import { ChevronLeft, ChevronRight, History, Megaphone, X } from 'lucide-svelte';
  import { t } from '@/i18n';
  import { accessibleDialog } from '@/actions/accessible-dialog';
  import { formatDate } from '@/utils';
  import MarkdownContent from './MarkdownContent.svelte';
  import type { Announcement, Language } from '@/types';

  type Props = {
    language: Language;
    announcements: Announcement[];
    open: boolean;
    initialMode?: 'detail' | 'history';
    onClose: () => void;
    onAcknowledge: () => void;
  };

  let { language, announcements, open, initialMode = 'detail', onClose, onAcknowledge }: Props = $props();
  let index = $state(0);
  let mode = $state<'detail' | 'history'>('detail');

  const current = $derived(announcements[index] ?? announcements[0] ?? null);
  const historyTitle = $derived(mode === 'history' ? t(language, 'announcement.allNotices') : current?.title ?? '');

  function priorityColor(priority: Announcement['priority']) {
    if (priority === 'urgent') return 'hsl(var(--danger))';
    if (priority === 'important') return 'hsl(var(--marker-yellow))';
    return 'hsl(var(--marker-blue))';
  }

  function selectAnnouncement(nextIndex: number) {
    index = nextIndex;
    mode = 'detail';
  }

  $effect(() => {
    if (!open) return;
    mode = initialMode;
    if (index >= announcements.length) index = 0;
  });
</script>

{#if open && current}
  <div class="fixed inset-0 z-[85] grid min-h-dvh place-items-center overflow-y-auto bg-[hsl(var(--ink)/0.48)] p-2 backdrop-blur-sm sm:p-6" role="presentation" onclick={(event) => event.target === event.currentTarget && onClose()}>
    <div class="grid max-h-[calc(100dvh-1rem)] w-full max-w-3xl grid-rows-[auto_minmax(0,1fr)_auto] overflow-hidden border-[3px] ink-line bg-[hsl(var(--paper))] shadow-[5px_5px_0_hsl(var(--ink))] sketch-enter sm:max-h-[calc(100dvh-3rem)] sm:shadow-[8px_8px_0_hsl(var(--ink))]" role="dialog" tabindex="-1" aria-modal="true" aria-labelledby="announcement-dialog-title" use:accessibleDialog={{ onClose }}>
      <header class="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-start gap-3 border-b-[3px] ink-line p-4">
        <div class="min-w-0 overflow-hidden">
          <span class="tape-label rotate-[-2deg]" style="background:hsl(var(--marker-pink))">{mode === 'history' ? t(language, 'announcement.history') : t(language, 'announcement.notice')}</span>
          <h2 id="announcement-dialog-title" class="mt-3 truncate text-2xl font-black sm:text-3xl">{historyTitle}</h2>
        </div>
        <button class="studio-button shrink-0 p-2" type="button" onclick={onClose} aria-label={t(language, 'common.close')}><X class="size-4" /></button>
      </header>

      <div class="min-h-0 overflow-y-auto p-4 sm:p-5">
        {#if mode === 'history'}
          <div class="grid gap-3">
            {#each announcements as item, itemIndex (item.id)}
              <button class="studio-table-row grid gap-2 py-4 text-left" type="button" onclick={() => selectAnnouncement(itemIndex)}>
                <div class="flex min-w-0 flex-wrap items-center gap-2">
                  <span class="tape-label rotate-[-1deg]" style={`background:${priorityColor(item.priority)}`}>{item.priority}</span>
                  <strong class="min-w-0 truncate text-lg font-black">{item.title}</strong>
                </div>
                <MarkdownContent content={item.content} clamp />
                <span class="text-xs font-bold text-[hsl(var(--ink-muted))]">{formatDate(item.updated_at || item.created_at, language)}</span>
              </button>
            {/each}
          </div>
        {:else}
          <article class="grid gap-4">
            <div class="flex flex-wrap items-center gap-2">
              <span class="tape-label rotate-[-1deg]" style={`background:${priorityColor(current.priority)}`}>{current.priority}</span>
              <span class="text-xs font-black text-[hsl(var(--ink-muted))]">{index + 1} / {announcements.length}</span>
            </div>
            <MarkdownContent content={current.content} />
            <div class="grid gap-1 border-t-2 border-dashed border-[hsl(var(--ink)/0.32)] pt-3 text-xs font-bold text-[hsl(var(--ink-muted))]">
              <span>{t(language, 'announcement.updatedAt')}: {formatDate(current.updated_at || current.created_at, language)}</span>
              {#if current.starts_at}<span>{t(language, 'announcement.startsAt')}: {formatDate(current.starts_at, language)}</span>{/if}
              {#if current.ends_at}<span>{t(language, 'announcement.endsAt')}: {formatDate(current.ends_at, language)}</span>{/if}
            </div>
          </article>
        {/if}
      </div>

      <footer class="flex flex-wrap items-center justify-between gap-3 border-t-[3px] ink-line p-4">
        <div class="flex flex-wrap gap-2">
          {#if mode === 'detail' && announcements.length > 1}
            <button class="studio-button p-2" type="button" onclick={() => (index = (index - 1 + announcements.length) % announcements.length)} aria-label={t(language, 'common.previous')}><ChevronLeft class="size-4" /></button>
            <button class="studio-button p-2" type="button" onclick={() => (index = (index + 1) % announcements.length)} aria-label={t(language, 'common.next')}><ChevronRight class="size-4" /></button>
          {/if}
        </div>
        <div class="flex flex-wrap justify-end gap-2">
          {#if mode === 'history'}
            <button class="studio-button" type="button" onclick={() => (mode = 'detail')}><Megaphone class="size-4" />{t(language, 'announcement.backToCurrent')}</button>
          {:else}
            <button class="studio-button" type="button" onclick={() => (mode = 'history')}><History class="size-4" />{t(language, 'announcement.viewAll')}</button>
          {/if}
          <button class="studio-button" data-tone="primary" type="button" onclick={onAcknowledge}>{t(language, 'announcement.gotIt')}</button>
        </div>
      </footer>
    </div>
  </div>
{/if}

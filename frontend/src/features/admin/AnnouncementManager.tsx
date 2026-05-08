"use client";

import { useEffect, useMemo, useState } from "react";
import { Archive, Check, Eye, Loader2, Plus, Trash2 } from "lucide-react";
import toast from "react-hot-toast";

import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Card, CardContent } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { Label } from "@/components/ui/Label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/Select";
import { Textarea } from "@/components/ui/Textarea";
import {
  adminArchiveAnnouncement,
  adminCreateAnnouncement,
  adminDeleteAnnouncement,
  adminGetAnnouncements,
  adminUpdateAnnouncement,
} from "@/lib/api";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { t } from "@/lib/i18n";
import { priorityClassName, renderSafeMarkdown } from "@/lib/markdown";
import { cn } from "@/lib/utils";
import type { Language } from "@/types";
import type { Announcement, AnnouncementInput, AnnouncementPriority, AnnouncementStatus } from "@/types";

const PAGE_SIZE = 6;

const EMPTY_FORM: AnnouncementInput = {
  title: "",
  content: "",
  status: "draft",
  priority: "normal",
  starts_at: null,
  ends_at: null,
  sort_order: 0,
};

type Props = {
  token: string | null;
};

export function AnnouncementManager({ token }: Props) {
  const language = useUiPreferencesStore((state) => state.language);
  const [items, setItems] = useState<Announcement[]>([]);
  const [selectedID, setSelectedID] = useState<number | null>(null);
  const [form, setForm] = useState<AnnouncementInput>(EMPTY_FORM);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [preview, setPreview] = useState(false);
  const [page, setPage] = useState(1);

  const selected = useMemo(
    () => items.find((item) => item.id === selectedID) ?? null,
    [items, selectedID]
  );
  const pageCount = Math.max(1, Math.ceil(items.length / PAGE_SIZE));
  const pagedItems = items.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  const load = async () => {
    if (!token) return;
    setLoading(true);
    try {
      const announcements = await adminGetAnnouncements(token);
      setItems(announcements);
      const nextPageCount = Math.max(1, Math.ceil(announcements.length / PAGE_SIZE));
      setPage((current) => Math.min(current, nextPageCount));
      if (selectedID && !announcements.some((item) => item.id === selectedID)) {
        setSelectedID(null);
        setForm(EMPTY_FORM);
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(language, "admin.announcementsLoadError"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const selectItem = (item: Announcement) => {
    setSelectedID(item.id);
    setForm(toForm(item));
    setPreview(false);
  };

  const createNew = () => {
    setSelectedID(null);
    setForm(EMPTY_FORM);
    setPreview(false);
  };

  const goToItemPage = (id: number, source: Announcement[]) => {
    const index = source.findIndex((item) => item.id === id);
    if (index >= 0) setPage(Math.floor(index / PAGE_SIZE) + 1);
  };

  const save = async () => {
    if (!token) return;
    setSaving(true);
    try {
      if (selectedID) {
        const updated = await adminUpdateAnnouncement(token, selectedID, form);
        setItems((prev) => {
          const next = prev.map((item) => (item.id === updated.id ? updated : item));
          goToItemPage(updated.id, next);
          return next;
        });
        toast.success(t(language, "admin.announcementsSaved"));
      } else {
        const created = await adminCreateAnnouncement(token, form);
        setItems((prev) => [created, ...prev]);
        setPage(1);
        setSelectedID(created.id);
        setForm(toForm(created));
        toast.success(t(language, "admin.announcementsCreated"));
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(language, "admin.announcementsSaveError"));
    } finally {
      setSaving(false);
    }
  };

  const archive = async () => {
    if (!token || !selectedID) return;
    setSaving(true);
    try {
      const archived = await adminArchiveAnnouncement(token, selectedID);
      setItems((prev) => {
        const next = prev.map((item) => (item.id === archived.id ? archived : item));
        goToItemPage(archived.id, next);
        return next;
      });
      setForm(toForm(archived));
      toast.success(t(language, "admin.announcementsArchived"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(language, "admin.announcementsArchiveError"));
    } finally {
      setSaving(false);
    }
  };

  const remove = async () => {
    if (!token || !selectedID || !selected) return;
    if (!window.confirm(t(language, "admin.announcementsDeleteConfirm", { title: selected.title }))) return;
    setSaving(true);
    try {
      await adminDeleteAnnouncement(token, selectedID);
      setItems((prev) => {
        const next = prev.filter((item) => item.id !== selectedID);
        const nextPageCount = Math.max(1, Math.ceil(next.length / PAGE_SIZE));
        setPage((current) => Math.min(current, nextPageCount));
        return next;
      });
      createNew();
      toast.success(t(language, "admin.announcementsDeleted"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(language, "admin.announcementsDeleteError"));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-[280px_1fr]">
      <Card>
        <CardContent className="space-y-3 pt-6">
          <div className="flex items-center justify-between gap-2">
            <h2 className="font-semibold">{t(language, "admin.announcementsListTitle")}</h2>
            <Button size="sm" variant="outline" onClick={createNew} className="gap-1 cursor-pointer">
              <Plus className="h-3.5 w-3.5" />
              {t(language, "admin.announcementsNew")}
            </Button>
          </div>
          {loading ? (
            <div className="flex justify-center py-8 text-muted-foreground">
              <Loader2 className="h-5 w-5 animate-spin" />
            </div>
          ) : (
            <div className="space-y-2">
              {pagedItems.map((item) => (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => selectItem(item)}
                  className={cn(
                    "w-full rounded-lg border px-3 py-2 text-left transition-colors cursor-pointer",
                    selectedID === item.id ? "border-primary bg-primary/5" : "border-border hover:bg-muted/50"
                  )}
                >
                  <div className="mb-1 flex items-center gap-2">
                    <StatusBadge language={language} status={item.status ?? "draft"} />
                    <Badge variant="outline" className={priorityClassName(item.priority)}>
                      {priorityLabel(language, item.priority)}
                    </Badge>
                  </div>
                  <div className="line-clamp-2 text-sm font-medium">{item.title}</div>
                  <div className="mt-1 text-xs text-muted-foreground">{formatDateTime(item.updated_at)}</div>
                </button>
              ))}
              {items.length === 0 && <div className="py-8 text-center text-sm text-muted-foreground">{t(language, "admin.announcementsEmpty")}</div>}
              {items.length > 0 && (
                <div className="flex items-center justify-between gap-2 pt-2 text-xs text-muted-foreground">
                  <span>
                    {t(language, "admin.announcementsPageInfo", { page, pageCount, count: items.length })}
                  </span>
                  <div className="flex items-center gap-1">
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      onClick={() => setPage((current) => Math.max(1, current - 1))}
                      disabled={page <= 1}
                      className="h-7 px-2 cursor-pointer"
                    >
                      {t(language, "admin.announcementsPrevious")}
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      onClick={() => setPage((current) => Math.min(pageCount, current + 1))}
                      disabled={page >= pageCount}
                      className="h-7 px-2 cursor-pointer"
                    >
                      {t(language, "admin.announcementsNext")}
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardContent className="space-y-4 pt-6">
          <div className="flex items-center justify-between gap-2">
            <h2 className="font-semibold">{selectedID ? t(language, "admin.announcementsEditTitle") : t(language, "admin.announcementsCreateTitle")}</h2>
            <Button size="sm" variant="outline" onClick={() => setPreview((value) => !value)} className="gap-1 cursor-pointer">
              <Eye className="h-3.5 w-3.5" />
              {preview ? t(language, "common.edit") : t(language, "admin.announcementsPreview")}
            </Button>
          </div>

          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="announcement-title">{t(language, "admin.announcementsTitle")}</Label>
              <Input
                id="announcement-title"
                value={form.title}
                onChange={(event) => setForm((prev) => ({ ...prev, title: event.target.value }))}
                className="h-8"
              />
            </div>
            <div className="space-y-2">
              <Label>{t(language, "admin.announcementsStatus")}</Label>
              <Select value={form.status} onValueChange={(value) => setForm((prev) => ({ ...prev, status: value as AnnouncementStatus }))}>
                <SelectTrigger className="h-8 cursor-pointer">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="draft">{statusLabel(language, "draft")}</SelectItem>
                  <SelectItem value="published">{statusLabel(language, "published")}</SelectItem>
                  <SelectItem value="archived">{statusLabel(language, "archived")}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>{t(language, "admin.announcementsPriority")}</Label>
              <Select value={form.priority} onValueChange={(value) => setForm((prev) => ({ ...prev, priority: value as AnnouncementPriority }))}>
                <SelectTrigger className="h-8 cursor-pointer">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="normal">{priorityLabel(language, "normal")}</SelectItem>
                  <SelectItem value="important">{priorityLabel(language, "important")}</SelectItem>
                  <SelectItem value="urgent">{priorityLabel(language, "urgent")}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="announcement-start">{t(language, "admin.announcementsStartTime")}</Label>
              <Input
                id="announcement-start"
                type="datetime-local"
                value={toDatetimeLocal(form.starts_at)}
                onChange={(event) => setForm((prev) => ({ ...prev, starts_at: fromDatetimeLocal(event.target.value) }))}
                className="h-8"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="announcement-end">{t(language, "admin.announcementsEndTime")}</Label>
              <Input
                id="announcement-end"
                type="datetime-local"
                value={toDatetimeLocal(form.ends_at)}
                onChange={(event) => setForm((prev) => ({ ...prev, ends_at: fromDatetimeLocal(event.target.value) }))}
                className="h-8"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="announcement-sort">{t(language, "admin.announcementsSortOrder")}</Label>
              <Input
                id="announcement-sort"
                type="number"
                value={form.sort_order}
                onChange={(event) => setForm((prev) => ({ ...prev, sort_order: Number(event.target.value) }))}
                className="h-8"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="announcement-content">{t(language, "admin.announcementsContent")}</Label>
            {preview ? (
              <div
                className="announcement-markdown min-h-52 rounded-md border bg-muted/20 p-4 text-sm leading-7"
                dangerouslySetInnerHTML={{ __html: renderSafeMarkdown(form.content) }}
              />
            ) : (
              <Textarea
                id="announcement-content"
                value={form.content}
                onChange={(event) => setForm((prev) => ({ ...prev, content: event.target.value }))}
                className="min-h-52 font-mono text-sm"
                placeholder={t(language, "admin.announcementsContentPlaceholder")}
              />
            )}
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <Button size="sm" onClick={save} disabled={saving || !form.title || !form.content} className="gap-1 cursor-pointer">
              {saving ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Check className="h-3.5 w-3.5" />}
              {t(language, "common.save")}
            </Button>
            {selected && (
              <Button size="sm" variant="destructive" onClick={remove} disabled={saving} className="gap-1 cursor-pointer">
                <Trash2 className="h-3.5 w-3.5" />
                {t(language, "admin.announcementsDelete")}
              </Button>
            )}
            {selected && selected.status !== "draft" && selected.status !== "archived" && (
              <Button size="sm" variant="outline" onClick={archive} disabled={saving} className="gap-1 cursor-pointer">
                <Archive className="h-3.5 w-3.5" />
                {t(language, "admin.announcementsArchive")}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function StatusBadge({ language, status }: { language: Language; status: AnnouncementStatus }) {
  const label = statusLabel(language, status);
  return (
    <Badge variant="outline" className={statusClassName(status)}>
      {label}
    </Badge>
  );
}

function statusLabel(language: Language, status: AnnouncementStatus): string {
  switch (status) {
    case "published":
      return t(language, "admin.announcementsStatusPublished");
    case "archived":
      return t(language, "admin.announcementsStatusArchived");
    default:
      return t(language, "admin.announcementsStatusDraft");
  }
}

function priorityLabel(language: Language, priority: AnnouncementPriority): string {
  switch (priority) {
    case "urgent":
      return t(language, "admin.announcementsPriorityUrgent");
    case "important":
      return t(language, "admin.announcementsPriorityImportant");
    default:
      return t(language, "admin.announcementsPriorityNormal");
  }
}

function statusClassName(status: AnnouncementStatus): string {
  switch (status) {
    case "published":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300";
    case "archived":
      return "border-slate-500/30 bg-slate-500/10 text-slate-600 dark:text-slate-300";
    default:
      return "border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-300";
  }
}

function toForm(item: Announcement): AnnouncementInput {
  return {
    title: item.title,
    content: item.content,
    status: item.status ?? "draft",
    priority: item.priority,
    starts_at: item.starts_at,
    ends_at: item.ends_at,
    sort_order: item.sort_order ?? 0,
  };
}

function toDatetimeLocal(value: string | null): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const offset = date.getTimezoneOffset() * 60000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

function fromDatetimeLocal(value: string): string | null {
  if (!value) return null;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;
  return date.toISOString();
}

function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return date.toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

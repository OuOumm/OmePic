"use client";

import { useEffect, useMemo, useState } from "react";
import { Bell, Clock3 } from "lucide-react";

import { Button } from "@/components/ui/Button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/Dialog";
import { Badge } from "@/components/ui/Badge";
import { getAnnouncements } from "@/lib/api";
import { priorityClassName, priorityLabel, renderSafeMarkdown } from "@/lib/markdown";
import { cn } from "@/lib/utils";
import type { Announcement } from "@/types";

const READ_KEY = "omepic_read_announcement_ids";

export function AnnouncementModal() {
  const [items, setItems] = useState<Announcement[]>([]);
  const [selectedID, setSelectedID] = useState<number | null>(null);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    let alive = true;
    getAnnouncements()
      .then((announcements) => {
        if (!alive) return;
        setItems(announcements);
        setSelectedID(announcements[0]?.id ?? null);
        const unread = announcements.some((item) => !getReadIDs().has(item.id));
        if (unread) setOpen(true);
      })
      .catch(() => {});
    return () => {
      alive = false;
    };
  }, []);

  const selected = useMemo(
    () => items.find((item) => item.id === selectedID) ?? items[0] ?? null,
    [items, selectedID]
  );

  const rendered = useMemo(
    () => ({ __html: selected ? renderSafeMarkdown(selected.content) : "" }),
    [selected]
  );

  if (items.length === 0) return null;

  const markRead = () => {
    const next = getReadIDs();
    for (const item of items) next.add(item.id);
    window.localStorage.setItem(READ_KEY, JSON.stringify([...next]));
    setOpen(false);
  };

  return (
    <>
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="fixed bottom-5 right-5 z-40 gap-2 rounded-full border-border/80 bg-background/90 shadow-lg backdrop-blur cursor-pointer"
        onClick={() => setOpen(true)}
      >
        <Bell className="h-4 w-4" />
        公告
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-2xl overflow-hidden p-0">
          <div className="border-b bg-muted/30 px-6 py-5">
            <DialogHeader>
              <div className="flex items-center justify-between gap-4">
                <DialogTitle className="flex items-center gap-2">
                  <Bell className="h-5 w-5" />
                  公告
                </DialogTitle>
                <Badge variant="outline">最新 {items.length} 条</Badge>
              </div>
              <DialogDescription>查看最新通知与历史公告</DialogDescription>
            </DialogHeader>
          </div>

          <div className="grid max-h-[70vh] grid-cols-1 md:grid-cols-[1fr_220px]">
            <section className="min-h-0 overflow-y-auto px-6 py-5">
              {selected && (
                <article className="space-y-4">
                  <div className="space-y-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <Badge variant="outline" className={priorityClassName(selected.priority)}>
                        {priorityLabel(selected.priority)}
                      </Badge>
                      <span className="flex items-center gap-1 text-xs text-muted-foreground">
                        <Clock3 className="h-3.5 w-3.5" />
                        {formatDate(selected.created_at)}
                      </span>
                    </div>
                    <h2 className="text-xl font-semibold tracking-tight">{selected.title}</h2>
                  </div>
                  <div
                    className="announcement-markdown text-sm leading-7 text-foreground/90"
                    dangerouslySetInnerHTML={rendered}
                  />
                </article>
              )}
            </section>

            <aside className="border-t bg-muted/20 p-3 md:border-l md:border-t-0">
              <div className="mb-2 px-2 text-xs font-medium text-muted-foreground">历史通知</div>
              <div className="space-y-1">
                {items.map((item, index) => (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => setSelectedID(item.id)}
                    className={cn(
                      "w-full rounded-md px-3 py-2 text-left transition-colors cursor-pointer",
                      selected?.id === item.id ? "bg-background shadow-sm" : "hover:bg-background/70"
                    )}
                  >
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <span>{index === 0 ? "最新" : priorityLabel(item.priority)}</span>
                      <span>{formatDate(item.created_at)}</span>
                    </div>
                    <div className="mt-1 line-clamp-2 text-sm font-medium">{item.title}</div>
                  </button>
                ))}
              </div>
            </aside>
          </div>

          <div className="flex justify-end gap-2 border-t bg-background px-6 py-4">
            <Button variant="outline" onClick={() => setOpen(false)} className="cursor-pointer">
              稍后查看
            </Button>
            <Button onClick={markRead} className="cursor-pointer">
              我知道了
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

function getReadIDs(): Set<number> {
  if (typeof window === "undefined") return new Set();
  try {
    const parsed: unknown = JSON.parse(window.localStorage.getItem(READ_KEY) ?? "[]");
    if (!Array.isArray(parsed)) return new Set();
    return new Set(parsed.filter((value): value is number => typeof value === "number"));
  } catch {
    return new Set();
  }
}

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return date.toLocaleDateString("zh-CN", { month: "2-digit", day: "2-digit" });
}

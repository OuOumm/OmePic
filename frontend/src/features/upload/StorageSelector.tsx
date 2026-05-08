"use client";

import { Button } from "@/components/ui/Button";
import { useUploadStore } from "@/stores/upload-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { t } from "@/lib/i18n";
import { HardDrive, RefreshCw } from "lucide-react";
import { cn } from "@/lib/utils";

type Props = {
  refreshing?: boolean;
  onRefresh?: () => void;
};

export function StorageSelector({ refreshing = false, onRefresh }: Props) {
  const language = useUiPreferencesStore((state) => state.language);
  const selectedStorageKey = useUploadStore((state) => state.selectedStorageKey);
  const runtimeSettings = useUploadStore((state) => state.runtimeSettings);
  const setSelectedStorageKey = useUploadStore((state) => state.setSelectedStorageKey);

  const options = runtimeSettings?.storage.options ?? [];
  const allowStorageSelection = runtimeSettings?.features.allow_storage_selection ?? true;

  if (!allowStorageSelection || options.length === 0) return null;

  const nonDefaultOptions = options.filter((o) => !o.is_default);
  const lang = language;

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <span className="text-xs text-muted-foreground flex items-center gap-1 shrink-0">
        <HardDrive className="h-3.5 w-3.5" />
        {t(lang, "upload.storageLabel")}
      </span>
      <div className="flex items-center gap-1 flex-wrap">
        <button
          type="button"
          onClick={() => setSelectedStorageKey("")}
          className={cn(
            "inline-flex items-center rounded-full px-3 py-1 text-xs font-medium",
            "transition-colors duration-150 cursor-pointer",
            "border",
            !selectedStorageKey
              ? "bg-primary text-primary-foreground border-primary"
              : "bg-background text-muted-foreground border-border hover:bg-accent hover:text-accent-foreground"
          )}
        >
          {t(lang, "common.default")}
        </button>
        {nonDefaultOptions.map((opt) => (
          <button
            key={opt.storage_key}
            type="button"
            onClick={() => setSelectedStorageKey(opt.storage_key)}
            className={cn(
              "inline-flex items-center rounded-full px-3 py-1 text-xs font-medium",
              "transition-colors duration-150 cursor-pointer",
              "border",
              selectedStorageKey === opt.storage_key
                ? "bg-primary text-primary-foreground border-primary"
                : "bg-background text-muted-foreground border-border hover:bg-accent hover:text-accent-foreground"
            )}
          >
            {opt.name}
          </button>
        ))}
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 cursor-pointer shrink-0"
          onClick={onRefresh}
          disabled={refreshing || !onRefresh}
          title={t(lang, "common.refresh")}
          aria-label={t(lang, "common.refresh")}
        >
          <RefreshCw className={cn("h-3.5 w-3.5", refreshing && "animate-spin")} />
        </Button>
      </div>
    </div>
  );
}

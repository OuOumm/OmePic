"use client";

import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/Button";
import { useUploadStore } from "@/stores/upload-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { getStorageOptions } from "@/lib/api";
import { t } from "@/lib/i18n";
import { RefreshCw, HardDrive } from "lucide-react";
import { cn } from "@/lib/utils";
import type { StorageOption } from "@/types";

export function StorageSelector() {
  const language = useUiPreferencesStore((state) => state.language);
  const selectedStorageKey = useUploadStore((state) => state.selectedStorageKey);
  const setSelectedStorageKey = useUploadStore((state) => state.setSelectedStorageKey);

  const [options, setOptions] = useState<StorageOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);

  const fetchOptions = useCallback(async () => {
    setLoading(true);
    setError(false);
    try {
      const opts = await getStorageOptions();
      setOptions(opts);
      if (selectedStorageKey && !opts.find((o) => o.storage_key === selectedStorageKey)) {
        setSelectedStorageKey("");
      }
    } catch (e) {
      console.error("Failed to load storage options:", e);
      setError(true);
    } finally {
      setLoading(false);
    }
  }, [selectedStorageKey, setSelectedStorageKey]);

  useEffect(() => {
    fetchOptions();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // Filter out default storage from the list — it's already represented by the "默认" chip
  const nonDefaultOptions = options.filter((o) => !o.is_default);

  // If no options fetched at all, hide the component
  if (options.length === 0 && !loading && !error) return null;

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
        {error && (
          <span className="text-xs text-muted-foreground italic">
            {t(lang, "common.error")}
          </span>
        )}
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 cursor-pointer shrink-0"
          onClick={fetchOptions}
          disabled={loading}
          aria-label={t(lang, "common.refresh")}
        >
          <RefreshCw className={cn("h-3 w-3", loading && "animate-spin")} />
        </Button>
      </div>
    </div>
  );
}

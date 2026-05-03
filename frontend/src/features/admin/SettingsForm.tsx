"use client";

import { useEffect, useState } from "react";
import { Plus } from "lucide-react";
import toast from "react-hot-toast";

import { PageSectionHeader } from "@/components/shared/PageLayout";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/Select";
import { Badge } from "@/components/ui/Badge";
import { useUiTranslations } from "@/hooks/useUiPreferences";
import {
  adminCreateStorageConfig,
  adminDeleteStorageConfig,
  adminGetConfig,
  adminSetDefaultStorageConfig,
  adminUpdateStorageConfig
} from "@/lib/api";
import { cn } from "@/lib/utils";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import type {
  AdminConfig,
  AdminStorageConfig,
  AdminStorageConfigCreateInput,
  AdminStorageConfigUpdateInput,
  StorageBackend
} from "@/types/admin";

const maskedPrefix = "****";

type StorageDraft = {
  name: string;
  storage_backend: StorageBackend;
  local_storage_path: string;
  s3_endpoint: string;
  s3_region: string;
  s3_bucket: string;
  s3_access_key: string;
  s3_secret_key: string;
  s3_use_ssl: boolean;
  s3_force_path_style: boolean;
  webdav_url: string;
  webdav_user: string;
  webdav_pass: string;
};

const emptyDraft: StorageDraft = {
  name: "",
  storage_backend: "local",
  local_storage_path: "",
  s3_endpoint: "",
  s3_region: "auto",
  s3_bucket: "",
  s3_access_key: "",
  s3_secret_key: "",
  s3_use_ssl: false,
  s3_force_path_style: true,
  webdav_url: "",
  webdav_user: "",
  webdav_pass: ""
};

export function SettingsForm() {
  const token = useAdminSessionStore((state) => state.token);
  const t = useUiTranslations();
  const [config, setConfig] = useState<AdminConfig | null>(null);
  const [activeKey, setActiveKey] = useState<string | null>(null);
  const [draft, setDraft] = useState<StorageDraft>(emptyDraft);
  const [saving, setSaving] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    void adminGetConfig(token)
      .then((result) => {
        if (cancelled) {
          return;
        }
        setConfig(result);
        setErrorMessage(null);
        if (result.storage_configs.length > 0) {
          const nextActiveKey = result.default_storage_key || result.storage_configs[0].storage_key;
          setActiveKey(nextActiveKey);
          setDraft(toDraft(result.storage_configs.find((item) => item.storage_key === nextActiveKey) ?? result.storage_configs[0]));
        }
      })
      .catch((error: Error) => {
        if (!cancelled) {
          setErrorMessage(error.message);
          toast.error(error.message);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [token]);

  const activeItem = config?.storage_configs.find((item) => item.storage_key === activeKey) ?? null;
  const isCreating = activeKey === null;

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!config) {
      return;
    }

    setSaving(true);
    setNotice(null);
    setErrorMessage(null);
    try {
      const previousKeys = new Set(config.storage_configs.map((item) => item.storage_key));
      const next = isCreating
        ? await adminCreateStorageConfig(token, buildCreatePayload(draft))
        : await adminUpdateStorageConfig(token, activeKey, buildUpdatePayload(draft));
      setConfig(next);

      if (isCreating) {
        const created = next.storage_configs.find((item) => !previousKeys.has(item.storage_key));
        const nextActive = created?.storage_key ?? next.default_storage_key ?? next.storage_configs[0]?.storage_key ?? null;
        setActiveKey(nextActive);
        if (nextActive) {
          const selected = next.storage_configs.find((item) => item.storage_key === nextActive);
          if (selected) {
            setDraft(toDraft(selected));
          }
        }
        setNotice(t.admin.storageCreateSuccessToast);
        toast.success(t.admin.storageCreateSuccessToast);
      } else {
        const updated = next.storage_configs.find((item) => item.storage_key === activeKey);
        if (updated) {
          setDraft(toDraft(updated));
        }
        setNotice(t.admin.storageUpdateSuccessToast);
        toast.success(t.admin.storageUpdateSuccessToast);
      }
    } catch (error) {
      const nextError = error instanceof Error ? error.message : t.admin.configUpdateFailed;
      setErrorMessage(nextError);
      toast.error(nextError);
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!config || !activeKey) {
      return;
    }

    setSaving(true);
    setNotice(null);
    setErrorMessage(null);
    try {
      const next = await adminDeleteStorageConfig(token, activeKey);
      setConfig(next);
      const nextActive = next.default_storage_key || next.storage_configs[0]?.storage_key || null;
      setActiveKey(nextActive);
      if (nextActive) {
        const selected = next.storage_configs.find((item) => item.storage_key === nextActive);
        if (selected) {
          setDraft(toDraft(selected));
        }
      } else {
        setDraft(emptyDraft);
      }
      setNotice(t.admin.storageDeleteSuccessToast);
      toast.success(t.admin.storageDeleteSuccessToast);
    } catch (error) {
      const nextError = error instanceof Error ? error.message : t.admin.storageDeleteFailed;
      setErrorMessage(nextError);
      toast.error(nextError);
    } finally {
      setSaving(false);
    }
  }

  async function handleSetDefault(storageKey: string) {
    setSaving(true);
    setNotice(null);
    setErrorMessage(null);
    try {
      const next = await adminSetDefaultStorageConfig(token, storageKey);
      setConfig(next);
      const selected = next.storage_configs.find((item) => item.storage_key === storageKey);
      if (selected) {
        setActiveKey(storageKey);
        setDraft(toDraft(selected));
      }
      setNotice(t.admin.defaultStorageUpdatedToast);
      toast.success(t.admin.defaultStorageUpdatedToast);
    } catch (error) {
      const nextError = error instanceof Error ? error.message : t.admin.defaultStorageUpdateFailed;
      setErrorMessage(nextError);
      toast.error(nextError);
    } finally {
      setSaving(false);
    }
  }

  if (!config) {
    return (
      <Card
        className={cn(
          "flex items-center gap-4 p-5 text-sm",
          errorMessage ? "border-rose-400/30 bg-rose-500/10 text-danger" : "text-muted-foreground"
        )}
        role={errorMessage ? "alert" : "status"}
        variant="strong"
      >
        {!errorMessage ? <span className="skeleton-glass h-10 w-10 rounded-md" /> : null}
        {errorMessage || t.admin.settingsLoading}
      </Card>
    );
  }

  return (
    <div className="grid gap-5 animate-fade-in xl:grid-cols-[340px_1fr]">
      <Card className="h-fit overflow-hidden p-5 xl:sticky xl:top-24" variant="strong">
        <PageSectionHeader
          description={t.admin.settingsDescription}
          title={t.admin.settingsTitle}
        />

        <Button
          className="mt-5 w-full"
          onClick={() => {
            setActiveKey(null);
            setDraft({
              ...emptyDraft,
              local_storage_path: config.storage_configs[0]?.local_storage_path ?? ""
            });
          }}
        >
          <Plus aria-hidden="true" className="h-4 w-4" />
          {t.admin.createStorageInstance}
        </Button>

        <div className="mt-5 space-y-3">
          {config.storage_configs.map((item) => {
            const isActive = item.storage_key === activeKey;
            return (
              <button
                className={cn(
                  "w-full rounded-md border p-4 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                  isActive
                    ? "border-primary bg-muted"
                    : "border-border bg-card hover:bg-muted/50"
                )}
                key={item.storage_key}
                onClick={() => {
                  setActiveKey(item.storage_key);
                  setDraft(toDraft(item));
                }}
                type="button"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="font-medium text-foreground">{item.name}</p>
                    <p className="mt-1 font-mono text-xs text-muted-foreground">{item.storage_key}</p>
                  </div>
                  {item.is_default ? <Badge>{t.admin.defaultBadge}</Badge> : null}
                </div>
                <p className="mt-3 text-sm text-muted-foreground">{backendLabel(item.storage_backend, t)}</p>
              </button>
            );
          })}
        </div>
      </Card>

      <Card className="overflow-hidden p-5 sm:p-6" variant="strong">
        <form className="space-y-6" onSubmit={handleSubmit}>
          <PageSectionHeader
            description={isCreating ? t.admin.createStorageDescription : t.admin.editStorageDescription}
            title={isCreating ? t.admin.createStorageTitle : t.admin.editStorageTitle}
          />

          {notice ? (
            <p className="rounded-md border border-border bg-muted/50 p-3 text-sm text-foreground" role="status">
              {notice}
            </p>
          ) : null}

          {errorMessage ? (
            <p className="rounded-md border border-rose-400/30 bg-rose-500/10 p-3 text-sm text-danger" role="alert">
              {errorMessage}
            </p>
          ) : null}

          <div className="grid gap-4 md:grid-cols-2">
            <Field label={t.admin.fields.storageName}>
              <Input
                onChange={(event) => setDraft({ ...draft, name: event.target.value })}
                value={draft.name}
              />
            </Field>

            <Field label={t.admin.fields.storageBackend}>
              <div className="space-y-2">
                <Select
                  disabled={!isCreating}
                  onValueChange={(value) =>
                    setDraft({ ...draft, storage_backend: value as StorageBackend })
                  }
                  value={draft.storage_backend}
                >
                  <SelectTrigger title={!isCreating ? t.admin.storageBackendLockedHint : undefined}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="local">{t.admin.backends.local}</SelectItem>
                    <SelectItem value="s3">{t.admin.backends.s3}</SelectItem>
                    <SelectItem value="webdav">{t.admin.backends.webdav}</SelectItem>
                  </SelectContent>
                </Select>
                {!isCreating ? (
                  <p className="text-xs text-muted-foreground">{t.admin.storageBackendLockedHint}</p>
                ) : null}
              </div>
            </Field>

            <Field label={t.admin.fields.localStoragePath}>
              <Input
                onChange={(event) => setDraft({ ...draft, local_storage_path: event.target.value })}
                value={draft.local_storage_path}
              />
            </Field>

            {activeItem ? (
              <Field label={t.admin.fields.storageKey}>
                <Input disabled value={activeItem.storage_key} />
              </Field>
            ) : null}

            {draft.storage_backend === "s3" ? (
              <>
                <Field label={t.admin.fields.s3Endpoint}>
                  <Input
                    onChange={(event) => setDraft({ ...draft, s3_endpoint: event.target.value })}
                    value={draft.s3_endpoint}
                  />
                </Field>
                <Field label={t.admin.fields.s3Region}>
                  <Input
                    onChange={(event) => setDraft({ ...draft, s3_region: event.target.value })}
                    value={draft.s3_region}
                  />
                </Field>
                <Field label={t.admin.fields.s3Bucket}>
                  <Input
                    onChange={(event) => setDraft({ ...draft, s3_bucket: event.target.value })}
                    value={draft.s3_bucket}
                  />
                </Field>
                <Field label={t.admin.fields.s3AccessKey}>
                  <Input
                    onChange={(event) => setDraft({ ...draft, s3_access_key: event.target.value })}
                    value={draft.s3_access_key}
                  />
                </Field>
                <Field label={t.admin.fields.s3SecretKey}>
                  <Input
                    onChange={(event) => setDraft({ ...draft, s3_secret_key: event.target.value })}
                    value={draft.s3_secret_key}
                  />
                </Field>
              </>
            ) : null}

            {draft.storage_backend === "webdav" ? (
              <>
                <Field label={t.admin.fields.webdavUrl}>
                  <Input
                    onChange={(event) => setDraft({ ...draft, webdav_url: event.target.value })}
                    value={draft.webdav_url}
                  />
                </Field>
                <Field label={t.admin.fields.webdavUser}>
                  <Input
                    onChange={(event) => setDraft({ ...draft, webdav_user: event.target.value })}
                    value={draft.webdav_user}
                  />
                </Field>
                <Field label={t.admin.fields.webdavPassword}>
                  <Input
                    onChange={(event) => setDraft({ ...draft, webdav_pass: event.target.value })}
                    value={draft.webdav_pass}
                  />
                </Field>
              </>
            ) : null}
          </div>

          {draft.storage_backend === "s3" ? (
            <div className="grid gap-3">
              <label className="flex items-center gap-3 text-sm">
                <input
                  checked={draft.s3_use_ssl}
                  className="h-4 w-4 rounded border-input text-primary focus:ring-ring"
                  onChange={(event) => setDraft({ ...draft, s3_use_ssl: event.target.checked })}
                  type="checkbox"
                />
                {t.admin.toggles.s3UseSsl}
              </label>
              <label className="flex items-center gap-3 text-sm">
                <input
                  checked={draft.s3_force_path_style}
                  className="h-4 w-4 rounded border-input text-primary focus:ring-ring"
                  onChange={(event) =>
                    setDraft({ ...draft, s3_force_path_style: event.target.checked })
                  }
                  type="checkbox"
                />
                {t.admin.toggles.s3ForcePathStyle}
              </label>
            </div>
          ) : null}

          <div className="flex flex-wrap gap-3">
            <Button disabled={saving} type="submit">
              {saving ? t.admin.saving : isCreating ? t.admin.createStorageSubmit : t.admin.saveSettings}
            </Button>

            {activeItem && !activeItem.is_default ? (
              <Button
                disabled={saving}
                onClick={() => void handleSetDefault(activeItem.storage_key)}
                type="button"
                variant="secondary"
              >
                {t.admin.makeDefault}
              </Button>
            ) : null}

            {activeItem ? (
              <Button
                disabled={saving || activeItem.is_default}
                onClick={() => void handleDelete()}
                title={activeItem.is_default ? t.admin.defaultDeleteBlockedHint : undefined}
                type="button"
                variant="danger"
              >
                {t.admin.deleteStorageInstance}
              </Button>
            ) : null}
          </div>
        </form>
      </Card>
    </div>
  );
}

function Field({
  children,
  label
}: {
  children: React.ReactNode;
  label: string;
}) {
  return (
    <label className="space-y-2">
      <span className="text-sm font-medium text-foreground">{label}</span>
      {children}
    </label>
  );
}

function toDraft(config: AdminStorageConfig): StorageDraft {
  return {
    name: config.name,
    storage_backend: config.storage_backend as StorageBackend,
    local_storage_path: config.local_storage_path,
    s3_endpoint: config.s3_endpoint,
    s3_region: config.s3_region,
    s3_bucket: config.s3_bucket,
    s3_access_key: config.s3_access_key,
    s3_secret_key: config.s3_secret_key,
    s3_use_ssl: config.s3_use_ssl,
    s3_force_path_style: config.s3_force_path_style,
    webdav_url: config.webdav_url,
    webdav_user: config.webdav_user,
    webdav_pass: config.webdav_pass
  };
}

function buildCreatePayload(draft: StorageDraft): AdminStorageConfigCreateInput {
  return { ...draft };
}

function buildUpdatePayload(draft: StorageDraft): AdminStorageConfigUpdateInput {
  return {
    ...draft,
    s3_access_key: draft.s3_access_key.startsWith(maskedPrefix) ? undefined : draft.s3_access_key,
    s3_secret_key: draft.s3_secret_key.startsWith(maskedPrefix) ? undefined : draft.s3_secret_key,
    webdav_pass: draft.webdav_pass.startsWith(maskedPrefix) ? undefined : draft.webdav_pass
  };
}

function backendLabel(backend: string, t: ReturnType<typeof useUiTranslations>) {
  switch (backend) {
    case "s3":
      return t.admin.backends.s3;
    case "webdav":
      return t.admin.backends.webdav;
    default:
      return t.admin.backends.local;
  }
}

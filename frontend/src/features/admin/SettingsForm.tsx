"use client";

import { useState, useEffect, useCallback } from "react";
import { Card, CardContent } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Label } from "@/components/ui/Label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/Select";
import { Separator } from "@/components/ui/Separator";
import { Badge } from "@/components/ui/Badge";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import {
  adminGetConfig,
  adminCreateStorageInstance,
  adminUpdateStorageInstance,
  adminDeleteStorageInstance,
  adminSetDefaultStorage,
} from "@/lib/api";
import { t } from "@/lib/i18n";
import {
  Plus,
  Trash2,
  Loader2,
  AlertCircle,
  Star,
  Database,
  Globe,
  FolderOpen,
  Settings,
  Check,
} from "lucide-react";
import toast from "react-hot-toast";
import type { StorageInstance, AdminConfig } from "@/types";

function maskSecret(val: string | undefined): string {
  if (!val) return "";
  if (val.startsWith("****")) return "****__MASKED__";
  return val;
}

function buildSubmitPayload(
  instance: Partial<StorageInstance>,
  isUpdate: boolean
): Partial<StorageInstance> {
  const payload = { ...instance };
  // For update, skip masked secret fields
  if (isUpdate) {
    if (payload.s3_secret_key?.startsWith("****")) {
      delete payload.s3_secret_key;
    }
    if (payload.webdav_pass?.startsWith("****")) {
      delete payload.webdav_pass;
    }
  }
  return payload;
}

const BACKENDS: { key: StorageInstance["storage_backend"]; label: string; icon: typeof FolderOpen }[] = [
  { key: "local", label: "Local", icon: FolderOpen },
  { key: "s3", label: "S3", icon: Database },
  { key: "webdav", label: "WebDAV", icon: Globe },
];

export function SettingsForm() {
  const token = useAdminSessionStore((state) => state.token);
  const language = useUiPreferencesStore((state) => state.language);

  const [config, setConfig] = useState<AdminConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedKey, setSelectedKey] = useState<string>("");
  const [isNew, setIsNew] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  // Form state
  const [form, setForm] = useState<Partial<StorageInstance>>({});

  const loadConfig = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError("");
    try {
      const c = await adminGetConfig(token);
      setConfig(c);
      // Auto-select default or first
      if (!selectedKey || !c.storage_configs.find((i) => i.storage_key === selectedKey)) {
        const def = c.storage_configs.find((i) => i.is_default);
        setSelectedKey(def?.storage_key ?? c.storage_configs[0]?.storage_key ?? "");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load config");
    } finally {
      setLoading(false);
    }
  }, [token, selectedKey]);

  useEffect(() => {
    loadConfig();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // Load instance into form
  useEffect(() => {
    if (!config) return;
    const inst = config.storage_configs.find((i) => i.storage_key === selectedKey);
    if (inst) {
      setIsNew(false);
      setForm({
        ...inst,
        s3_secret_key: maskSecret(inst.s3_secret_key),
        webdav_pass: maskSecret(inst.webdav_pass),
      });
    }
  }, [config, selectedKey]);

  const handleNew = () => {
    setIsNew(true);
    setSelectedKey("");
    setForm({ name: "", storage_backend: "local", local_storage_path: "" });
  };

  const handleSelect = (key: string) => {
    setIsNew(false);
    setSelectedKey(key);
  };

  const updateField = (field: string, value: string | boolean) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  const handleSave = async () => {
    if (!token || !form.name) return;
    setSaving(true);
    try {
      if (isNew) {
        if (!form.storage_backend) return;
        const createdConfig = await adminCreateStorageInstance(
          token,
          buildSubmitPayload(form, false)
        );
        toast.success(t(language, "admin.settingsCreated"));
        setConfig(createdConfig);
        const created = createdConfig.storage_configs.find((inst) => !config?.storage_configs.some((oldInst) => oldInst.storage_key === inst.storage_key));
        setSelectedKey(created?.storage_key ?? createdConfig.default_storage_key ?? createdConfig.storage_configs[0]?.storage_key ?? "");
      } else {
        const updatedConfig = await adminUpdateStorageInstance(
          token,
          selectedKey,
          buildSubmitPayload(form, true)
        );
        toast.success(t(language, "admin.settingsSaved"));
        setConfig(updatedConfig);
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(language, "common.error"));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!token || !selectedKey) return;
    const inst = config?.storage_configs.find((i) => i.storage_key === selectedKey);
    if (!inst || inst.is_default) {
      toast.error("Cannot delete the default storage instance");
      return;
    }
    if (!window.confirm(t(language, "admin.settingsDeleteConfirm", { name: inst.name }))) return;
    setDeleting(true);
    try {
      const deletedConfig = await adminDeleteStorageInstance(token, selectedKey);
      toast.success(t(language, "admin.settingsDeleted"));
      setConfig(deletedConfig);
      const def = deletedConfig.storage_configs.find((i) => i.is_default);
      setSelectedKey(def?.storage_key ?? deletedConfig.storage_configs[0]?.storage_key ?? "");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(language, "common.error"));
    } finally {
      setDeleting(false);
    }
  };

  const handleSetDefault = async () => {
    if (!token || !selectedKey) return;
    setSaving(true);
    try {
      await adminSetDefaultStorage(token, selectedKey);
      toast.success(t(language, "common.success"));
      await loadConfig();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(language, "common.error"));
    } finally {
      setSaving(false);
    }
  };

  const lang = language;

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center gap-2 text-destructive" role="alert">
        <AlertCircle className="h-5 w-5" />
        {error}
      </div>
    );
  }

  const selectedInst = config?.storage_configs.find((i) => i.storage_key === selectedKey);
  const backendType = isNew ? form.storage_backend : selectedInst?.storage_backend;
  const isDefault = selectedInst?.is_default ?? false;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h1 className="text-xl font-bold">{t(lang, "admin.settingsTitle")}</h1>
        <Button size="sm" onClick={handleNew} className="cursor-pointer gap-1" variant="outline">
          <Plus className="h-3.5 w-3.5" />
          {t(lang, "admin.settingsNew")}
        </Button>
      </div>

      <div className="flex gap-6">
        {/* Instance list */}
        <div className="w-56 shrink-0 space-y-1">
          {config?.storage_configs.map((inst) => (
            <Button
              key={inst.storage_key}
              variant={selectedKey === inst.storage_key && !isNew ? "secondary" : "ghost"}
              size="sm"
              onClick={() => handleSelect(inst.storage_key)}
              className="w-full justify-start cursor-pointer gap-2"
            >
              {inst.is_default ? <Star className="h-3.5 w-3.5 text-amber-500" /> : <Settings className="h-3.5 w-3.5" />}
              <span className="truncate">{inst.name}</span>
              <Badge variant="outline" className="ml-auto text-[10px] text-xs">
                {inst.storage_backend}
              </Badge>
            </Button>
          ))}
          {isNew && (
            <Button variant="secondary" size="sm" className="w-full justify-start cursor-pointer gap-2">
              <Plus className="h-3.5 w-3.5" />
              <span className="truncate">{form.name || "New Instance"}</span>
            </Button>
          )}
        </div>

        {/* Form */}
        <Card className="flex-1">
          <CardContent className="pt-6 space-y-4">
            {/* Name */}
            <div className="space-y-2">
              <Label htmlFor="inst-name">{t(lang, "admin.settingsName")}</Label>
              <Input
                id="inst-name"
                value={form.name || ""}
                onChange={(e) => updateField("name", e.target.value)}
                className="h-8"
              />
            </div>

            {/* Backend (locked on edit) */}
            <div className="space-y-2">
              <Label htmlFor="inst-backend">{t(lang, "admin.settingsBackend")}</Label>
              {isNew ? (
                <Select value={String(form.storage_backend)} onValueChange={(v) => updateField("storage_backend", v as StorageInstance["storage_backend"])}>
                  <SelectTrigger id="inst-backend" className="h-8 cursor-pointer">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {BACKENDS.map((b) => (
                      <SelectItem key={b.key} value={b.key}>{b.label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : (
                <Input value={backendType} disabled className="h-8" />
              )}
            </div>

            {/* Storage Key (read-only display) */}
            {!isNew && (
              <div className="space-y-2">
                <Label>{t(lang, "admin.settingsKey")}</Label>
                <Input value={selectedKey} disabled className="h-8 font-mono text-xs" />
              </div>
            )}

            <Separator />

            {/* Local fields */}
            {backendType === "local" && (
              <div className="space-y-2">
                <Label htmlFor="local-path">{t(lang, "admin.settingsLocalPath")}</Label>
                <Input
                  id="local-path"
                  value={form.local_storage_path || ""}
                  onChange={(e) => updateField("local_storage_path", e.target.value)}
                  className="h-8"
                  placeholder="/data/images"
                />
              </div>
            )}

            {/* S3 fields */}
            {backendType === "s3" && (
              <div className="space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label htmlFor="s3-endpoint">{t(lang, "admin.settingsS3Endpoint")}</Label>
                    <Input id="s3-endpoint" value={form.s3_endpoint || ""} onChange={(e) => updateField("s3_endpoint", e.target.value)} className="h-8" />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="s3-region">{t(lang, "admin.settingsS3Region")}</Label>
                    <Input id="s3-region" value={form.s3_region || ""} onChange={(e) => updateField("s3_region", e.target.value)} className="h-8" />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="s3-bucket">{t(lang, "admin.settingsS3Bucket")}</Label>
                  <Input id="s3-bucket" value={form.s3_bucket || ""} onChange={(e) => updateField("s3_bucket", e.target.value)} className="h-8" />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label htmlFor="s3-access">{t(lang, "admin.settingsS3AccessKey")}</Label>
                    <Input id="s3-access" value={form.s3_access_key || ""} onChange={(e) => updateField("s3_access_key", e.target.value)} className="h-8" />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="s3-secret">{t(lang, "admin.settingsS3SecretKey")}</Label>
                    <Input id="s3-secret" type="password" value={form.s3_secret_key || ""} onChange={(e) => updateField("s3_secret_key", e.target.value)} className="h-8" />
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <label className="flex items-center gap-2 text-sm cursor-pointer">
                    <input
                      type="checkbox"
                      checked={!!form.s3_use_ssl}
                      onChange={(e) => updateField("s3_use_ssl", e.target.checked)}
                      className="cursor-pointer"
                    />
                    {t(lang, "admin.settingsS3SSL")}
                  </label>
                  <label className="flex items-center gap-2 text-sm cursor-pointer">
                    <input
                      type="checkbox"
                      checked={!!form.s3_force_path_style}
                      onChange={(e) => updateField("s3_force_path_style", e.target.checked)}
                      className="cursor-pointer"
                    />
                    {t(lang, "admin.settingsS3PathStyle")}
                  </label>
                </div>
              </div>
            )}

            {/* WebDAV fields */}
            {backendType === "webdav" && (
              <div className="space-y-3">
                <div className="space-y-2">
                  <Label htmlFor="webdav-url">{t(lang, "admin.settingsWebdavUrl")}</Label>
                  <Input id="webdav-url" value={form.webdav_url || ""} onChange={(e) => updateField("webdav_url", e.target.value)} className="h-8" />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label htmlFor="webdav-user">{t(lang, "admin.settingsWebdavUser")}</Label>
                    <Input id="webdav-user" value={form.webdav_user || ""} onChange={(e) => updateField("webdav_user", e.target.value)} className="h-8" />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="webdav-pass">{t(lang, "admin.settingsWebdavPassword")}</Label>
                    <Input id="webdav-pass" type="password" value={form.webdav_pass || ""} onChange={(e) => updateField("webdav_pass", e.target.value)} className="h-8" />
                  </div>
                </div>
              </div>
            )}

            <Separator />

            {/* Actions */}
            <div className="flex items-center gap-2">
              <Button size="sm" onClick={handleSave} disabled={saving || !form.name} className="cursor-pointer gap-1">
                {saving ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Check className="h-3.5 w-3.5" />}
                {t(lang, "common.save")}
              </Button>
              {!isNew && !isDefault && (
                <>
                  <Button size="sm" variant="outline" onClick={handleSetDefault} disabled={saving} className="cursor-pointer">
                    <Star className="h-3.5 w-3.5" />
                    {t(lang, "admin.settingsSetDefault")}
                  </Button>
                  <Button size="sm" variant="destructive" onClick={handleDelete} disabled={deleting} className="cursor-pointer">
                    {deleting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Trash2 className="h-3.5 w-3.5" />}
                    {t(lang, "admin.settingsDelete")}
                  </Button>
                </>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

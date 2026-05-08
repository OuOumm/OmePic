"use client";

import { useState, useEffect, useCallback } from "react";
import { Card, CardContent } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Label } from "@/components/ui/Label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/Select";
import { Separator } from "@/components/ui/Separator";
import { Badge } from "@/components/ui/Badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/Tabs";
import { AnnouncementManager } from "./AnnouncementManager";
import { DEFAULT_RUNTIME_SETTINGS, normalizeRuntimeSettings } from "./runtime-settings";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import {
  adminGetConfig,
  adminCreateStorageInstance,
  adminUpdateStorageInstance,
  adminDeleteStorageInstance,
  adminSetDefaultStorage,
  adminGetSystemSettings,
  adminUpdateSystemSettings,
} from "@/lib/api";
import { t } from "@/lib/i18n";
import { cn } from "@/lib/utils";
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
import type { StorageInstance, AdminConfig, RuntimeSettings } from "@/types";

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
  const [runtimeForm, setRuntimeForm] = useState<RuntimeSettings>(DEFAULT_RUNTIME_SETTINGS);
  const [systemSaving, setSystemSaving] = useState(false);

  // Form state
  const [form, setForm] = useState<Partial<StorageInstance>>({});

  const loadSystemSettings = useCallback(async () => {
    if (!token) return;
    const settings = await adminGetSystemSettings(token);
    const runtime = normalizeRuntimeSettings(settings.runtime);
    setRuntimeForm(runtime);
  }, [token]);

  const loadConfig = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError("");
    try {
      const [c] = await Promise.all([
        adminGetConfig(token),
        loadSystemSettings(),
      ]);
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
  }, [token, selectedKey, loadSystemSettings]);

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

  const updateRuntimeField = <K extends keyof RuntimeSettings>(field: K, value: RuntimeSettings[K]) => {
    setRuntimeForm((prev) => ({ ...prev, [field]: value }));
  };

  const handleSaveRuntimeSettings = async () => {
    if (!token) return;
    setSystemSaving(true);
    try {
      const saved = await adminUpdateSystemSettings(token, runtimeForm);
      const runtime = normalizeRuntimeSettings(saved.runtime);
      setRuntimeForm(runtime);
      toast.success(t(language, "admin.settingsSaved"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(language, "common.error"));
    } finally {
      setSystemSaving(false);
    }
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
      </div>

      <Tabs defaultValue="storage" className="space-y-4">
        <TabsList>
          <TabsTrigger value="storage">{t(lang, "admin.settingsTabStorage")}</TabsTrigger>
          <TabsTrigger value="runtime">{t(lang, "admin.settingsTabRuntime")}</TabsTrigger>
          <TabsTrigger value="announcements">{t(lang, "admin.settingsTabAnnouncements")}</TabsTrigger>
        </TabsList>
        <TabsContent value="storage" className="space-y-4">
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-[280px_1fr]">
            <Card>
              <CardContent className="space-y-3 pt-6">
                <div className="flex items-center justify-between gap-2">
                  <div>
                    <h2 className="font-semibold">{t(lang, "admin.settingsStorageListTitle")}</h2>
                    <p className="mt-1 text-xs text-muted-foreground">{t(lang, "admin.settingsStorageListDescription")}</p>
                  </div>
                  <Button size="sm" onClick={handleNew} className="cursor-pointer gap-1" variant="outline">
                    <Plus className="h-3.5 w-3.5" />
                    {t(lang, "admin.settingsNew")}
                  </Button>
                </div>

                <div className="space-y-2">
                  {config?.storage_configs.map((inst) => {
                    const backend = BACKENDS.find((item) => item.key === inst.storage_backend);
                    const Icon = backend?.icon ?? Settings;
                    const active = selectedKey === inst.storage_key && !isNew;
                    return (
                      <button
                        key={inst.storage_key}
                        type="button"
                        onClick={() => handleSelect(inst.storage_key)}
                        className={cn(
                          "w-full rounded-lg border px-3 py-2 text-left transition-colors cursor-pointer",
                          active ? "border-primary bg-primary/5" : "border-border hover:bg-muted/50"
                        )}
                      >
                        <div className="mb-1 flex items-center gap-2">
                          <Badge variant="outline" className={backendClassName(inst.storage_backend)}>
                            <Icon className="mr-1 h-3 w-3" />
                            {backend?.label ?? inst.storage_backend}
                          </Badge>
                          {inst.is_default && (
                            <Badge variant="outline" className="border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300">
                              <Star className="mr-1 h-3 w-3" />
                              {t(lang, "admin.settingsDefault")}
                            </Badge>
                          )}
                        </div>
                        <div className="line-clamp-1 text-sm font-medium">{inst.name}</div>
                        <div className="mt-1 line-clamp-1 font-mono text-xs text-muted-foreground">{inst.storage_key}</div>
                      </button>
                    );
                  })}
                  {isNew && (
                    <button
                      type="button"
                      className="w-full rounded-lg border border-primary bg-primary/5 px-3 py-2 text-left cursor-pointer"
                    >
                      <div className="mb-1 flex items-center gap-2">
                        <Badge variant="outline" className="border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-300">
                          <Plus className="mr-1 h-3 w-3" />
                          {t(lang, "admin.settingsNewBadge")}
                        </Badge>
                      </div>
                      <div className="line-clamp-1 text-sm font-medium">{form.name || "New Instance"}</div>
                      <div className="mt-1 text-xs text-muted-foreground">{t(lang, "admin.settingsNewHint")}</div>
                    </button>
                  )}
                  {!isNew && (config?.storage_configs.length ?? 0) === 0 && (
                    <div className="py-8 text-center text-sm text-muted-foreground">{t(lang, "admin.settingsNoInstances")}</div>
                  )}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="space-y-4 pt-6">
                <div className="flex items-center justify-between gap-2">
                  <div>
                    <h2 className="font-semibold">{isNew ? t(lang, "admin.settingsCreateStorage") : t(lang, "admin.settingsEditStorage")}</h2>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {isNew ? t(lang, "admin.settingsCreateStorageDescription") : selectedInst?.is_default ? t(lang, "admin.settingsDefaultStorageDescription") : t(lang, "admin.settingsEditStorageDescription")}
                    </p>
                  </div>
                  {!isNew && selectedInst && (
                    <Badge variant="outline" className={selectedInst.is_default ? "border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300" : "border-border bg-muted text-muted-foreground"}>
                      {selectedInst.is_default ? t(lang, "admin.settingsDefaultInstance") : t(lang, "admin.settingsNormalInstance")}
                    </Badge>
                  )}
                </div>

                <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                  <div className="space-y-2 md:col-span-2">
                    <Label htmlFor="inst-name">{t(lang, "admin.settingsName")}</Label>
                    <Input
                      id="inst-name"
                      value={form.name || ""}
                      onChange={(e) => updateField("name", e.target.value)}
                      className="h-8"
                    />
                  </div>

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

                  {!isNew && (
                    <div className="space-y-2">
                      <Label>{t(lang, "admin.settingsKey")}</Label>
                      <Input value={selectedKey} disabled className="h-8 font-mono text-xs" />
                    </div>
                  )}
                </div>

                <Separator />

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

                {backendType === "s3" && (
                  <div className="space-y-3">
                    <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
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
                    <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                      <div className="space-y-2">
                        <Label htmlFor="s3-access">{t(lang, "admin.settingsS3AccessKey")}</Label>
                        <Input id="s3-access" value={form.s3_access_key || ""} onChange={(e) => updateField("s3_access_key", e.target.value)} className="h-8" />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="s3-secret">{t(lang, "admin.settingsS3SecretKey")}</Label>
                        <Input id="s3-secret" type="password" value={form.s3_secret_key || ""} onChange={(e) => updateField("s3_secret_key", e.target.value)} className="h-8" />
                      </div>
                    </div>
                    <div className="flex flex-wrap items-center gap-4 rounded-lg border bg-muted/20 px-3 py-2">
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

                {backendType === "webdav" && (
                  <div className="space-y-3">
                    <div className="space-y-2">
                      <Label htmlFor="webdav-url">{t(lang, "admin.settingsWebdavUrl")}</Label>
                      <Input id="webdav-url" value={form.webdav_url || ""} onChange={(e) => updateField("webdav_url", e.target.value)} className="h-8" />
                    </div>
                    <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
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

                <div className="flex flex-wrap items-center gap-2">
                  <Button size="sm" onClick={handleSave} disabled={saving || !form.name} className="cursor-pointer gap-1">
                    {saving ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Check className="h-3.5 w-3.5" />}
                    {t(lang, "common.save")}
                  </Button>
                  {!isNew && !isDefault && (
                    <>
                      <Button size="sm" variant="outline" onClick={handleSetDefault} disabled={saving} className="cursor-pointer gap-1">
                        <Star className="h-3.5 w-3.5" />
                        {t(lang, "admin.settingsSetDefault")}
                      </Button>
                      <Button size="sm" variant="destructive" onClick={handleDelete} disabled={deleting} className="cursor-pointer gap-1">
                        {deleting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Trash2 className="h-3.5 w-3.5" />}
                        {t(lang, "admin.settingsDelete")}
                      </Button>
                    </>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
        <TabsContent value="runtime">
          <Card>
            <CardContent className="pt-6 space-y-5">
              <>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="runtime-public-url">{t(lang, "admin.settingsRuntimePublicUrl")}</Label>
                      <Input
                        id="runtime-public-url"
                        value={runtimeForm.public_base_url}
                        onChange={(e) => updateRuntimeField("public_base_url", e.target.value)}
                        placeholder="https://example.com"
                        className="h-8"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="runtime-max-size">{t(lang, "admin.settingsRuntimeMaxSize")}</Label>
                      <Input
                        id="runtime-max-size"
                        type="number"
                        min={0}
                        value={runtimeForm.max_upload_size_mb}
                        onChange={(e) => updateRuntimeField("max_upload_size_mb", Number(e.target.value))}
                        className="h-8"
                      />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="runtime-mime-types">{t(lang, "admin.settingsRuntimeMimeTypes")}</Label>
                    <Input
                      id="runtime-mime-types"
                      value={runtimeForm.allowed_mime_types.join(",")}
                      onChange={(e) => updateRuntimeField("allowed_mime_types", e.target.value.split(",").map((item) => item.trim()).filter(Boolean))}
                      placeholder="image/jpeg,image/png,image/gif,image/webp,image/avif"
                      className="h-8"
                    />
                    <p className="text-xs text-muted-foreground">{t(lang, "admin.settingsRuntimeMimeHint")}</p>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <label className="flex items-center gap-2 text-sm cursor-pointer">
                      <input
                        type="checkbox"
                        checked={runtimeForm.allow_storage_selection}
                        onChange={(e) => updateRuntimeField("allow_storage_selection", e.target.checked)}
                        className="cursor-pointer"
                      />
                      {t(lang, "admin.settingsRuntimeAllowStorageSelection")}
                    </label>
                    <label className="flex items-center gap-2 text-sm cursor-pointer">
                      <input
                        type="checkbox"
                        checked={runtimeForm.maintenance_mode}
                        onChange={(e) => updateRuntimeField("maintenance_mode", e.target.checked)}
                        className="cursor-pointer"
                      />
                      {t(lang, "admin.settingsRuntimeMaintenanceMode")}
                    </label>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="runtime-maintenance-message">{t(lang, "admin.settingsRuntimeMaintenanceMessage")}</Label>
                    <Input
                      id="runtime-maintenance-message"
                      value={runtimeForm.maintenance_message}
                      onChange={(e) => updateRuntimeField("maintenance_message", e.target.value)}
                      placeholder={t(lang, "admin.settingsRuntimeMaintenancePlaceholder")}
                      className="h-8"
                    />
                  </div>
                  <Button size="sm" onClick={handleSaveRuntimeSettings} disabled={systemSaving} className="cursor-pointer gap-1">
                    {systemSaving ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Check className="h-3.5 w-3.5" />}
                    {t(lang, "common.save")}
                  </Button>
                </>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="announcements">
          <AnnouncementManager token={token} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function backendClassName(backend: StorageInstance["storage_backend"]): string {
  switch (backend) {
    case "s3":
      return "border-cyan-500/30 bg-cyan-500/10 text-cyan-700 dark:text-cyan-300";
    case "webdav":
      return "border-violet-500/30 bg-violet-500/10 text-violet-700 dark:text-violet-300";
    default:
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300";
  }
}

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { Separator } from "@/components/ui/Separator";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { t } from "@/lib/i18n";

export function ApiPageContent() {
  const language = useUiPreferencesStore((state) => state.language);
  const lang = language;

  const endpoints = [
    {
      method: "POST",
      path: "/v1/image",
      desc: lang === "zh"
        ? "上传图片，multipart/form-data，字段 file（必填），可选 storage_key。Header X-Token 必填。"
        : "Upload image. multipart/form-data, field file (required), optional storage_key. Header X-Token required.",
      curl: `curl -X POST "<BASE>/v1/image" \\
  -H "X-Token: <your-client-token>" \\
  -F "file=@image.png" \\
  -F "storage_key=<optional-storage-key>"`,
    },
    {
      method: "GET",
      path: "/v1/runtime-settings",
      desc: lang === "zh"
        ? "获取公开运行时设置：上传限制、维护模式、多存储选项和可见存储实例。不会暴露密钥。"
        : "Get public runtime settings: upload limits, maintenance mode, storage selection, and visible storage options. Secrets are not exposed.",
      curl: `curl -X GET "<BASE>/v1/runtime-settings"`,
    },
    {
      method: "DELETE",
      path: "/i/:uid.avif",
      desc: lang === "zh"
        ? "删除当前 token 拥有的图片。Header X-Token 必填，只能删除该 token 上传的图片。"
        : "Delete image owned by the current token. Header X-Token required. Only deletes images uploaded with this token.",
      curl: `curl -X DELETE "<BASE>/i/<uid>.avif" \\
  -H "X-Token: <your-client-token>"`,
    },
  ];

  const methodColors: Record<string, string> = {
    POST: "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400",
    GET: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
    DELETE: "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400",
  };

  return (
    <div className="space-y-6 max-w-3xl" id="main-content">
      <div>
        <h1 className="text-xl font-bold">{t(lang, "api.title")}</h1>
      </div>

      {endpoints.map((ep, i) => (
        <Card key={i}>
          <CardHeader className="pb-2">
            <div className="flex items-center gap-2">
              <span
                className={`inline-flex items-center rounded px-2 py-0.5 text-xs font-bold ${methodColors[ep.method]}`}
              >
                {ep.method}
              </span>
              <CardTitle className="text-sm font-mono">{ep.path}</CardTitle>
            </div>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-sm text-muted-foreground">{ep.desc}</p>
            <Separator />
            <pre className="text-xs bg-muted p-3 rounded-md overflow-x-auto font-mono whitespace-pre-wrap">
              {ep.curl}
            </pre>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

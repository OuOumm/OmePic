"use client";

import { Badge } from "@/components/ui/Badge";
import { PageIntro } from "@/components/shared/PageLayout";
import { Card } from "@/components/ui/Card";
import { useUiTranslations } from "@/hooks/useUiPreferences";

export function ApiPageContent() {
  const t = useUiTranslations();
  const endpoints = [
    {
      method: "POST",
      path: "/v1/image",
      description: t.apiPage.uploadDescription,
      example: `curl -X POST http://localhost:8080/v1/image \\
  -H "X-Token: <your-token>" \\
  -F "file=@./example.png" \\
  -F "storage_key=<optional-storage-key>"`
    },
    {
      method: "GET",
      path: "/v1/storage-options",
      description: t.apiPage.storageOptionsDescription,
      example: "curl http://localhost:8080/v1/storage-options"
    },
    {
      method: "DELETE",
      path: "/i/:uid.avif",
      description: t.apiPage.deleteDescription,
      example: `curl -X DELETE http://localhost:8080/i/<uid>.avif \\
  -H "X-Token: <your-token>"`
    }
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      <PageIntro
        description={t.apiPage.uploadDescription}
        eyebrow={t.apiPage.eyebrow}
        title={t.apiPage.title}
      />

      <div className="grid gap-4">
        {endpoints.map((endpoint, index) => (
          <Card
            className="overflow-hidden transition-all duration-300 hover:-translate-y-0.5 hover:border-violet-300/50 hover:shadow-glow dark:hover:border-violet-400/30"
            key={`${endpoint.method}-${endpoint.path}`}
            style={{ animationDelay: `${index * 45}ms` }}
            variant="strong"
          >
            <div className="flex flex-col gap-4 border-b border-white/50 p-5 md:flex-row md:items-start md:justify-between dark:border-white/10">
              <div className="flex min-w-0 flex-wrap items-center gap-3">
                <Badge className={methodClass(endpoint.method)}>{endpoint.method}</Badge>
                <h2 className="break-all font-mono text-lg font-bold text-slate-900 dark:text-white">
                  {endpoint.path}
                </h2>
              </div>
              <p className="max-w-2xl text-sm leading-6 text-muted">{endpoint.description}</p>
            </div>
            <div className="bg-slate-950/95 p-4">
              <pre className="overflow-x-auto rounded-2xl border border-white/10 bg-[linear-gradient(110deg,rgba(15,23,42,0.96),rgba(30,41,59,0.96),rgba(15,23,42,0.96))] bg-[length:200%_100%] p-4 text-sm leading-6 text-slate-100 shadow-inner">
                <code>{endpoint.example}</code>
              </pre>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}

function methodClass(method: string) {
  switch (method) {
    case "POST":
      return "border-violet-300/35 bg-gradient-to-r from-violet-500 to-cyan-500 text-white";
    case "DELETE":
      return "border-rose-300/35 bg-gradient-to-r from-rose-500 to-orange-400 text-white";
    default:
      return "border-cyan-300/35 bg-gradient-to-r from-cyan-500 to-sky-500 text-white";
  }
}

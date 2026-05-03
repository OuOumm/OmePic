"use client";

import { PageIntro } from "@/components/shared/PageLayout";
import { Badge } from "@/components/ui/Badge";
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
            className="overflow-hidden"
            key={`${endpoint.method}-${endpoint.path}`}
            style={{ animationDelay: `${index * 45}ms` }}
            variant="strong"
          >
            <div className="flex flex-col gap-4 border-b border-border p-5 md:flex-row md:items-start md:justify-between">
              <div className="flex min-w-0 flex-wrap items-center gap-3">
                <Badge className={methodClass(endpoint.method)}>{endpoint.method}</Badge>
                <h2 className="break-all font-mono text-lg font-semibold text-foreground">
                  {endpoint.path}
                </h2>
              </div>
              <p className="max-w-2xl text-sm leading-6 text-muted-foreground">{endpoint.description}</p>
            </div>
            <div className="bg-slate-950/95 p-4">
              <pre className="overflow-x-auto rounded-md border border-slate-800 bg-slate-950 p-4 text-sm leading-6 text-slate-100 shadow-inner">
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
      return "border-transparent bg-primary text-primary-foreground";
    case "DELETE":
      return "border-transparent bg-destructive text-destructive-foreground";
    default:
      return "border-border bg-secondary text-secondary-foreground";
  }
}

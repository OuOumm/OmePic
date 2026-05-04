"use client";

import { useState } from "react";
import { Loader2, LockKeyhole } from "lucide-react";
import toast from "react-hot-toast";

import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { useUiTranslations } from "@/hooks/useUiPreferences";
import { adminLogin } from "@/lib/api";
import { useAdminSessionStore } from "@/stores/admin-session-store";

export function LoginForm() {
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const setToken = useAdminSessionStore((state) => state.setToken);
  const t = useUiTranslations();

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setErrorMessage(null);
    try {
      const token = await adminLogin(password);
      setToken(token);
      toast.success(t.admin.loginSuccessToast);
    } catch (error) {
      const nextError = error instanceof Error ? error.message : t.admin.loginFailed;
      setErrorMessage(nextError);
      toast.error(nextError);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="mx-auto flex min-h-[calc(100vh-170px)] max-w-md items-center justify-center animate-scale-in">
      <Card className="w-full p-6 sm:p-8" variant="strong">
        <div className="space-y-6">
          <div className="space-y-3 text-center">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-lg border border-border bg-muted text-muted-foreground">
              <LockKeyhole aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                {t.admin.loginEyebrow}
              </p>
              <h1 className="mt-2 text-2xl font-semibold tracking-tight text-foreground">
                {t.admin.loginTitle}
              </h1>
            </div>
          </div>
          <form className="space-y-4" onSubmit={handleSubmit}>
            <label className="block space-y-2">
              <span className="text-sm font-medium text-foreground">{t.admin.password}</span>
              <Input
                aria-describedby={errorMessage ? "admin-login-error" : undefined}
                aria-invalid={errorMessage ? true : undefined}
                autoComplete="current-password"
                onChange={(event) => setPassword(event.target.value)}
                placeholder={t.admin.passwordPlaceholder}
                type="password"
                value={password}
              />
            </label>
            {errorMessage ? (
              <p className="rounded-md border border-rose-400/30 bg-rose-500/10 p-3 text-sm text-danger" id="admin-login-error" role="alert">
                {errorMessage}
              </p>
            ) : null}
            <Button className="w-full" disabled={submitting} type="submit">
              {submitting ? <Loader2 aria-hidden="true" className="h-4 w-4 animate-spin" /> : <LockKeyhole aria-hidden="true" className="h-4 w-4" />}
              {submitting ? t.admin.signingIn : t.admin.signIn}
            </Button>
          </form>
        </div>
      </Card>
    </div>
  );
}

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
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
  const router = useRouter();
  const t = useUiTranslations();

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setErrorMessage(null);
    try {
      const token = await adminLogin(password);
      setToken(token);
      toast.success(t.admin.loginSuccessToast);
      router.push("/admin/dashboard");
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
      <Card className="relative w-full overflow-hidden p-6 sm:p-8" variant="strong">
        <div className="absolute inset-0 bg-gradient-to-br from-violet-500/10 via-transparent to-cyan-500/10" />
        <div className="relative space-y-6">
          <div className="space-y-3 text-center">
            <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-500 to-cyan-500 text-white shadow-lg shadow-violet-500/25">
              <LockIcon />
            </div>
            <div>
              <p className="text-xs font-bold uppercase tracking-[0.24em] text-violet-600 dark:text-violet-300">
                {t.admin.loginEyebrow}
              </p>
              <h1 className="mt-2 text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
                {t.admin.loginTitle}
              </h1>
            </div>
          </div>
          <form className="space-y-4" onSubmit={handleSubmit}>
            <label className="block space-y-2">
              <span className="text-sm font-semibold text-slate-800 dark:text-slate-200">{t.admin.password}</span>
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
              <p className="rounded-xl border border-rose-400/30 bg-rose-500/10 p-3 text-sm text-danger" id="admin-login-error" role="alert">
                {errorMessage}
              </p>
            ) : null}
            <Button className="w-full" disabled={submitting} type="submit">
              {submitting ? <SpinnerIcon /> : <LockIcon />}
              {submitting ? t.admin.signingIn : t.admin.signIn}
            </Button>
          </form>
        </div>
      </Card>
    </div>
  );
}

function LockIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M8 10V7a4 4 0 1 1 8 0v3m-9 0h10a1 1 0 0 1 1 1v9H6v-9a1 1 0 0 1 1-1Z" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function SpinnerIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth={4} />
      <path className="opacity-75" d="M4 12a8 8 0 0 1 8-8" stroke="currentColor" strokeLinecap="round" strokeWidth={4} />
    </svg>
  );
}

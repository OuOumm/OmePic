"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { Button } from "@/components/ui/Button";
import { Label } from "@/components/ui/Label";
import { adminLogin } from "@/lib/api";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { t } from "@/lib/i18n";
import { Loader2, Lock, AlertCircle } from "lucide-react";
import toast from "react-hot-toast";

export function LoginForm() {
  const language = useUiPreferencesStore((state) => state.language);
  const setToken = useAdminSessionStore((state) => state.setToken);
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!password.trim()) return;
    setLoading(true);
    setError("");
    try {
      const token = await adminLogin(password);
      setToken(token);
      toast.success(t(language, "admin.loginSuccess"));
    } catch (err) {
      const msg = err instanceof Error ? err.message : t(language, "admin.loginError");
      setError(msg);
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  const lang = language;

  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <div className="flex items-center gap-2">
            <Lock className="h-5 w-5 text-muted-foreground" />
            <CardTitle>{t(lang, "admin.login")}</CardTitle>
          </div>
          <CardDescription>OmePic Admin</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="admin-password">{t(lang, "admin.password")}</Label>
              <Input
                id="admin-password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter password"
                disabled={loading}
                autoFocus
              />
            </div>
            {error && (
              <div className="flex items-center gap-2 text-sm text-destructive" role="alert">
                <AlertCircle className="h-4 w-4" />
                {error}
              </div>
            )}
            <Button type="submit" className="w-full cursor-pointer" disabled={loading}>
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
              {loading ? t(lang, "admin.loggingIn") : t(lang, "admin.loginBtn")}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

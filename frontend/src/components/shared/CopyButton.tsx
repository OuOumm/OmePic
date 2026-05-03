"use client";

import { useState } from "react";
import { Check, Copy } from "lucide-react";
import toast from "react-hot-toast";

import { useUiTranslations } from "@/hooks/useUiPreferences";
import { Button } from "@/components/ui/Button";

type CopyButtonProps = {
  label: string;
  value: string;
};

export function CopyButton({ label, value }: CopyButtonProps) {
  const [copied, setCopied] = useState(false);
  const t = useUiTranslations();

  async function handleCopy() {
    await navigator.clipboard.writeText(value);
    setCopied(true);
    toast.success(t.copyButton.copied(label));
    window.setTimeout(() => setCopied(false), 1200);
  }

  return (
    <Button onClick={handleCopy} size="sm" variant={copied ? "primary" : "secondary"}>
      {copied ? <Check aria-hidden="true" className="h-4 w-4" /> : <Copy aria-hidden="true" className="h-4 w-4" />}
      {copied ? t.copyButton.copied(label) : t.copyButton.copy(label)}
    </Button>
  );
}

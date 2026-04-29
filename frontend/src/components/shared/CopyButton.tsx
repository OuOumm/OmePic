"use client";

import { useState } from "react";
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
      {copied ? <CheckIcon /> : <CopyIcon />}
      {copied ? t.copyButton.copied(label) : t.copyButton.copy(label)}
    </Button>
  );
}

function CopyIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M8 8h11v11H8z" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M5 16H4a1 1 0 0 1-1-1V5a2 2 0 0 1 2-2h10a1 1 0 0 1 1 1v1" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="m5 13 4 4L19 7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

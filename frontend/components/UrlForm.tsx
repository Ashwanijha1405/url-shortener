"use client";

import React, { useState, useRef, useEffect } from "react";
import { sounds } from "@/lib/sound";

interface UrlFormProps {
  onSubmit: (url: string) => Promise<void>;
  isLoading: boolean;
  /** Ref to the input element, used by the visualizer to animate from the form */
  inputRef?: React.RefObject<HTMLInputElement | null>;
}

export default function UrlForm({ onSubmit, isLoading, inputRef }: UrlFormProps) {
  const [url, setUrl] = useState("");
  const [error, setError] = useState<string | null>(null);
  const localInputRef = useRef<HTMLInputElement>(null);
  const activeRef = inputRef || localInputRef;

  // Focus input on mount
  useEffect(() => {
    activeRef.current?.focus();
  }, [activeRef]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = url.trim();

    if (!trimmed) {
      setError("Enter a URL to shorten.");
      return;
    }

    if (!trimmed.startsWith("http://") && !trimmed.startsWith("https://")) {
      setError("URL must start with http:// or https://");
      return;
    }

    try {
      new URL(trimmed);
    } catch {
      setError("Invalid URL format.");
      return;
    }

    setError(null);
    sounds.playClick();
    await onSubmit(trimmed);
  };

  return (
    <div className="w-full max-w-xl mx-auto">
      <form onSubmit={handleSubmit}>
        <div className="flex gap-2">
          <input
            ref={activeRef}
            type="url"
            value={url}
            onChange={(e) => {
              setUrl(e.target.value);
              if (error) setError(null);
            }}
            placeholder="https://example.com/long-url"
            disabled={isLoading}
            aria-label="URL to shorten"
            className="flex-1 h-11 px-4 rounded-lg bg-[var(--color-bg-elevated)] border border-[var(--color-border)] text-[var(--color-text)] text-sm font-mono placeholder:text-[var(--color-text-tertiary)] focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)]/40 focus:border-[var(--color-accent)]/60 disabled:opacity-50 transition-all"
          />
          <button
            type="submit"
            disabled={isLoading || !url.trim()}
            className="h-11 px-5 rounded-lg bg-[var(--color-accent)] text-[var(--color-bg)] text-sm font-medium whitespace-nowrap disabled:opacity-40 disabled:cursor-not-allowed hover:opacity-90 transition-opacity cursor-pointer flex items-center gap-2"
          >
            {isLoading ? (
              <>
                <svg className="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                <span>Shortening…</span>
              </>
            ) : (
              "Shorten"
            )}
          </button>
        </div>
      </form>

      {error && (
        <p className="mt-2 text-xs text-[var(--color-error)] font-mono pl-1">
          {error}
        </p>
      )}
    </div>
  );
}

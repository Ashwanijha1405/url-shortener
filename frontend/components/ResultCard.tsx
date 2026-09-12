"use client";

import { useState } from "react";
import { formatShortUrl, ShortenResponse } from "@/lib/api";
import { sounds } from "@/lib/sound";
import Waterfall from "@/components/Waterfall";

interface ResultCardProps {
  result: ShortenResponse;
  originalUrl: string;
  onReset?: () => void;
}

export default function ResultCard({ result, originalUrl, onReset }: ResultCardProps) {
  const [copied, setCopied] = useState(false);
  const fullUrl = formatShortUrl(result.short_code);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(fullUrl);
      sounds.playClick();
    } catch {
      const el = document.createElement("textarea");
      el.value = fullUrl;
      document.body.appendChild(el);
      el.select();
      document.execCommand("copy");
      document.body.removeChild(el);
      sounds.playClick();
    }
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="w-full max-w-xl mx-auto rounded-xl border border-[var(--color-border)] bg-[var(--color-bg-elevated)] p-5 shadow-lg shadow-black/5">
      {/* Short URL Row */}
      <div className="flex items-center gap-3 mb-4">
        <div className="flex-1 min-w-0">
          <p className="text-xs text-[var(--color-text-tertiary)] font-mono mb-1">Shortened URL</p>
          <p className="text-base sm:text-lg font-mono font-semibold text-[var(--color-accent)] truncate select-all">
            {fullUrl}
          </p>
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          <button
            onClick={handleCopy}
            className={`h-9 px-3.5 rounded-lg text-xs font-mono transition-all cursor-pointer flex items-center gap-1.5 ${
              copied
                ? "bg-[var(--color-success)] text-[var(--color-bg)] font-semibold shadow-sm"
                : "border border-[var(--color-border)] text-[var(--color-text)] hover:border-[var(--color-accent)]/60 hover:bg-[var(--color-bg)]"
            }`}
          >
            {copied ? (
              <>
                <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <polyline points="20 6 9 17 4 12" />
                </svg>
                Copied!
              </>
            ) : (
              <>
                <svg className="w-3.5 h-3.5 text-[var(--color-text-tertiary)]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                </svg>
                Copy
              </>
            )}
          </button>
          <a
            href={fullUrl}
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => sounds.playClick()}
            className="h-9 px-3 rounded-lg border border-[var(--color-border)] text-xs font-mono text-[var(--color-text-secondary)] hover:text-[var(--color-text)] hover:border-[var(--color-accent)]/60 hover:bg-[var(--color-bg)] transition-all flex items-center gap-1.5"
          >
            Visit
            <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
              <polyline points="15 3 21 3 21 9" />
              <line x1="10" y1="14" x2="21" y2="3" />
            </svg>
          </a>
        </div>
      </div>

      {/* Metadata summary row */}
      <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5 text-[11px] font-mono text-[var(--color-text-tertiary)] pt-3 border-t border-[var(--color-border-subtle)]">
        <span className="text-[var(--color-text-secondary)] truncate max-w-[180px] sm:max-w-[240px]" title={originalUrl}>
          {originalUrl}
        </span>

        {result.cache && (
          <span className={`px-1.5 py-0.5 rounded text-[10px] font-bold uppercase ${
            result.cache === "hit"
              ? "bg-[var(--color-success)]/15 text-[var(--color-success)] border border-[var(--color-success)]/20"
              : "bg-[var(--color-warning)]/15 text-[var(--color-warning)] border border-[var(--color-warning)]/20"
          }`}>
            Cache {result.cache}
          </span>
        )}

        {result.latency_ms !== undefined && (
          <span>{result.latency_ms}ms</span>
        )}

        {onReset && (
          <button
            onClick={() => {
              sounds.playClick();
              onReset();
            }}
            className="ml-auto text-[var(--color-text-tertiary)] hover:text-[var(--color-accent)] transition-colors cursor-pointer"
          >
            Shorten another →
          </button>
        )}
      </div>

      {/* Expandable Waterfall Inspector */}
      <Waterfall result={result} />
    </div>
  );
}

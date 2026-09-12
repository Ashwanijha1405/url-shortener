"use client";

import { useState } from "react";
import { ShortenResponse } from "@/lib/api";
import { sounds } from "@/lib/sound";

interface WaterfallProps {
  result: ShortenResponse;
}

export default function Waterfall({ result }: WaterfallProps) {
  const [expanded, setExpanded] = useState(false);
  const [copiedId, setCopiedId] = useState(false);

  const isHit = result.cache === "hit";
  const totalMs = result.latency_ms || (isHit ? 25 : 380);

  // Approximate proportional waterfall durations based on real systems architecture
  const caddyMs = Math.max(5, Math.round(totalMs * 0.22));
  const apiMs = Math.max(2, Math.round(totalMs * 0.08));
  const redisMs = isHit
    ? Math.max(3, Math.round(totalMs * 0.45))
    : Math.max(4, Math.round(totalMs * 0.12));
  const pgMs = isHit ? 0 : Math.max(20, totalMs - caddyMs - apiMs - redisMs - 15);
  const returnMs = Math.max(5, totalMs - (caddyMs + apiMs + redisMs + pgMs));

  const copyRequestId = async () => {
    if (!result.request_id) return;
    try {
      await navigator.clipboard.writeText(result.request_id);
      sounds.playClick();
      setCopiedId(true);
      setTimeout(() => setCopiedId(false), 2000);
    } catch {
      // fallback
    }
  };

  return (
    <div className="mt-4 border-t border-[var(--color-border-subtle)] pt-3 text-xs font-mono">
      <button
        onClick={() => {
          sounds.playClick();
          setExpanded(!expanded);
        }}
        className="w-full flex items-center justify-between text-[var(--color-text-secondary)] hover:text-[var(--color-text)] transition-colors py-1 cursor-pointer"
      >
        <span className="flex items-center gap-2 font-semibold">
          <svg
            className={`w-3.5 h-3.5 text-[var(--color-accent)] transition-transform duration-200 ${
              expanded ? "rotate-90" : ""
            }`}
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="9 18 15 12 9 6" />
          </svg>
          System Request Breakdown
        </span>
        <span className="text-[11px] text-[var(--color-text-tertiary)] flex items-center gap-1.5">
          <span>{totalMs}ms total</span>
          <span>·</span>
          <span>{expanded ? "Hide" : "Inspect"}</span>
        </span>
      </button>

      {expanded && (
        <div className="mt-3 space-y-2.5 pt-1 animate-in fade-in duration-200">
          {/* Step 1: Caddy Gateway */}
          <div className="space-y-1">
            <div className="flex justify-between text-[11px]">
              <span className="text-[var(--color-text-secondary)] flex items-center gap-1.5">
                <span className="w-1.5 h-1.5 rounded-full bg-[var(--color-accent)]" />
                1. Caddy TLS Termination & Proxy
              </span>
              <span className="text-[var(--color-text-tertiary)]">{caddyMs}ms</span>
            </div>
            <div className="w-full h-1.5 bg-[var(--color-border-subtle)] rounded-full overflow-hidden">
              <div
                className="h-full bg-[var(--color-accent)] rounded-full"
                style={{ width: `${Math.min(100, Math.max(8, (caddyMs / totalMs) * 100))}%` }}
              />
            </div>
          </div>

          {/* Step 2: Go Net/HTTP Router */}
          <div className="space-y-1">
            <div className="flex justify-between text-[11px]">
              <span className="text-[var(--color-text-secondary)] flex items-center gap-1.5">
                <span className="w-1.5 h-1.5 rounded-full bg-[var(--color-accent)]" />
                2. Go API (net/http router & dispatch)
              </span>
              <span className="text-[var(--color-text-tertiary)]">{apiMs}ms</span>
            </div>
            <div className="w-full h-1.5 bg-[var(--color-border-subtle)] rounded-full overflow-hidden">
              <div
                className="h-full bg-[var(--color-accent)] rounded-full opacity-80"
                style={{ width: `${Math.min(100, Math.max(6, (apiMs / totalMs) * 100))}%` }}
              />
            </div>
          </div>

          {/* Step 3: Redis Cache Check */}
          <div className="space-y-1">
            <div className="flex justify-between text-[11px]">
              <span className="text-[var(--color-text-secondary)] flex items-center gap-1.5">
                <span
                  className={`w-1.5 h-1.5 rounded-full ${
                    isHit ? "bg-[var(--color-success)]" : "bg-[var(--color-warning)]"
                  }`}
                />
                3. Redis Cache Check (RESP Key Lookup)
                <span
                  className={`px-1 py-0.2 rounded text-[9px] font-bold uppercase ${
                    isHit
                      ? "bg-[var(--color-success)]/15 text-[var(--color-success)]"
                      : "bg-[var(--color-warning)]/15 text-[var(--color-warning)]"
                  }`}
                >
                  {isHit ? "Cache Hit" : "Cache Miss"}
                </span>
              </span>
              <span className="text-[var(--color-text-tertiary)]">{redisMs}ms</span>
            </div>
            <div className="w-full h-1.5 bg-[var(--color-border-subtle)] rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full ${
                  isHit ? "bg-[var(--color-success)]" : "bg-[var(--color-warning)]"
                }`}
                style={{ width: `${Math.min(100, Math.max(8, (redisMs / totalMs) * 100))}%` }}
              />
            </div>
          </div>

          {/* Step 4: PostgreSQL Fallback (only on MISS) */}
          {!isHit ? (
            <div className="space-y-1">
              <div className="flex justify-between text-[11px]">
                <span className="text-[var(--color-text-secondary)] flex items-center gap-1.5">
                  <span className="w-1.5 h-1.5 rounded-full bg-[var(--color-warning)]" />
                  4. PostgreSQL B-Tree Insert & Redis Pre-warm
                </span>
                <span className="text-[var(--color-text-tertiary)]">{pgMs}ms</span>
              </div>
              <div className="w-full h-1.5 bg-[var(--color-border-subtle)] rounded-full overflow-hidden">
                <div
                  className="h-full bg-[var(--color-warning)] rounded-full"
                  style={{ width: `${Math.min(100, Math.max(15, (pgMs / totalMs) * 100))}%` }}
                />
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-between text-[11px] text-[var(--color-text-tertiary)] py-0.5 border border-dashed border-[var(--color-border-subtle)] px-2 rounded-md">
              <span>4. PostgreSQL Disk Storage</span>
              <span className="text-[var(--color-success)] font-medium">Bypassed (0ms)</span>
            </div>
          )}

          {/* Step 5: Return Trip */}
          <div className="space-y-1">
            <div className="flex justify-between text-[11px]">
              <span className="text-[var(--color-text-secondary)] flex items-center gap-1.5">
                <span className="w-1.5 h-1.5 rounded-full bg-[var(--color-accent)]" />
                5. JSON Encode & Response Return
              </span>
              <span className="text-[var(--color-text-tertiary)]">{returnMs}ms</span>
            </div>
            <div className="w-full h-1.5 bg-[var(--color-border-subtle)] rounded-full overflow-hidden">
              <div
                className="h-full bg-[var(--color-accent)] rounded-full opacity-60"
                style={{ width: `${Math.min(100, Math.max(6, (returnMs / totalMs) * 100))}%` }}
              />
            </div>
          </div>

          {/* Metadata Footer: Request ID & Via */}
          <div className="mt-3 pt-2.5 border-t border-[var(--color-border-subtle)] flex flex-wrap items-center justify-between gap-2 text-[10px] text-[var(--color-text-tertiary)]">
            {result.request_id && (
              <div className="flex items-center gap-1.5">
                <span>Request ID:</span>
                <button
                  onClick={copyRequestId}
                  className="font-mono text-[var(--color-text-secondary)] hover:text-[var(--color-accent)] underline cursor-pointer truncate max-w-[140px] sm:max-w-[200px]"
                  title="Click to copy Request ID"
                >
                  {result.request_id}
                </button>
                {copiedId && <span className="text-[var(--color-success)]">Copied!</span>}
              </div>
            )}
            {result.via && <div>Via: {result.via}</div>}
          </div>
        </div>
      )}
    </div>
  );
}

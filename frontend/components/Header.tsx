"use client";

import { useState, useEffect, useRef } from "react";
import { getBackendHealth } from "@/lib/api";
import { sounds } from "@/lib/sound";

type Theme = "dark" | "light";

export default function Header() {
  const [theme, setThemeState] = useState<Theme>("dark");
  const [backendOk, setBackendOk] = useState<boolean | null>(null);
  const [soundEnabled, setSoundEnabled] = useState(false);
  const initializedRef = useRef(false);

  // Sync theme from localStorage on mount (DOM-only side effect, no setState in render)
  useEffect(() => {
    if (initializedRef.current) return;
    initializedRef.current = true;
    const stored = localStorage.getItem("theme") as Theme | null;
    if (stored && stored !== "dark") {
      document.documentElement.setAttribute("data-theme", stored);
      queueMicrotask(() => setThemeState(stored));
    }
  }, []);

  // Subscribe to sound engine state
  useEffect(() => {
    const unsub = sounds.subscribe((enabled) => setSoundEnabled(enabled));
    return unsub;
  }, []);

  // Poll backend health every 30s
  useEffect(() => {
    let mounted = true;
    const check = async () => {
      const h = await getBackendHealth();
      if (mounted) setBackendOk(h.status === "ok");
    };
    check();
    const interval = setInterval(check, 30000);
    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, []);

  const toggleTheme = () => {
    sounds.playClick();
    const next: Theme = theme === "dark" ? "light" : "dark";
    document.documentElement.setAttribute("data-theme", next);
    localStorage.setItem("theme", next);
    setThemeState(next);
  };

  const toggleSound = () => {
    sounds.toggle();
  };

  return (
    <header className="w-full border-b border-[var(--color-border-subtle)] sticky top-0 z-50 bg-[var(--color-bg)]/90 backdrop-blur-md">
      <div className="max-w-3xl mx-auto px-5 h-12 flex items-center justify-between">
        {/* Left: brand */}
        <div className="flex items-center gap-2.5">
          <div className="w-6 h-6 rounded-md bg-[var(--color-accent)]/10 border border-[var(--color-accent)]/30 flex items-center justify-center">
            <svg
              className="w-3.5 h-3.5 text-[var(--color-accent)]"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
            </svg>
          </div>
          <span className="text-sm font-semibold tracking-tight text-[var(--color-text)]">
            URL Shortener
          </span>
        </div>

        {/* Right: status + sound toggle + theme toggle */}
        <div className="flex items-center gap-2.5">
          {/* Backend status dot */}
          <div className="flex items-center gap-1.5 text-xs text-[var(--color-text-tertiary)] mr-1">
            <span
              className={`w-1.5 h-1.5 rounded-full ${
                backendOk === null
                  ? "bg-[var(--color-text-tertiary)]"
                  : backendOk
                  ? "bg-[var(--color-success)]"
                  : "bg-[var(--color-error)]"
              }`}
            />
            <span className="hidden sm:inline font-mono text-[11px]">
              {backendOk === null ? "..." : backendOk ? "API" : "Offline"}
            </span>
          </div>

          {/* Sound FX toggle */}
          <button
            onClick={toggleSound}
            className={`w-7 h-7 rounded-md border flex items-center justify-center transition-colors cursor-pointer ${
              soundEnabled
                ? "border-[var(--color-accent)] text-[var(--color-accent)] bg-[var(--color-accent)]/10"
                : "border-[var(--color-border-subtle)] text-[var(--color-text-tertiary)] hover:text-[var(--color-text-secondary)] hover:border-[var(--color-border)]"
            }`}
            aria-label={soundEnabled ? "Mute sound effects" : "Enable sound effects"}
            title={soundEnabled ? "Sound effects enabled" : "Enable sound effects"}
          >
            {soundEnabled ? (
              <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
                <path d="M15.54 8.46a5 5 0 0 1 0 7.07" />
                <path d="M19.07 4.93a10 10 0 0 1 0 14.14" />
              </svg>
            ) : (
              <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
                <line x1="23" y1="9" x2="17" y2="15" />
                <line x1="17" y1="9" x2="23" y2="15" />
              </svg>
            )}
          </button>

          {/* Theme toggle */}
          <button
            onClick={toggleTheme}
            className="w-7 h-7 rounded-md border border-[var(--color-border-subtle)] flex items-center justify-center text-[var(--color-text-tertiary)] hover:text-[var(--color-text-secondary)] hover:border-[var(--color-border)] transition-colors cursor-pointer"
            aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} theme`}
            title={`Switch to ${theme === "dark" ? "light" : "dark"} theme`}
          >
            {theme === "dark" ? (
              <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="5" />
                <line x1="12" y1="1" x2="12" y2="3" /><line x1="12" y1="21" x2="12" y2="23" />
                <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" /><line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
                <line x1="1" y1="12" x2="3" y2="12" /><line x1="21" y1="12" x2="23" y2="12" />
                <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" /><line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
              </svg>
            ) : (
              <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
              </svg>
            )}
          </button>
        </div>
      </div>
    </header>
  );
}

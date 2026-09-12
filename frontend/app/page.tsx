"use client";

/**
 * Home Page — URL Shortener
 *
 * Idle: a minimal centered URL input. Nothing else.
 * On submit: the API call fires AND the animation starts simultaneously.
 *   - The forward journey (Client → Caddy → API) animates while the real
 *     network call is in flight.
 *   - At the API node, the animation awaits the real response.
 *   - The response journey uses real cache HIT/MISS data.
 *   - After animation completes, the result card slides in via GSAP.
 *
 * Normal users get a fast URL shortener.
 * Engineers get to watch the system work.
 */

import { useState, useRef, useEffect, useCallback } from "react";
import { gsap } from "gsap";
import Header from "@/components/Header";
import UrlForm from "@/components/UrlForm";
import ResultCard from "@/components/ResultCard";
import ErrorMessage from "@/components/ErrorMessage";
import { Visualizer, VisualizerHandle, PipelineResult, CacheOutcome } from "@/components/Visualizer";
import { shortenUrl, ShortenResponse, ApiError } from "@/lib/api";

export default function Home() {
  const visualizerRef = useRef<VisualizerHandle>(null);
  const heroRef = useRef<HTMLDivElement>(null);
  const resultRef = useRef<HTMLDivElement>(null);

  const [isProcessing, setIsProcessing] = useState(false);
  const [result, setResult] = useState<ShortenResponse | null>(null);
  const [submittedUrl, setSubmittedUrl] = useState("");
  const [error, setError] = useState<{ message: string; status?: number } | null>(null);

  // GSAP-driven hero collapse (GPU-accelerated transform, not CSS margin)
  useEffect(() => {
    if (!heroRef.current) return;
    const isCompact = isProcessing || !!result || !!error;
    gsap.to(heroRef.current, {
      y: isCompact ? -40 : 0,
      scale: isCompact ? 0.85 : 1,
      duration: 0.5,
      ease: "power3.out",
    });
  }, [isProcessing, result, error]);

  // GSAP entrance for result card
  useEffect(() => {
    if (result && resultRef.current) {
      gsap.fromTo(resultRef.current,
        { opacity: 0, y: 20, scale: 0.97 },
        { opacity: 1, y: 0, scale: 1, duration: 0.5, ease: "power3.out", delay: 0.1 }
      );
    }
  }, [result]);

  const handleShorten = useCallback(async (targetUrl: string) => {
    if (isProcessing) return;

    setIsProcessing(true);
    setError(null);
    setResult(null);
    setSubmittedUrl(targetUrl);

    // Track API result for showing the card after animation
    let apiResult: ShortenResponse | null = null;
    let apiError: { message: string; status?: number } | null = null;

    /**
     * Create the result promise — always resolves, never rejects.
     * The visualizer awaits this at the API sync point.
     * The .catch() converts errors into a PipelineResult with isError=true.
     */
    const resultPromise: Promise<PipelineResult> = shortenUrl(targetUrl)
      .then((response) => {
        apiResult = response;
        return {
          cacheOutcome: (response.cache || "miss") as CacheOutcome,
          shortCode: response.short_code,
          requestId: response.request_id,
          latencyMs: response.latency_ms,
        };
      })
      .catch((err) => {
        const message = err instanceof ApiError ? err.message : "Request failed";
        const status = err instanceof ApiError ? err.status : undefined;
        apiError = { message, status };
        return {
          cacheOutcome: "none" as CacheOutcome,
          isError: true,
          errorMessage: message,
        };
      });

    // Start animation IMMEDIATELY — runs concurrently with the API call
    if (visualizerRef.current) {
      await visualizerRef.current.runPipeline({
        url: targetUrl,
        resultPromise,
      });
    } else {
      // Fallback: if visualizer isn't mounted, just await the API
      await resultPromise;
    }

    // Animation is done. Show the result or error card.
    if (apiResult) {
      setResult(apiResult);
    } else if (apiError) {
      setError(apiError);
    }

    setIsProcessing(false);
  }, [isProcessing]);

  const handleReset = useCallback(() => {
    setResult(null);
    setSubmittedUrl("");
    setError(null);
    visualizerRef.current?.reset();
  }, []);

  return (
    <div className="min-h-screen flex flex-col">
      <Header />

      <main className="flex-1 flex flex-col items-center justify-start px-4 sm:px-6">
        {/* Hero — GSAP-driven transform, not CSS margin */}
        <div
          ref={heroRef}
          className="w-full max-w-xl text-center mt-[18vh] mb-8"
        >
          <h1 className="text-3xl sm:text-4xl font-semibold tracking-tight text-[var(--color-text)]">
            Shorten any URL
          </h1>
          {!result && !isProcessing && !error && (
            <p className="mt-3 text-sm text-[var(--color-text-secondary)]">
              Paste a link and watch it flow through the system.
            </p>
          )}
        </div>

        {/* URL Input Form */}
        <div className="w-full mb-6">
          <UrlForm onSubmit={handleShorten} isLoading={isProcessing} />
        </div>

        {/* Architecture Visualizer */}
        <div className="w-full mb-6">
          <Visualizer ref={visualizerRef} />
        </div>

        {/* Error */}
        {error && (
          <div className="w-full mb-4">
            <ErrorMessage
              message={error.message}
              status={error.status}
              onDismiss={() => setError(null)}
              onRetry={() => { if (submittedUrl) handleShorten(submittedUrl); }}
            />
          </div>
        )}

        {/* Result Card (GSAP entrance via ref) */}
        {result && (
          <div ref={resultRef} className="w-full mb-8" style={{ opacity: 0 }}>
            <ResultCard
              result={result}
              originalUrl={submittedUrl}
              onReset={handleReset}
            />
          </div>
        )}
      </main>

      <footer className="mt-auto py-6 text-center text-[11px] font-mono text-[var(--color-text-tertiary)]">
        Go · Redis · PostgreSQL · Caddy
      </footer>
    </div>
  );
}

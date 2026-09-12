/**
 * Centralized API client for the Go URL shortener backend.
 * All network calls go through Next.js API routes to avoid CORS issues.
 */

export interface ShortenResponse {
  short_code: string;
  cache?: "hit" | "miss";
  request_id?: string;
  via?: string;
  latency_ms?: number;
}

export interface BackendHealth {
  status: "ok" | "degraded" | "offline";
  via?: string;
  latency_ms: number;
  request_id?: string;
  backend_url?: string;
  error?: string;
}

export class ApiError extends Error {
  constructor(
    message: string,
    public status?: number,
    public requestId?: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

/** Base URL for constructing clickable short URLs */
export function getShortUrlBase(): string {
  return process.env.NEXT_PUBLIC_API_URL || "http://16.171.135.9";
}

/** Format a short code into a full redirect URL */
export function formatShortUrl(shortCode: string): string {
  return `${getShortUrlBase().replace(/\/+$/, "")}/${shortCode}`;
}

/** Shorten a URL via the Next.js proxy route */
export async function shortenUrl(
  targetUrl: string,
  signal?: AbortSignal
): Promise<ShortenResponse> {
  const trimmed = targetUrl.trim();
  if (!trimmed) throw new ApiError("URL cannot be empty", 400);

  let response: Response;
  try {
    response = await fetch("/api/shorten", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url: trimmed }),
      signal,
    });
  } catch (err: unknown) {
    if (err instanceof DOMException && err.name === "AbortError") throw err;
    throw new ApiError(
      err instanceof Error ? err.message : "Network error",
      0
    );
  }

  let json: Record<string, unknown>;
  try {
    json = await response.json();
  } catch {
    throw new ApiError(`Malformed response (HTTP ${response.status})`, response.status);
  }

  if (!response.ok) {
    const msg =
      (typeof json.error === "string" && json.error) ||
      (typeof json.message === "string" && json.message) ||
      `HTTP ${response.status}`;
    throw new ApiError(msg, response.status, typeof json.request_id === "string" ? json.request_id : undefined);
  }

  if (typeof json.short_code !== "string" || !json.short_code) {
    throw new ApiError("Invalid response: missing short_code", response.status);
  }

  return {
    short_code: json.short_code,
    cache: json.cache === "hit" || json.cache === "miss" ? json.cache : undefined,
    request_id: typeof json.request_id === "string" ? json.request_id : undefined,
    via: typeof json.via === "string" ? json.via : undefined,
    latency_ms: typeof json.latency_ms === "number" ? json.latency_ms : undefined,
  };
}

/** Check Go backend health via the Next.js proxy route */
export async function getBackendHealth(signal?: AbortSignal): Promise<BackendHealth> {
  try {
    const res = await fetch("/api/health", { cache: "no-store", signal });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      return { status: "degraded", latency_ms: data.latency_ms || 0, via: data.via, error: data.error };
    }
    return {
      status: data.status === "ok" ? "ok" : "degraded",
      via: data.via,
      latency_ms: data.latency_ms || 0,
      request_id: data.request_id,
      backend_url: data.backend_url,
    };
  } catch (err: unknown) {
    return { status: "offline", latency_ms: 0, error: err instanceof Error ? err.message : "Unreachable" };
  }
}

import { NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL || "http://api:8080";

export async function GET() {
  const startTime = Date.now();
  try {
    const res = await fetch(`${BACKEND_URL}/health`, {
      method: "GET",
      signal: AbortSignal.timeout(5000),
      cache: "no-store",
    });

    const latencyMs = Date.now() - startTime;
    const via = res.headers.get("via") || "Direct";
    const requestId = res.headers.get("x-request-id") || undefined;

    if (!res.ok) {
      return NextResponse.json(
        { status: "degraded", latency_ms: latencyMs, via, request_id: requestId },
        { status: res.status }
      );
    }

    const data = await res.json().catch(() => ({ status: "ok" }));
    return NextResponse.json({
      status: data.status || "ok",
      via,
      latency_ms: latencyMs,
      request_id: requestId,
      backend_url: BACKEND_URL,
    });
  } catch (err: unknown) {
    const latencyMs = Date.now() - startTime;
    return NextResponse.json(
      {
        status: "offline",
        latency_ms: latencyMs,
        error: err instanceof Error ? err.message : "Unreachable",
        backend_url: BACKEND_URL,
      },
      { status: 503 }
    );
  }
}

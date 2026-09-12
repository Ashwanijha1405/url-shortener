/**
 * Node Layouts — SVG coordinate system & Infrastructure Details
 *
 * Defines node positions, connections, and metadata for the SVG visualization.
 * Desktop: horizontal left-to-right flow (800 × 260)
 * Mobile: vertical top-to-bottom flow (300 × 560)
 */

import { SystemNodeId, NodePosition } from "./types";

export type LayoutMode = "desktop" | "mobile";

export interface NodeDetail {
  tech: string;
  role: string;
  protocol: string;
  desc: string;
}

export interface NodeMeta {
  id: SystemNodeId;
  label: string;
  sublabel: string;
  detail: NodeDetail;
}

export const NODES: NodeMeta[] = [
  {
    id: "client",
    label: "Client",
    sublabel: "Next.js",
    detail: {
      tech: "Next.js 16 · React 19",
      role: "Client Browser",
      protocol: "HTTPS / TLS 1.3",
      desc: "Initiates short URL creation, normalizes inputs, and streams response JSON.",
    },
  },
  {
    id: "caddy",
    label: "Caddy",
    sublabel: "Proxy",
    detail: {
      tech: "Caddy v2.8",
      role: "Reverse Proxy & TLS Gateway",
      protocol: "Reverse Proxy (HTTP/2 ➔ HTTP/1.1)",
      desc: "Automated TLS 1.3 termination, adds X-Request-Id & Via headers, load balances to Go net/http.",
    },
  },
  {
    id: "api",
    label: "Go API",
    sublabel: "net/http",
    detail: {
      tech: "Go 1.22 · net/http",
      role: "Backend Application Server",
      protocol: "REST API Endpoint",
      desc: "URL validation, collision detection, and routing to Redis cache or PostgreSQL store.",
    },
  },
  {
    id: "redis",
    label: "Redis",
    sublabel: "Cache",
    detail: {
      tech: "Redis 7.2 Alpine",
      role: "In-Memory Deduplication Cache",
      protocol: "RESP (REdis Serialization Protocol)",
      desc: "Sub-millisecond key-value lookups with 24h TTL. Cache hits bypass database completely.",
    },
  },
  {
    id: "postgres",
    label: "PostgreSQL",
    sublabel: "Storage",
    detail: {
      tech: "PostgreSQL 16 Alpine",
      role: "Durable ACID Relational Store",
      protocol: "SQL / pgx driver",
      desc: "Persists short_code ➔ original_url mappings with unique B-Tree indexing.",
    },
  },
];

/** SVG icon paths for each node, centered at (0,0), ~12 unit scale */
export const NODE_ICON_PATHS: Record<SystemNodeId, { d: string; filled: boolean }> = {
  // Browser window
  client: {
    d: "M-5.5 -3.5h11a1 1 0 011 1v5.5a1 1 0 01-1 1h-11a1 1 0 01-1-1v-5.5a1 1 0 011-1zM-6.5 -1.5h13",
    filled: false,
  },
  // Shield
  caddy: {
    d: "M0 -5.5l5.5 2.5v3.5c0 3-2.2 5.5-5.5 6.5-3.3-1-5.5-3.5-5.5-6.5v-3.5L0 -5.5z",
    filled: false,
  },
  // Lightning bolt
  api: {
    d: "M1.5 -5.5L-3 0h4.5L0 5.5 5 0H1l.5-5.5z",
    filled: true,
  },
  // Layers / stack
  redis: {
    d: "M0 -4l6.5 3L0 2l-6.5-3L0 -4zM-6.5 0L0 3l6.5-3M-6.5 2.5L0 5.5l6.5-3",
    filled: false,
  },
  // Database cylinder
  postgres: {
    d: "M-4.5 -3.5c0-1.4 2-2.5 4.5-2.5s4.5 1.1 4.5 2.5v7c0 1.4-2 2.5-4.5 2.5s-4.5-1.1-4.5-2.5v-7zM-4.5 -3.5c0 1.4 2 2.5 4.5 2.5s4.5-1.1 4.5-2.5",
    filled: false,
  },
};

export interface ConnectionDef {
  from: SystemNodeId;
  to: SystemNodeId;
  label: string;
}

export const CONNECTIONS: ConnectionDef[] = [
  { from: "client",  to: "caddy",    label: "HTTPS" },
  { from: "caddy",   to: "api",      label: "Reverse Proxy" },
  { from: "api",     to: "redis",    label: "RESP" },
  { from: "redis",   to: "postgres", label: "SQL Fallback" },
];

const DESKTOP: Record<SystemNodeId, NodePosition> = {
  client:   { x: 80,  y: 120 },
  caddy:    { x: 240, y: 120 },
  api:      { x: 400, y: 120 },
  redis:    { x: 560, y: 120 },
  postgres: { x: 720, y: 120 },
};

const MOBILE: Record<SystemNodeId, NodePosition> = {
  client:   { x: 150, y: 55  },
  caddy:    { x: 150, y: 165 },
  api:      { x: 150, y: 275 },
  redis:    { x: 150, y: 385 },
  postgres: { x: 150, y: 495 },
};

export function getNodePositions(mode: LayoutMode): Record<SystemNodeId, NodePosition> {
  return mode === "desktop" ? DESKTOP : MOBILE;
}

export const DESKTOP_VIEWBOX = "0 0 800 260";
export const MOBILE_VIEWBOX = "0 0 300 560";

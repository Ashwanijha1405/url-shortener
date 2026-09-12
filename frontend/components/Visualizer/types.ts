/**
 * Architecture Visualizer — Type Definitions
 *
 * Core types for the SVG-based network visualization.
 * The key architectural pattern: the visualizer receives a Promise for the
 * API result, starts the forward animation immediately, and awaits the
 * response at a natural sync point (the API node).
 */

export type SystemNodeId = "client" | "caddy" | "api" | "redis" | "postgres";

export type CacheOutcome = "hit" | "miss" | "none";

/** Result data fed into the visualizer from the actual API response */
export interface PipelineResult {
  cacheOutcome: CacheOutcome;
  shortCode?: string;
  requestId?: string;
  latencyMs?: number;
  isError?: boolean;
  errorMessage?: string;
}

/**
 * Options passed into runPipeline.
 * 
 * `resultPromise` is awaited at the sync point (API node), so the forward
 * journey animates concurrently with the real network call.
 */
export interface PipelineOptions {
  url: string;
  resultPromise: Promise<PipelineResult>;
}

/** Imperative handle exposed by the Visualizer via forwardRef */
export interface VisualizerHandle {
  runPipeline: (options: PipelineOptions) => Promise<void>;
  reset: () => void;
  replay: () => void;
}

export interface NodePosition {
  x: number;
  y: number;
}

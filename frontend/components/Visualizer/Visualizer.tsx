"use client";

/**
 * Visualizer — GSAP-Driven SVG Architecture Animation
 *
 * An interactive, cinematic network architecture visualizer:
 *   Client ➔ Caddy ➔ Go API ➔ Redis ⇄ PostgreSQL
 *
 * Key features:
 * - Dynamic Spotlight Glow that follows the packet in real-time
 * - Component personality tweens: Caddy TLS shield shimmer, Go API sparks,
 *   PostgreSQL mechanical disk stamp
 * - Interactive Node Inspector hover cards with infrastructure specs
 * - Replay & Speed controls (1x, 0.5x slow-mo, 0.25x deep inspect)
 * - Optional Web Audio API micro-sound synthesis (muted by default)
 * - True Redis CACHE HIT bypasses PostgreSQL completely
 */

import {
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from "react";
import { gsap } from "gsap";
import { SystemNodeId, PipelineOptions, PipelineResult, VisualizerHandle } from "./types";
import {
  NODES,
  CONNECTIONS,
  NODE_ICON_PATHS,
  getNodePositions,
  DESKTOP_VIEWBOX,
  MOBILE_VIEWBOX,
  LayoutMode,
  NodeMeta,
} from "./layout";
import { sounds } from "@/lib/sound";

/* ═══════════════════════════════════════════════════════════════════════
   Constants
   ═══════════════════════════════════════════════════════════════════════ */

const NODE_RADIUS = 18;
const PACKET_MAX_CHARS = 30;

function truncateUrl(url: string): string {
  const s = url.replace(/^https?:\/\//, "");
  return s.length <= PACKET_MAX_CHARS ? s : s.substring(0, PACKET_MAX_CHARS) + "…";
}

function dist(ax: number, ay: number, bx: number, by: number): number {
  return Math.sqrt((bx - ax) ** 2 + (by - ay) ** 2);
}

/* ═══════════════════════════════════════════════════════════════════════
   Component
   ═══════════════════════════════════════════════════════════════════════ */

const Visualizer = forwardRef<VisualizerHandle>(function Visualizer(_, ref) {
  /* ── Element Refs ───────────────────────────────────────────────────── */
  const svgRef = useRef<SVGSVGElement>(null);
  const packetGroup = useRef<SVGGElement>(null);
  const packetBg = useRef<SVGRectElement>(null);
  const packetText = useRef<SVGTextElement>(null);

  // Spotlight refs
  const spotlightRef = useRef<SVGCircleElement>(null);
  const spotlightStop0 = useRef<SVGStopElement>(null);
  const spotlightStop1 = useRef<SVGStopElement>(null);

  // Cache badge refs
  const cacheBadge = useRef<SVGGElement>(null);
  const cacheBadgeRect = useRef<SVGRectElement>(null);
  const cacheBadgeText = useRef<SVGTextElement>(null);

  // Component personality refs
  const caddyShield = useRef<SVGCircleElement>(null);
  const postgresIconGroup = useRef<SVGGElement>(null);
  const apiSparks = useRef<SVGCircleElement[]>([]);

  // Per-node maps
  const circles = useRef<Partial<Record<SystemNodeId, SVGCircleElement>>>({});
  const rings = useRef<Partial<Record<SystemNodeId, SVGCircleElement>>>({});

  // Per-connection flows & particles
  const flows = useRef<Partial<Record<string, SVGLineElement>>>({});
  const particles = useRef<Partial<Record<string, (SVGCircleElement | null)[]>>>({});

  // Timeline & State management
  const tlRef = useRef<gsap.core.Timeline | null>(null);
  const processingTween = useRef<gsap.core.Tween | null>(null);
  const lastOptionsRef = useRef<PipelineOptions | null>(null);
  const lastResultRef = useRef<PipelineResult | null>(null);

  /* ── UI State ───────────────────────────────────────────────────────── */
  const [visible, setVisible] = useState(false);
  const [packetLabel, setPacketLabel] = useState("");
  const [statusText, setStatusText] = useState("");
  const [layout, setLayout] = useState<LayoutMode>("desktop");
  const [speed, setSpeed] = useState<1 | 0.5 | 0.25>(1);
  const [hasCompleted, setHasCompleted] = useState(false);
  const [hoveredNode, setHoveredNode] = useState<NodeMeta | null>(null);

  // Speed multiplier ref so animations can read the latest value
  const speedRef = useRef<number>(1);
  speedRef.current = speed;

  /* ── Responsive layout ─────────────────────────────────────────────── */
  useEffect(() => {
    const check = () => setLayout(window.innerWidth < 768 ? "mobile" : "desktop");
    check();
    window.addEventListener("resize", check);
    return () => window.removeEventListener("resize", check);
  }, []);

  const pos = getNodePositions(layout);
  const viewBox = layout === "desktop" ? DESKTOP_VIEWBOX : MOBILE_VIEWBOX;

  /* ── Cleanup on unmount ─────────────────────────────────────────────── */
  useEffect(() => {
    return () => {
      tlRef.current?.kill();
      processingTween.current?.kill();
    };
  }, []);

  const killAll = useCallback(() => {
    tlRef.current?.kill();
    tlRef.current = null;
    processingTween.current?.kill();
    processingTween.current = null;
  }, []);

  /* ── Reset to idle ──────────────────────────────────────────────────── */
  const resetAll = useCallback(() => {
    killAll();
    setVisible(false);
    setPacketLabel("");
    setStatusText("");
    setHasCompleted(false);
    setHoveredNode(null);

    if (svgRef.current) gsap.set(svgRef.current, { opacity: 0, y: 12 });
    if (packetGroup.current) gsap.set(packetGroup.current, { scale: 0, opacity: 0 });
    if (cacheBadge.current) gsap.set(cacheBadge.current, { scale: 0, opacity: 0 });
    if (spotlightRef.current) gsap.set(spotlightRef.current, { opacity: 0 });

    for (const node of NODES) {
      const c = circles.current[node.id];
      if (c) {
        gsap.set(c, {
          attr: {
            r: NODE_RADIUS,
            fill: "var(--color-node-bg)",
            stroke: "var(--color-node-border)",
            "stroke-width": 1.5,
          },
        });
      }
      const r = rings.current[node.id];
      if (r) gsap.set(r, { opacity: 0 });
    }

    for (const conn of CONNECTIONS) {
      const key = `${conn.from}-${conn.to}`;
      const f = flows.current[key];
      if (f) gsap.set(f, { opacity: 0 });
    }
  }, [killAll]);

  /* ═══════════════════════════════════════════════════════════════════════
     GSAP Animation Helpers
     ═══════════════════════════════════════════════════════════════════════ */

  function getConnKey(fromId: SystemNodeId, toId: SystemNodeId): string {
    const direct = `${fromId}-${toId}`;
    if (flows.current[direct]) return direct;
    const reverse = `${toId}-${fromId}`;
    if (flows.current[reverse]) return reverse;
    return direct;
  }

  function pulseNode(
    tl: gsap.core.Timeline,
    nodeId: SystemNodeId,
    color: string,
    position: string
  ) {
    const circle = circles.current[nodeId];
    const ring = rings.current[nodeId];
    if (!circle) return;

    tl.set(circle, { attr: { fill: color, stroke: color, "stroke-width": 2.5 } }, position);

    tl.to(circle, {
      attr: { r: NODE_RADIUS * 1.18 },
      duration: 0.1,
      ease: "power2.out",
    }, position);
    tl.to(circle, {
      attr: { r: NODE_RADIUS },
      duration: 0.35,
      ease: "elastic.out(1, 0.4)",
    }, `${position}+=0.1`);

    if (ring) {
      tl.fromTo(ring,
        { attr: { r: NODE_RADIUS, "stroke-width": 2 }, opacity: 0.5, stroke: color },
        { attr: { r: NODE_RADIUS + 22 }, opacity: 0, duration: 0.65, ease: "power2.out" },
        position
      );
    }
  }

  function moveSpotlight(tl: gsap.core.Timeline, targetNode: SystemNodeId, position: string, duration: number) {
    if (!spotlightRef.current) return;
    const tp = pos[targetNode];
    tl.to(spotlightRef.current, {
      attr: { cx: tp.x, cy: tp.y },
      duration,
      ease: "power2.inOut",
    }, position);
  }

  function setSpotlightColor(color: string) {
    if (spotlightStop0.current) spotlightStop0.current.setAttribute("stop-color", color);
    if (spotlightStop1.current) spotlightStop1.current.setAttribute("stop-color", color);
  }

  function flowConn(
    tl: gsap.core.Timeline,
    fromId: SystemNodeId,
    toId: SystemNodeId,
    position: string,
    duration: number
  ) {
    const key = getConnKey(fromId, toId);
    const line = flows.current[key];
    if (!line) return;

    const fp = pos[fromId];
    const tp = pos[toId];
    const len = dist(fp.x, fp.y, tp.x, tp.y);
    const dash = Math.round(len * 0.12);
    const gap = Math.round(len * 0.1);

    const isForward = key === `${fromId}-${toId}`;

    if (isForward) {
      tl.set(line, {
        attr: { "stroke-dasharray": `${dash} ${gap}`, "stroke-dashoffset": len.toString() },
        opacity: 0.75,
      }, position);

      tl.to(line, {
        attr: { "stroke-dashoffset": "0" },
        duration,
        ease: "none",
      }, position);
    } else {
      tl.set(line, {
        attr: { "stroke-dasharray": `${dash} ${gap}`, "stroke-dashoffset": "0" },
        opacity: 0.75,
      }, position);

      tl.to(line, {
        attr: { "stroke-dashoffset": len.toString() },
        duration,
        ease: "none",
      }, position);
    }
  }

  function fadeConn(
    tl: gsap.core.Timeline,
    fromId: SystemNodeId,
    toId: SystemNodeId,
    position: string
  ) {
    const key = getConnKey(fromId, toId);
    const line = flows.current[key];
    if (line) tl.to(line, { opacity: 0, duration: 0.12 }, position);
  }

  function fireParticles(
    tl: gsap.core.Timeline,
    fromId: SystemNodeId,
    toId: SystemNodeId,
    position: string,
    duration: number
  ) {
    const key = getConnKey(fromId, toId);
    const ps = particles.current[key];
    if (!ps) return;

    const fp = pos[fromId];
    const tp = pos[toId];

    ps.forEach((p, i) => {
      if (!p) return;
      const delay = i * 0.08;
      tl.fromTo(p,
        { attr: { cx: fp.x, cy: fp.y }, opacity: 0.75 },
        { attr: { cx: tp.x, cy: tp.y }, opacity: 0, duration: duration * 0.85, ease: "power1.in" },
        `${position}+=${delay}`
      );
    });
  }

  function crossfadeLabel(tl: gsap.core.Timeline, label: string, position: string) {
    tl.to(packetText.current, { opacity: 0, duration: 0.06 }, position);
    tl.call(() => {
      setPacketLabel(label);
      if (packetText.current) packetText.current.textContent = label;
    }, [], `${position}+=0.06`);
    tl.to(packetText.current, { opacity: 1, duration: 0.08 }, `${position}+=0.06`);
  }

  function morphPacketColor(tl: gsap.core.Timeline, color: string, position: string) {
    tl.to(packetBg.current, { opacity: 0.4, duration: 0.08 }, position);
    tl.set(packetBg.current, { attr: { fill: color } }, `${position}+=0.08`);
    tl.to(packetBg.current, { opacity: 0.95, duration: 0.12 }, `${position}+=0.08`);
  }

  /* ═══════════════════════════════════════════════════════════════════════
     Main Pipeline — Full Animation
     ═══════════════════════════════════════════════════════════════════════ */

  const runPipeline = useCallback(
    async (options: PipelineOptions) => {
      killAll();
      lastOptionsRef.current = options;
      setHasCompleted(false);

      const reducedMotion =
        typeof window !== "undefined" &&
        window.matchMedia("(prefers-reduced-motion: reduce)").matches;

      // Apply speed factor: 1x -> 1.0, 0.5x -> 2.0, 0.25x -> 3.5
      const speedMult = 1 / speedRef.current;
      const baseDur = reducedMotion ? 0.02 : 0.32;
      const dur = baseDur * speedMult;

      // ── Initialize elements ──
      const initialLabel = truncateUrl(options.url);
      setPacketLabel(initialLabel);
      setStatusText("");
      setVisible(true);

      if (svgRef.current) gsap.set(svgRef.current, { opacity: 0, y: 12 });

      // Packet starts at client
      if (packetGroup.current) gsap.set(packetGroup.current, { x: pos.client.x, y: pos.client.y - 30, scale: 0, opacity: 0 });
      if (packetBg.current) gsap.set(packetBg.current, { attr: { fill: "var(--color-packet-bg)" }, opacity: 0.95 });
      if (packetText.current) packetText.current.textContent = initialLabel;

      // Spotlight setup
      setSpotlightColor("var(--color-accent)");
      if (spotlightRef.current) {
        gsap.set(spotlightRef.current, {
          attr: { cx: pos.client.x, cy: pos.client.y, r: 65 },
          opacity: 0.5,
        });
      }

      // Reset nodes
      for (const node of NODES) {
        const c = circles.current[node.id];
        if (c) gsap.set(c, { attr: { r: 0, fill: "var(--color-node-bg)", stroke: "var(--color-node-border)", "stroke-width": 1.5 } });
        const r = rings.current[node.id];
        if (r) gsap.set(r, { opacity: 0 });
      }

      for (const conn of CONNECTIONS) {
        const f = flows.current[`${conn.from}-${conn.to}`];
        if (f) gsap.set(f, { opacity: 0 });
      }

      if (cacheBadge.current) gsap.set(cacheBadge.current, { scale: 0, opacity: 0 });

      /* ──────────────────────────────────────────────────────────────────
         PHASE 1: FORWARD TIMELINE (Concurrent with API call)
         ────────────────────────────────────────────────────────────────── */

      const fwd = gsap.timeline();
      tlRef.current = fwd;

      // ▸ Entrance: SVG fades in
      fwd.to(svgRef.current, { opacity: 1, y: 0, duration: 0.4, ease: "power3.out" });

      // ▸ Nodes stagger-appear
      fwd.addLabel("nodes", 0.15);
      NODES.forEach((node, i) => {
        const c = circles.current[node.id];
        if (c) {
          fwd.to(c, {
            attr: { r: NODE_RADIUS },
            duration: 0.28,
            ease: "back.out(2)",
          }, `nodes+=${i * 0.05}`);
        }
      });

      // ▸ Packet creation at Client (sound: Launch)
      fwd.addLabel("packet", "nodes+=0.3");
      fwd.call(() => sounds.playLaunch(), [], "packet");
      fwd.fromTo(packetGroup.current,
        { scale: 0, opacity: 0, y: pos.client.y - 65 },
        { scale: 1, opacity: 1, y: pos.client.y - 30, duration: 0.35, ease: "back.out(2.5)" },
        "packet"
      );

      // ▸ Step 1: Client → Caddy
      fwd.addLabel("toCaddy", "packet+=0.25");
      fwd.call(() => {
        sounds.playHop();
        setStatusText("Dispatching HTTPS POST request…");
      }, [], "toCaddy");

      pulseNode(fwd, "client", "var(--color-accent)", "toCaddy");
      flowConn(fwd, "client", "caddy", "toCaddy", dur);
      fireParticles(fwd, "client", "caddy", "toCaddy", dur);
      moveSpotlight(fwd, "caddy", "toCaddy", dur);

      fwd.to(packetGroup.current, {
        x: pos.caddy.x, y: pos.caddy.y - 30,
        duration: dur,
        ease: "power2.inOut",
      }, "toCaddy");

      // ▸ Step 2: Caddy → Go API (with TLS Shield Shimmer)
      fwd.addLabel("toApi", ">");
      fwd.call(() => {
        sounds.playHop();
        setStatusText("Caddy terminating TLS, proxying to Go API…");
      }, [], "toApi");

      // Caddy Shield Personality Animation: shimmer ring
      if (caddyShield.current) {
        fwd.fromTo(caddyShield.current,
          { attr: { r: NODE_RADIUS + 2 }, opacity: 0.7, stroke: "var(--color-accent)" },
          { attr: { r: NODE_RADIUS + 12 }, opacity: 0, duration: dur * 0.8, ease: "power2.out" },
          "toApi"
        );
      }

      pulseNode(fwd, "caddy", "var(--color-accent)", "toApi");
      fadeConn(fwd, "client", "caddy", "toApi");
      flowConn(fwd, "caddy", "api", "toApi", dur);
      fireParticles(fwd, "caddy", "api", "toApi", dur);
      moveSpotlight(fwd, "api", "toApi", dur);

      fwd.to(packetGroup.current, {
        x: pos.api.x, y: pos.api.y - 30,
        duration: dur,
        ease: "power2.inOut",
      }, "toApi");

      // ▸ Arrive at Go API (Sparks Personality Animation)
      fwd.addLabel("atApi", ">");
      fwd.call(() => {
        sounds.playHop();
        setStatusText("Go API routing & validating URL…");
      }, [], "atApi");

      // Go API Spark burst
      apiSparks.current.forEach((spark, idx) => {
        if (!spark) return;
        const angle = (idx * (Math.PI * 2)) / 3;
        const targetX = pos.api.x + Math.cos(angle) * 18;
        const targetY = pos.api.y + Math.sin(angle) * 18;
        fwd.fromTo(spark,
          { attr: { cx: pos.api.x, cy: pos.api.y }, opacity: 0.9 },
          { attr: { cx: targetX, cy: targetY }, opacity: 0, duration: 0.35, ease: "power2.out" },
          "atApi"
        );
      });

      pulseNode(fwd, "api", "var(--color-accent)", "atApi");
      fadeConn(fwd, "caddy", "api", "atApi");

      await new Promise<void>((resolve) => {
        fwd.eventCallback("onComplete", resolve);
      });

      /* ──────────────────────────────────────────────────────────────────
         SYNC POINT: Await real API response
         ────────────────────────────────────────────────────────────────── */

      const apiRing = rings.current["api"];
      if (apiRing) {
        processingTween.current = gsap.fromTo(apiRing,
          { attr: { r: NODE_RADIUS + 2 }, opacity: 0.35, stroke: "var(--color-accent)" },
          { attr: { r: NODE_RADIUS + 15 }, opacity: 0, duration: 0.7, repeat: -1, ease: "power2.out" }
        );
      }

      const result = await options.resultPromise;
      lastResultRef.current = result;

      processingTween.current?.kill();
      processingTween.current = null;
      if (apiRing) gsap.set(apiRing, { opacity: 0 });

      /* ──────────────────────────────────────────────────────────────────
         PHASE 2: RESPONSE TIMELINE
         ────────────────────────────────────────────────────────────────── */

      const res = gsap.timeline();
      tlRef.current = res;

      if (result.isError) {
        // ── Error Path ──
        setSpotlightColor("var(--color-error)");
        morphPacketColor(res, "var(--color-error)", "0");
        crossfadeLabel(res, "ERROR", "0");
        res.call(() => setStatusText(result.errorMessage || "Request failed"));

        flowConn(res, "api", "caddy", ">", dur * 0.8);
        moveSpotlight(res, "caddy", "<", dur * 0.8);
        res.to(packetGroup.current, { x: pos.caddy.x, y: pos.caddy.y - 30, duration: dur * 0.8, ease: "power2.inOut" }, "<");
        fadeConn(res, "api", "caddy", ">");

        flowConn(res, "caddy", "client", ">", dur * 0.8);
        moveSpotlight(res, "client", "<", dur * 0.8);
        res.to(packetGroup.current, { x: pos.client.x, y: pos.client.y - 30, duration: dur * 0.8, ease: "power2.inOut" }, "<");
        fadeConn(res, "caddy", "client", ">");

        pulseNode(res, "client", "var(--color-error)", ">");
        res.to(packetGroup.current, { scale: 0.6, opacity: 0, duration: 0.3, ease: "power2.in" });

      } else {
        // ── Success Path ──
        const isHit = result.cacheOutcome === "hit";
        const returnDur = dur * 0.7;

        // ▸ Step 3: API → Redis
        res.call(() => setStatusText("Querying Redis in-memory cache…"));
        flowConn(res, "api", "redis", ">", dur);
        fireParticles(res, "api", "redis", "<", dur);
        moveSpotlight(res, "redis", "<", dur);

        res.to(packetGroup.current, {
          x: pos.redis.x, y: pos.redis.y - 30,
          duration: dur, ease: "power2.inOut",
        }, "<");

        // ▸ Redis arrival + cache outcome
        res.addLabel("redis", ">");
        res.call(() => {
          if (isHit) {
            sounds.playCacheHit();
            setSpotlightColor("var(--color-success)");
          } else {
            sounds.playHop();
            setSpotlightColor("var(--color-warning)");
          }
        }, [], "redis");

        pulseNode(res, "redis", isHit ? "var(--color-success)" : "var(--color-warning)", "redis");
        fadeConn(res, "api", "redis", "redis");

        // Cache badge pops in
        res.call(() => {
          if (cacheBadgeRect.current) {
            cacheBadgeRect.current.setAttribute(
              "fill",
              isHit ? "var(--color-success)" : "var(--color-warning)"
            );
          }
          if (cacheBadgeText.current) {
            cacheBadgeText.current.textContent = isHit ? "HIT" : "MISS";
          }
        }, [], "redis");

        res.fromTo(cacheBadge.current,
          { scale: 0, opacity: 0 },
          { scale: 1, opacity: 1, duration: 0.25, ease: "back.out(2.5)", transformOrigin: "center center" },
          "redis+=0.05"
        );

        crossfadeLabel(res, isHit ? "HIT ✓" : "MISS", "redis+=0.05");

        res.call(() => setStatusText(
          isHit
            ? `Redis HIT — returning cached ${result.shortCode || "code"} [Fast Path]`
            : "Redis MISS — querying PostgreSQL durable store…"
        ), [], "redis+=0.1");

        // ▸ Step 4: On MISS → PostgreSQL round-trip
        if (!isHit) {
          res.addLabel("toPostgres", "redis+=0.35");

          flowConn(res, "redis", "postgres", "toPostgres", dur);
          fireParticles(res, "redis", "postgres", "toPostgres", dur);
          moveSpotlight(res, "postgres", "toPostgres", dur);

          res.to(packetGroup.current, {
            x: pos.postgres.x, y: pos.postgres.y - 30,
            duration: dur, ease: "power2.inOut",
          }, "toPostgres");

          res.addLabel("atPostgres", ">");

          // PostgreSQL Personality: Mechanical Stamp & Sound
          res.call(() => {
            sounds.playDbWrite();
            setStatusText(`SQL INSERT — persisting ${result.shortCode || "entry"}…`);
          }, [], "atPostgres");

          if (postgresIconGroup.current) {
            res.fromTo(postgresIconGroup.current,
              { scaleY: 1 },
              { scaleY: 0.72, duration: 0.12, yoyo: true, repeat: 1, transformOrigin: "center bottom" },
              "atPostgres"
            );
          }

          pulseNode(res, "postgres", "var(--color-accent)", "atPostgres");
          fadeConn(res, "redis", "postgres", "atPostgres");
          crossfadeLabel(res, "SAVED", "atPostgres");

          res.to({}, { duration: dur * 0.35 });

          // Return to Redis for cache pre-warm
          res.call(() => setStatusText("Pre-warming Redis cache with 24h TTL…"));
          flowConn(res, "postgres", "redis", ">", returnDur);
          fireParticles(res, "postgres", "redis", "<", returnDur);
          moveSpotlight(res, "redis", "<", returnDur);

          res.to(packetGroup.current, {
            x: pos.redis.x, y: pos.redis.y - 30,
            duration: returnDur, ease: "power2.inOut",
          }, "<");

          res.addLabel("backRedis", ">");
          pulseNode(res, "redis", "var(--color-success)", "backRedis");
          fadeConn(res, "postgres", "redis", "backRedis");
        }

        // ▸ Step 5: Return journey (Redis → API → Caddy → Client)
        res.addLabel("return", ">");
        setSpotlightColor("var(--color-success)");
        morphPacketColor(res, "var(--color-success)", "return");
        crossfadeLabel(res, result.shortCode || "OK", "return");
        res.call(() => setStatusText("Packaging HTTP JSON response…"), [], "return");

        // Redis → API
        flowConn(res, "redis", "api", "return+=0.15", returnDur);
        fireParticles(res, "redis", "api", "return+=0.15", returnDur);
        moveSpotlight(res, "api", "return+=0.15", returnDur);
        res.to(packetGroup.current, { x: pos.api.x, y: pos.api.y - 30, duration: returnDur, ease: "power2.inOut" }, "return+=0.15");

        // API → Caddy
        res.addLabel("retCaddy", ">");
        res.call(() => {
          sounds.playHop();
          setStatusText("Caddy proxying HTTP 200/201 response…");
        }, [], "retCaddy");
        fadeConn(res, "redis", "api", "retCaddy");
        flowConn(res, "api", "caddy", "retCaddy", returnDur);
        fireParticles(res, "api", "caddy", "retCaddy", returnDur);
        moveSpotlight(res, "caddy", "retCaddy", returnDur);
        res.to(packetGroup.current, { x: pos.caddy.x, y: pos.caddy.y - 30, duration: returnDur, ease: "power2.inOut" }, "retCaddy");

        // Caddy → Client
        res.addLabel("retClient", ">");
        res.call(() => {
          sounds.playHop();
          setStatusText("Streaming response payload to client browser…");
        }, [], "retClient");
        fadeConn(res, "api", "caddy", "retClient");
        flowConn(res, "caddy", "client", "retClient", returnDur);
        fireParticles(res, "caddy", "client", "retClient", returnDur);
        moveSpotlight(res, "client", "retClient", returnDur);
        res.to(packetGroup.current, { x: pos.client.x, y: pos.client.y - 30, duration: returnDur, ease: "power2.inOut" }, "retClient");

        // ▸ Step 6: Completion celebration (sound: Complete Chord)
        res.addLabel("done", ">");
        res.call(() => sounds.playComplete(), [], "done");
        fadeConn(res, "caddy", "client", "done");

        // All node rings burst outward simultaneously
        for (const node of NODES) {
          const r = rings.current[node.id];
          if (r) {
            res.fromTo(r,
              { attr: { r: NODE_RADIUS }, opacity: 0.45, stroke: "var(--color-success)", "stroke-width": 1.5 },
              { attr: { r: NODE_RADIUS + 24 }, opacity: 0, duration: 0.7, ease: "power2.out" },
              "done"
            );
          }
          const c = circles.current[node.id];
          if (c) {
            res.set(c, { attr: { stroke: "var(--color-success)", fill: "var(--color-node-bg)", "stroke-width": 2 } }, "done+=0.05");
          }
        }

        // Dissolve packet upward
        res.to(packetGroup.current, {
          scale: 1.4, opacity: 0, y: pos.client.y - 55,
          duration: 0.4, ease: "power2.in",
        }, "done+=0.05");

        // Final status
        const latStr = result.latencyMs ? ` in ${result.latencyMs}ms` : "";
        const cacheLabel = isHit ? " [Redis Cache HIT]" : " [Database Insert]";
        res.call(() => {
          setStatusText(`Complete${latStr}${cacheLabel} — ${result.shortCode || "ready"}`);
          setHasCompleted(true);
        }, [], "done+=0.1");
      }

      await new Promise<void>((resolve) => {
        res.eventCallback("onComplete", resolve);
      });
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps -- helpers close over pos & stable refs
    [killAll, pos]
  );

  /** Replay the last run pipeline */
  const replay = useCallback(() => {
    if (!lastOptionsRef.current || !lastResultRef.current) return;
    const opts: PipelineOptions = {
      url: lastOptionsRef.current.url,
      resultPromise: Promise.resolve(lastResultRef.current),
    };
    runPipeline(opts);
  }, [runPipeline]);

  useImperativeHandle(ref, () => ({ runPipeline, reset: resetAll, replay }));

  /* ═══════════════════════════════════════════════════════════════════════
     SVG Render
     ═══════════════════════════════════════════════════════════════════════ */

  return (
    <div className={`w-full max-w-3xl mx-auto ${visible ? "" : "h-0 overflow-hidden"}`}>
      <div className="rounded-xl border border-[var(--color-border-subtle)] bg-[var(--color-canvas-bg)] p-3 sm:p-5 relative overflow-hidden shadow-xl shadow-black/5">
        <svg
          ref={svgRef}
          viewBox={viewBox}
          className="w-full h-auto"
          style={{ opacity: 0 }}
          role="img"
          aria-label="Architecture visualization showing request flow"
        >
          <defs>
            {/* Subtle background dot grid */}
            <pattern id="dotgrid" width="20" height="20" patternUnits="userSpaceOnUse">
              <circle cx="10" cy="10" r="0.5" fill="var(--color-text-tertiary)" opacity="0.12" />
            </pattern>

            {/* Ambient Dynamic Spotlight Radial Gradient */}
            <radialGradient id="spotlightGrad" cx="50%" cy="50%" r="50%">
              <stop ref={spotlightStop0} offset="0%" stopColor="var(--color-accent)" stopOpacity="0.4" />
              <stop ref={spotlightStop1} offset="60%" stopColor="var(--color-accent)" stopOpacity="0.08" />
              <stop offset="100%" stopColor="var(--color-accent)" stopOpacity="0" />
            </radialGradient>
          </defs>

          {/* Background grid */}
          <rect width="100%" height="100%" fill="url(#dotgrid)" />

          {/* Dynamic Ambient Spotlight Glow (tracks underneath the active packet) */}
          <circle
            ref={spotlightRef}
            cx={pos.client.x}
            cy={pos.client.y}
            r={65}
            fill="url(#spotlightGrad)"
            opacity={0}
            pointerEvents="none"
          />

          {/* ── Connection Lines ──────────────────────────────────────── */}
          {CONNECTIONS.map((conn) => {
            const fp = pos[conn.from];
            const tp = pos[conn.to];
            const key = `${conn.from}-${conn.to}`;
            const mx = (fp.x + tp.x) / 2;
            const my = (fp.y + tp.y) / 2 - 8;

            return (
              <g key={key}>
                {/* Base line */}
                <line
                  x1={fp.x} y1={fp.y} x2={tp.x} y2={tp.y}
                  stroke="var(--color-connection)" strokeWidth={1} opacity={0.35}
                />
                {/* Flow overlay line (animated by GSAP) */}
                <line
                  ref={(el) => { if (el) flows.current[key] = el; }}
                  x1={fp.x} y1={fp.y} x2={tp.x} y2={tp.y}
                  stroke="var(--color-connection-active)" strokeWidth={2.5}
                  opacity={0} strokeLinecap="round"
                />
                {/* Connection label */}
                <text
                  x={mx} y={my}
                  textAnchor="middle" className="text-[7px] font-mono"
                  fill="var(--color-text-tertiary)" opacity={0.4}
                >
                  {conn.label}
                </text>
                {/* Particles */}
                {[0, 1, 2].map((i) => (
                  <circle
                    key={i}
                    ref={(el) => {
                      if (!particles.current[key]) particles.current[key] = [];
                      particles.current[key]![i] = el;
                    }}
                    cx={fp.x} cy={fp.y} r={2.5}
                    fill="var(--color-accent)" opacity={0}
                  />
                ))}
              </g>
            );
          })}

          {/* ── Architecture Nodes ────────────────────────────────────── */}
          {NODES.map((node) => {
            const p = pos[node.id];
            const icon = NODE_ICON_PATHS[node.id];

            return (
              <g
                key={node.id}
                className="cursor-pointer"
                onMouseEnter={() => setHoveredNode(node)}
                onMouseLeave={() => setHoveredNode(null)}
              >
                {/* Ripple ring */}
                <circle
                  ref={(el) => { if (el) rings.current[node.id] = el; }}
                  cx={p.x} cy={p.y} r={NODE_RADIUS}
                  fill="none" stroke="var(--color-accent)" strokeWidth={1.5}
                  opacity={0}
                />

                {/* Caddy TLS Shield Shimmer Ring */}
                {node.id === "caddy" && (
                  <circle
                    ref={caddyShield}
                    cx={p.x} cy={p.y} r={NODE_RADIUS + 2}
                    fill="none" stroke="var(--color-accent)" strokeWidth={1.5}
                    opacity={0}
                  />
                )}

                {/* Main node circle */}
                <circle
                  ref={(el) => { if (el) circles.current[node.id] = el; }}
                  cx={p.x} cy={p.y} r={NODE_RADIUS}
                  fill="var(--color-node-bg)" stroke="var(--color-node-border)"
                  strokeWidth={1.5}
                  className="transition-colors hover:stroke-[var(--color-accent)]"
                />

                {/* Node icon with personality ref hooks */}
                <g
                  ref={node.id === "postgres" ? postgresIconGroup : undefined}
                  transform={`translate(${p.x}, ${p.y})`}
                >
                  <path
                    d={icon.d}
                    fill={icon.filled ? "var(--color-text-secondary)" : "none"}
                    stroke={icon.filled ? "none" : "var(--color-text-secondary)"}
                    strokeWidth={1.1} strokeLinecap="round" strokeLinejoin="round"
                  />
                </g>

                {/* Go API spark particles */}
                {node.id === "api" &&
                  [0, 1, 2].map((i) => (
                    <circle
                      key={i}
                      ref={(el) => { if (el) apiSparks.current[i] = el; }}
                      cx={p.x} cy={p.y} r={1.5}
                      fill="var(--color-accent)" opacity={0}
                    />
                  ))}

                {/* Label */}
                <text
                  x={p.x} y={p.y + NODE_RADIUS + 14}
                  textAnchor="middle" className="text-[10px] font-semibold"
                  fill="var(--color-text-secondary)"
                >
                  {node.label}
                </text>
                {/* Sublabel */}
                <text
                  x={p.x} y={p.y + NODE_RADIUS + 25}
                  textAnchor="middle" className="text-[8px] font-mono"
                  fill="var(--color-text-tertiary)" opacity={0.7}
                >
                  {node.sublabel}
                </text>
              </g>
            );
          })}

          {/* ── Cache Badge (on Redis node) ───────────────────────────── */}
          <g
            ref={cacheBadge}
            transform={`translate(${pos.redis.x + NODE_RADIUS + 8}, ${pos.redis.y})`}
            style={{ opacity: 0 }}
          >
            <rect
              ref={cacheBadgeRect}
              x={-16} y={-9} width={32} height={18} rx={4}
              fill="var(--color-success)"
            />
            <text
              ref={cacheBadgeText}
              x={0} y={1} textAnchor="middle" dominantBaseline="central"
              className="text-[7px] font-mono font-bold"
              fill="var(--color-bg)"
            >
              HIT
            </text>
          </g>

          {/* ── Request Packet ────────────────────────────────────────── */}
          <g ref={packetGroup}>
            <rect
              ref={packetBg}
              x={-68} y={-13} width={136} height={26} rx={8}
              fill="var(--color-packet-bg)" opacity={0.95}
            />
            <text
              ref={packetText}
              x={0} y={1} textAnchor="middle" dominantBaseline="central"
              className="text-[8px] font-mono font-bold"
              fill="var(--color-packet-text)"
            >
              {packetLabel}
            </text>
          </g>
        </svg>

        {/* Live Status Text */}
        <p className="mt-2 text-center text-[11px] font-mono text-[var(--color-text-tertiary)] min-h-[1.2em]">
          {statusText}
        </p>

        {/* Floating Node Inspector Card (on hover) */}
        {hoveredNode && (
          <div className="absolute top-3 right-3 sm:top-4 sm:right-4 z-20 max-w-[220px] rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-elevated)]/95 backdrop-blur-md p-2.5 shadow-lg text-left animate-in fade-in zoom-in-95 duration-150">
            <div className="flex items-center justify-between gap-1.5 mb-1">
              <span className="text-xs font-semibold text-[var(--color-text)]">
                {hoveredNode.label}
              </span>
              <span className="text-[9px] font-mono text-[var(--color-accent)] uppercase">
                {hoveredNode.detail.role}
              </span>
            </div>
            <div className="text-[10px] font-mono text-[var(--color-text-tertiary)] mb-1">
              {hoveredNode.detail.tech} · {hoveredNode.detail.protocol}
            </div>
            <p className="text-[10px] text-[var(--color-text-secondary)] leading-relaxed">
              {hoveredNode.detail.desc}
            </p>
          </div>
        )}

        {/* Playback & Speed Controls Toolbar (Available when completed) */}
        {hasCompleted && (
          <div className="mt-3 pt-2.5 border-t border-[var(--color-border-subtle)] flex items-center justify-between gap-2 text-xs font-mono">
            {/* Replay Button */}
            <button
              onClick={() => {
                sounds.playClick();
                replay();
              }}
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-md border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text)] hover:border-[var(--color-accent)] transition-all cursor-pointer text-[11px]"
            >
              <svg className="w-3 h-3 text-[var(--color-accent)]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <polyline points="1 4 1 10 7 10" />
                <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10" />
              </svg>
              Replay Flow
            </button>

            {/* Speed Selector */}
            <div className="flex items-center gap-1 bg-[var(--color-bg)] p-0.5 rounded-md border border-[var(--color-border-subtle)] text-[10px]">
              <span className="text-[var(--color-text-tertiary)] px-1.5 hidden sm:inline">Speed:</span>
              {([1, 0.5, 0.25] as const).map((s) => (
                <button
                  key={s}
                  onClick={() => {
                    sounds.playClick();
                    setSpeed(s);
                  }}
                  className={`px-1.5 py-0.5 rounded cursor-pointer transition-all ${
                    speed === s
                      ? "bg-[var(--color-accent)] text-[var(--color-bg)] font-bold shadow-xs"
                      : "text-[var(--color-text-secondary)] hover:text-[var(--color-text)]"
                  }`}
                >
                  {s === 1 ? "1x" : s === 0.5 ? "0.5x" : "0.25x"}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
});

export default Visualizer;

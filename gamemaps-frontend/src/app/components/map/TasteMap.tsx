"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import styles from "./TasteMap.module.css";

export type MapNode = {
  id: string;
  label: string;
  kind: "user" | "liked" | "recommended";
  app_id?: number;
  x: number;
  y: number;
  score?: number;
  anchor?: string;
};

export type MapEdge = {
  from: string;
  to: string;
  kind: "liked" | "recommended";
};

export type MapPayload = {
  user_id: number;
  nodes: MapNode[];
  edges: MapEdge[];
};

type TasteMapProps = {
  payload: MapPayload;
  zoom?: number;
  onZoomChange?: (zoom: number) => void;
};

const MIN_ZOOM = 0.6;
const MAX_ZOOM = 2.2;

export default function TasteMap({ payload, zoom = 1, onZoomChange }: TasteMapProps) {
  const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [dragging, setDragging] = useState(false);
  const lastPoint = useRef<{ x: number; y: number } | null>(null);
  const wrapRef = useRef<HTMLDivElement | null>(null);

  const layout = useMemo(() => {
    if (!payload || payload.nodes.length === 0) {
      return {
        viewBox: "-400 -260 800 520",
        nodeById: new Map<string, MapNode>(),
        base: { minX: -400, minY: -260, width: 800, height: 520 },
      };
    }
    let minX = payload.nodes[0].x;
    let maxX = payload.nodes[0].x;
    let minY = payload.nodes[0].y;
    let maxY = payload.nodes[0].y;
    const nodeById = new Map<string, MapNode>();

    for (const node of payload.nodes) {
      nodeById.set(node.id, node);
      minX = Math.min(minX, node.x);
      maxX = Math.max(maxX, node.x);
      minY = Math.min(minY, node.y);
      maxY = Math.max(maxY, node.y);
    }

    const pad = 90;
    const base = {
      minX: minX - pad,
      minY: minY - pad,
      width: maxX - minX + pad * 2,
      height: maxY - minY + pad * 2,
    };
    return { viewBox: "", nodeById, base };
  }, [payload]);

  const viewBox = useMemo(() => {
    const clamped = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, zoom));
    const centerX = layout.base.minX + layout.base.width / 2;
    const centerY = layout.base.minY + layout.base.height / 2;
    const width = layout.base.width / clamped;
    const height = layout.base.height / clamped;
    const minX = centerX - width / 2 + pan.x;
    const minY = centerY - height / 2 + pan.y;
    return `${minX} ${minY} ${width} ${height}`;
  }, [layout.base, pan, zoom]);

  const highlighted = useMemo(() => {
    if (!hoveredNodeId || !payload) {
      return { nodes: new Set<string>(), edges: new Set<string>() };
    }
    const nodes = new Set<string>([hoveredNodeId]);
    const edges = new Set<string>();
    for (const edge of payload.edges) {
      if (edge.from === hoveredNodeId || edge.to === hoveredNodeId) {
        edges.add(`${edge.from}-${edge.to}`);
        nodes.add(edge.from);
        nodes.add(edge.to);
      }
    }
    return { nodes, edges };
  }, [hoveredNodeId, payload]);

  function applyWheel(deltaY: number) {
    if (!onZoomChange) return;
    const delta = deltaY > 0 ? -0.1 : 0.1;
    const next = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, zoom + delta));
    onZoomChange(Number(next.toFixed(2)));
  }

  useEffect(() => {
    const el = wrapRef.current;
    if (!el) return;
    const handler = (event: WheelEvent) => {
      event.preventDefault();
      event.stopPropagation();
      applyWheel(event.deltaY);
    };
    el.addEventListener("wheel", handler, { passive: false });
    return () => el.removeEventListener("wheel", handler);
  }, [zoom, onZoomChange]);

  function handlePointerDown(event: React.PointerEvent<HTMLDivElement>) {
    if (event.button !== 0) return;
    event.preventDefault();
    setDragging(true);
    lastPoint.current = { x: event.clientX, y: event.clientY };
    (event.currentTarget as HTMLDivElement).setPointerCapture(event.pointerId);
  }

  function handlePointerMove(event: React.PointerEvent<HTMLDivElement>) {
    if (!dragging || !lastPoint.current || !wrapRef.current) return;
    const rect = wrapRef.current.getBoundingClientRect();
    const dx = event.clientX - lastPoint.current.x;
    const dy = event.clientY - lastPoint.current.y;
    lastPoint.current = { x: event.clientX, y: event.clientY };

    const clamped = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, zoom));
    const viewWidth = layout.base.width / clamped;
    const viewHeight = layout.base.height / clamped;
    const scaleX = viewWidth / rect.width;
    const scaleY = viewHeight / rect.height;

    setPan((prev) => ({
      x: prev.x - dx * scaleX,
      y: prev.y - dy * scaleY,
    }));
  }

  function handlePointerUp(event: React.PointerEvent<HTMLDivElement>) {
    setDragging(false);
    lastPoint.current = null;
    (event.currentTarget as HTMLDivElement).releasePointerCapture(event.pointerId);
  }

  return (
    <div
      ref={wrapRef}
      className={`${styles.mapWrap}${dragging ? ` ${styles.dragging}` : ""}`}
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerUp}
      onPointerLeave={() => {
        setDragging(false);
        lastPoint.current = null;
      }}
    >
      <svg className={styles.mapSvg} viewBox={viewBox} role="img">
        <g className={styles.mapGuides}>
          <circle cx="0" cy="0" r="160" />
          <circle cx="0" cy="0" r="280" />
          <circle cx="0" cy="0" r="360" />
        </g>
        <g className={styles.mapEdges}>
          {payload.edges.map((edge, idx) => {
            const from = layout.nodeById.get(edge.from);
            const to = layout.nodeById.get(edge.to);
            if (!from || !to) return null;
            const key = `${edge.from}-${edge.to}`;
            const isHot = highlighted.edges.size > 0 && highlighted.edges.has(key);
            return (
              <line
                key={`${edge.from}-${edge.to}-${idx}`}
                x1={from.x}
                y1={from.y}
                x2={to.x}
                y2={to.y}
                className={
                  edge.kind === "liked"
                    ? `${styles.edgeLiked}${isHot ? ` ${styles.edgeHot}` : ""}`
                    : `${styles.edgeRec}${isHot ? ` ${styles.edgeHot}` : ""}`
                }
              />
            );
          })}
        </g>
        <g className={styles.mapNodes}>
          {payload.nodes.map((node, idx) => {
            const r = node.kind === "user" ? 18 : node.kind === "liked" ? 12 : 9;
            const isHot = highlighted.nodes.size > 0 && highlighted.nodes.has(node.id);
            return (
              <g
                key={node.id}
                className={`${styles.node}${isHot ? ` ${styles.nodeHot}` : ""}`}
                style={{ animationDelay: `${idx * 40}ms` }}
                onMouseEnter={() => setHoveredNodeId(node.id)}
                onMouseLeave={() => setHoveredNodeId(null)}
              >
                <circle
                  cx={node.x}
                  cy={node.y}
                  r={r}
                  className={
                    node.kind === "user"
                      ? styles.nodeUser
                      : node.kind === "liked"
                      ? styles.nodeLiked
                      : styles.nodeRec
                  }
                />
                <text x={node.x} y={node.y - r - 6} className={styles.nodeLabel}>
                  {node.label}
                </text>
              </g>
            );
          })}
        </g>
      </svg>
    </div>
  );
}

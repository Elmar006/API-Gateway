/**
 * Client-side view types derived from the real backend API.
 *
 * The backend only exposes `Route` (8 fields) — every other concept here is
 * either:
 *   - a 1:1 alias of a backend type (`Route`, `RequestLog`, …), or
 *   - an explicitly `derived` view-model produced by the client from those
 *     real values (e.g. a "service" is the host extracted from `target_url`).
 *
 * No fictional data lives in this file or in the store anymore.
 */

import type { HTTPMethod, Route } from "@/api";

export type { HTTPMethod, Route, RouteInput, RequestLog, MetricsSummary, MetricsPeriod, RouteStat, LogFilters, LogsPage } from "@/api";

/** Shared status palette used by node + inspector visualisation. */
export type Status = "active" | "degraded" | "inactive";

/** Service node in the graph — derived from unique target hosts. */
export interface DerivedService {
  /** stable id, e.g. `svc:users.svc.cluster.local:8080`. */
  id: string;
  /** hostname (or host:port) extracted from a route's target_url. */
  host: string;
  /** scheme (http / https / ws / wss) from the same target_url. */
  scheme: string;
  /** number of routes pointing at this service. */
  routeCount: number;
  /** active route count (used for status: active>0 → active, else inactive). */
  activeRouteCount: number;
  /** ids of routes pointing at this service. */
  routeIds: number[];
}

/** Single, statically-known gateway node — the backend itself. */
export interface GatewayNodeInfo {
  id: "gateway";
  /** human label shown on the node. */
  label: string;
  /** API base URL the client is connected to (for badges/tooltips). */
  endpoint: string;
  /** ready/degraded/down derived from /ready polling. */
  status: Status;
}

export type GraphNodeKind = "gateway" | "route" | "service";

export interface InspectorTarget {
  id: string;
  kind: GraphNodeKind;
}

/** Discriminated union of inspector payloads. */
export type InspectorData =
  | { kind: "gateway"; node: GatewayNodeInfo }
  | { kind: "route"; route: Route }
  | { kind: "service"; service: DerivedService };

export const HTTP_METHODS_DISPLAY: readonly HTTPMethod[] = [
  "ALL",
  "GET",
  "POST",
  "PUT",
  "PATCH",
  "DELETE",
  "HEAD",
  "OPTIONS",
] as const;

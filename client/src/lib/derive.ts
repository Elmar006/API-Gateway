/**
 * Pure helpers that derive view-models from the real backend `Route[]`.
 *
 * The backend has no Service entity — we synthesise a "service" per unique
 * target host. This is honest (every field is computed from real data) and
 * also useful in the graph (one service node per upstream).
 */

import { API_BASE_URL } from "@/api";
import type { DerivedService, GatewayNodeInfo, Route, Status } from "@/types/domain";

export function parseTargetURL(target: string): { scheme: string; host: string } | null {
  try {
    const parsed = new URL(target);
    return { scheme: parsed.protocol.replace(":", ""), host: parsed.host };
  } catch {
    return null;
  }
}

export function serviceIdFor(target: string): string | null {
  const parsed = parseTargetURL(target);
  if (!parsed) return null;
  return `svc:${parsed.scheme}://${parsed.host}`;
}

export function deriveServices(routes: readonly Route[]): DerivedService[] {
  const map = new Map<string, DerivedService>();
  for (const route of routes) {
    const parsed = parseTargetURL(route.target_url);
    if (!parsed) continue;
    const id = `svc:${parsed.scheme}://${parsed.host}`;
    const existing = map.get(id);
    if (existing) {
      existing.routeCount += 1;
      if (route.is_active) existing.activeRouteCount += 1;
      existing.routeIds.push(route.id);
    } else {
      map.set(id, {
        id,
        host: parsed.host,
        scheme: parsed.scheme,
        routeCount: 1,
        activeRouteCount: route.is_active ? 1 : 0,
        routeIds: [route.id],
      });
    }
  }
  return Array.from(map.values()).sort((a, b) => a.host.localeCompare(b.host));
}

export function statusForRoute(route: Route): Status {
  return route.is_active ? "active" : "inactive";
}

export function statusForService(service: DerivedService): Status {
  if (service.activeRouteCount === 0) return "inactive";
  if (service.activeRouteCount < service.routeCount) return "degraded";
  return "active";
}

export function defaultGatewayInfo(status: Status = "active"): GatewayNodeInfo {
  // The admin endpoint we're talking to is the most informative label.
  return {
    id: "gateway",
    label: "API Gateway",
    endpoint: API_BASE_URL,
    status,
  };
}

export function methodTone(method: string): "accent" | "muted" {
  return method === "ALL" || method === "GET" ? "muted" : "accent";
}

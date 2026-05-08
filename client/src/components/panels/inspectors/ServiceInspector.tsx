import { Boxes } from "lucide-react";
import { useMemo } from "react";

import { Badge } from "@/components/ui/Badge";
import { Card, CardSection, CardTitle } from "@/components/ui/Card";
import { statusForService } from "@/lib/derive";
import type { DerivedService, Route } from "@/types/domain";

export function ServiceInspector({
  service,
  routes,
}: {
  service: DerivedService;
  routes: readonly Route[];
}) {
  const serviceRoutes = useMemo(
    () => routes.filter((r) => service.routeIds.includes(r.id)),
    [routes, service.routeIds],
  );
  const status = statusForService(service);

  return (
    <div className="space-y-3">
      <Card>
        <CardSection className="space-y-3">
          <CardTitle>Upstream</CardTitle>
          <div className="grid grid-cols-2 gap-2">
            <Pair label="Host" value={service.host} mono />
            <Pair label="Scheme" value={service.scheme.toUpperCase()} />
            <Pair label="Routes" value={String(service.routeCount)} />
            <Pair label="Active" value={`${service.activeRouteCount} / ${service.routeCount}`} />
          </div>
        </CardSection>
        <CardSection className="space-y-2">
          <CardTitle>Status</CardTitle>
          <div className="flex flex-wrap items-center gap-1.5">
            <Badge tone={status === "active" ? "ok" : status === "degraded" ? "warn" : "muted"} dot>
              {status}
            </Badge>
            <span className="text-[11px] text-text-muted">
              Derived from <code className="font-mono">target_url</code> hosts of routes
              pointing here.
            </span>
          </div>
        </CardSection>
        <CardSection className="space-y-2">
          <CardTitle>Routes pointing at this service</CardTitle>
          <ul className="space-y-1.5">
            {serviceRoutes.map((r) => (
              <li
                key={r.id}
                className="flex items-center gap-2 rounded-md border border-line-subtle bg-bg-glass/40 px-2.5 py-1.5"
              >
                <span className="rounded bg-bg-chip/70 px-1.5 py-0.5 font-mono text-[10px] uppercase text-text-secondary">
                  {r.method}
                </span>
                <span className="flex-1 truncate font-mono text-xs text-text-primary">
                  {r.path_pattern}
                </span>
                <Boxes className="h-3 w-3 text-text-muted" />
                <span className="text-[10px] uppercase tracking-wider text-text-muted">
                  {r.is_active ? "active" : "off"}
                </span>
              </li>
            ))}
            {serviceRoutes.length === 0 && (
              <li className="rounded-md border border-dashed border-line-subtle px-3 py-3 text-center text-xs text-text-muted">
                No routes resolve to this upstream.
              </li>
            )}
          </ul>
        </CardSection>
      </Card>
    </div>
  );
}

function Pair({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex flex-col rounded-md border border-line-subtle bg-bg-glass/40 px-2.5 py-1.5">
      <span className="text-[10px] uppercase tracking-wider text-text-muted">{label}</span>
      <span className={`text-sm text-text-primary ${mono ? "font-mono" : ""}`}>{value}</span>
    </div>
  );
}

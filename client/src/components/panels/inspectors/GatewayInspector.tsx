import { Badge } from "@/components/ui/Badge";
import { Card, CardSection, CardTitle } from "@/components/ui/Card";
import type { GatewayNodeInfo } from "@/types/domain";

export function GatewayInspector({
  gateway,
  routeCount,
  activeRouteCount,
  serviceCount,
}: {
  gateway: GatewayNodeInfo;
  routeCount: number;
  activeRouteCount: number;
  serviceCount: number;
}) {
  return (
    <div className="space-y-3">
      <Card>
        <CardSection className="space-y-3">
          <CardTitle>Gateway</CardTitle>
          <div className="grid grid-cols-2 gap-2">
            <Pair label="Endpoint" value={gateway.endpoint} mono />
            <Pair label="Status" value={gateway.status} />
            <Pair label="Routes" value={String(routeCount)} />
            <Pair label="Active" value={`${activeRouteCount} / ${routeCount}`} />
            <Pair label="Upstreams" value={String(serviceCount)} />
          </div>
        </CardSection>
        <CardSection className="space-y-2">
          <CardTitle>About</CardTitle>
          <p className="text-xs leading-relaxed text-text-muted">
            This is the live admin endpoint of your gateway instance. Health,
            readiness, and route data come from <code className="font-mono">/admin</code>.
          </p>
          <div className="flex flex-wrap gap-1.5">
            <Badge tone="muted">prometheus: <span className="font-mono ml-1">/metrics</span></Badge>
            <Badge tone="muted">logs: <span className="font-mono ml-1">/logs</span></Badge>
            <Badge tone="muted">summary: <span className="font-mono ml-1">/metrics-summary</span></Badge>
          </div>
        </CardSection>
      </Card>
    </div>
  );
}

function Pair({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex flex-col rounded-md border border-line-subtle bg-bg-glass/40 px-2.5 py-1.5">
      <span className="text-[10px] uppercase tracking-wider text-text-muted">{label}</span>
      <span className={`truncate text-sm text-text-primary ${mono ? "font-mono" : ""}`} title={value}>{value}</span>
    </div>
  );
}

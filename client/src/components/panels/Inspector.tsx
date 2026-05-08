import { AnimatePresence, motion } from "framer-motion";
import { PanelRightClose, ScrollText, X } from "lucide-react";
import { useMemo } from "react";

import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { ScrollArea } from "@/components/ui/ScrollArea";
import { StatusDot } from "@/components/ui/StatusDot";
import { useGatewayHealth } from "@/hooks/useHealth";
import { useRoutes } from "@/hooks/useRoutes";
import { defaultGatewayInfo, deriveServices, statusForRoute, statusForService } from "@/lib/derive";
import { useAppStore } from "@/store/appStore";
import { useGraphStore } from "@/store/graphStore";
import type { Status } from "@/types/domain";

import { GatewayInspector } from "./inspectors/GatewayInspector";
import { RouteInspector } from "./inspectors/RouteInspector";
import { ServiceInspector } from "./inspectors/ServiceInspector";

export function Inspector() {
  const selection = useGraphStore((s) => s.selection);
  const select = useGraphStore((s) => s.select);
  const setOpen = useAppStore((s) => s.setInspectorOpen);

  const { data: routes = [] } = useRoutes();
  const services = useMemo(() => deriveServices(routes), [routes]);
  const gatewayHealth = useGatewayHealth();
  const gateway = useMemo(
    () => defaultGatewayInfo(gatewayHealth.ok ? "active" : "degraded"),
    [gatewayHealth.ok],
  );

  const resolved = useMemo(() => {
    if (!selection) return null;
    if (selection.kind === "gateway") {
      return { kind: "gateway" as const, gateway };
    }
    if (selection.kind === "route") {
      const numericId = Number(selection.id);
      const route = routes.find((r) => r.id === numericId);
      return route ? { kind: "route" as const, route } : null;
    }
    if (selection.kind === "service") {
      const service = services.find((s) => s.id === selection.id);
      return service ? { kind: "service" as const, service } : null;
    }
    return null;
  }, [selection, routes, services, gateway]);

  let name = "Nothing selected";
  let identifier = "—";
  let status: Status = "inactive";
  if (resolved?.kind === "gateway") {
    name = resolved.gateway.label;
    identifier = resolved.gateway.endpoint;
    status = resolved.gateway.status;
  } else if (resolved?.kind === "route") {
    name = `${resolved.route.method} ${resolved.route.path_pattern}`;
    identifier = `route #${resolved.route.id}`;
    status = statusForRoute(resolved.route);
  } else if (resolved?.kind === "service") {
    name = resolved.service.host;
    identifier = `${resolved.service.scheme}://${resolved.service.host}`;
    status = statusForService(resolved.service);
  }

  const activeRouteCount = useMemo(() => routes.filter((r) => r.is_active).length, [routes]);

  return (
    <aside className="surface-panel relative flex w-[400px] shrink-0 flex-col border-l border-line-subtle/80">
      <div className="flex items-start justify-between gap-3 px-5 pt-5">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 text-step text-text-muted">
            <ScrollText className="h-3.5 w-3.5" /> Inspector
          </div>
          <div className="mt-1 flex items-center gap-2.5">
            <h2 className="truncate text-lg font-semibold tracking-tight text-text-primary">
              {name}
            </h2>
            {resolved && <StatusDot status={status} />}
          </div>
          <div className="mt-0.5 truncate font-mono text-[11px] text-text-muted" title={identifier}>
            {identifier}
          </div>
        </div>
        <div className="flex items-center gap-1">
          {selection && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => select(null)}
              title="Clear selection"
              aria-label="Clear selection"
            >
              <X className="h-3.5 w-3.5" />
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => setOpen(false)}
            title="Hide inspector"
            aria-label="Hide inspector"
          >
            <PanelRightClose className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      {resolved?.kind === "route" && (
        <div className="mt-2 flex flex-wrap items-center gap-1.5 px-5">
          <Badge tone="accent" dot>Route</Badge>
          <Badge tone="muted">priority {resolved.route.priority}</Badge>
          {resolved.route.require_auth && <Badge tone="muted">auth required</Badge>}
          {resolved.route.rate_limit > 0 && (
            <Badge tone="muted">{resolved.route.rate_limit} rps</Badge>
          )}
        </div>
      )}

      <ScrollArea className="mt-4 flex-1">
        <div className="px-5 pb-6 pt-1">
          <AnimatePresence mode="wait">
            <motion.div
              key={resolved ? `${resolved.kind}:${selection?.id}` : "empty"}
              initial={{ opacity: 0, y: 4 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -4 }}
              transition={{ duration: 0.16, ease: "easeOut" }}
            >
              {!resolved && <EmptyState />}
              {resolved?.kind === "route" && <RouteInspector route={resolved.route} />}
              {resolved?.kind === "service" && (
                <ServiceInspector service={resolved.service} routes={routes} />
              )}
              {resolved?.kind === "gateway" && (
                <GatewayInspector
                  gateway={resolved.gateway}
                  routeCount={routes.length}
                  activeRouteCount={activeRouteCount}
                  serviceCount={services.length}
                />
              )}
            </motion.div>
          </AnimatePresence>
        </div>
      </ScrollArea>
    </aside>
  );
}

function EmptyState() {
  return (
    <div className="rounded-md border border-dashed border-line-subtle bg-bg-glass/40 p-6 text-center">
      <div className="text-sm text-text-secondary">Select a node to inspect details</div>
      <div className="mt-1 text-xs text-text-muted">
        Click the gateway, a route, or an upstream in the canvas
      </div>
    </div>
  );
}

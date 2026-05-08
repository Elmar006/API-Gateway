import { useQueryClient } from "@tanstack/react-query";
import { Download, PanelRightOpen, Upload } from "lucide-react";

import { Button } from "@/components/ui/Button";
import { ROUTES_QUERY_KEY, useCreateRoute, useRoutes } from "@/hooks/useRoutes";
import { exportRoutes, importRoutes } from "@/lib/exchange";
import { notify } from "@/lib/tauri";
import { useAppStore } from "@/store/appStore";
import { AddRouteButton } from "@/components/views/AddRouteModal";

const CANVAS_SECTIONS = new Set([
  "dashboard",
  "routes",
  "services",
  "gateways",
  "clusters",
  "middlewares",
]);

export function WorkspaceHeader() {
  const section = useAppStore((s) => s.section);
  const inspectorOpen = useAppStore((s) => s.inspectorOpen);
  const setInspectorOpen = useAppStore((s) => s.setInspectorOpen);
  const { data: routes = [] } = useRoutes();
  const createMutation = useCreateRoute();
  const queryClient = useQueryClient();

  const isCanvas = CANVAS_SECTIONS.has(section);

  return (
    <div className="flex items-end justify-between gap-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-text-primary">
          {sectionTitle(section)}
        </h1>
        <p className="mt-1 text-sm text-text-muted">{sectionSubtitle(section)}</p>
      </div>
      <div className="flex items-center gap-2">
        {isCanvas && (
          <>
            <Button
              variant="secondary"
              size="md"
              onClick={async () => {
                const imported = await importRoutes();
                if (!imported || imported.length === 0) return;
                let ok = 0;
                for (const input of imported) {
                  try {
                    await createMutation.mutateAsync(input);
                    ok += 1;
                  } catch (err) {
                    console.warn("import: skipping invalid route", input, err);
                  }
                }
                await queryClient.invalidateQueries({ queryKey: ROUTES_QUERY_KEY });
                await notify(
                  "Routes imported",
                  `${ok} of ${imported.length} routes imported successfully`,
                );
              }}
              disabled={createMutation.isPending}
            >
              <Upload className="h-4 w-4" /> Import
            </Button>
            <Button
              variant="secondary"
              size="md"
              onClick={async () => {
                const ok = await exportRoutes(routes);
                if (ok) await notify("Routes exported", `Saved ${routes.length} routes`);
              }}
              disabled={routes.length === 0}
            >
              <Download className="h-4 w-4" /> Export
            </Button>
            <AddRouteButton />
          </>
        )}
        {!inspectorOpen && (
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setInspectorOpen(true)}
            aria-label="Show inspector"
            title="Show inspector"
          >
            <PanelRightOpen className="h-4 w-4" />
          </Button>
        )}
      </div>
    </div>
  );
}

function sectionTitle(section: string) {
  switch (section) {
    case "routes":
      return "Routing";
    case "services":
      return "Services";
    case "gateways":
      return "Gateway";
    case "metrics":
      return "Metrics";
    case "logs":
      return "Logs";
    case "alerts":
      return "Alerts";
    case "settings":
      return "Settings";
    case "traces":
      return "Traces";
    case "clusters":
      return "Clusters";
    case "middlewares":
      return "Middlewares";
    case "environments":
      return "Environments";
    case "users":
      return "Users";
    case "tokens":
      return "API tokens";
    default:
      return "Dashboard";
  }
}

function sectionSubtitle(section: string) {
  switch (section) {
    case "routes":
      return "Live routes from /admin/routes — drag to rearrange, double-click to release pin.";
    case "services":
      return "Upstream hosts derived from each route's target URL.";
    case "gateways":
      return "Health and reachability of the connected admin endpoint.";
    case "metrics":
      return "Real-time aggregates from /admin/metrics-summary.";
    case "logs":
      return "Tail structured access logs from /admin/logs with rich filters.";
    case "alerts":
      return "Recent server errors (status ≥ 500) over the past 24 hours.";
    case "traces":
    case "clusters":
    case "middlewares":
    case "environments":
    case "users":
    case "tokens":
      return "Roadmap surface — backend integration arrives in a future milestone.";
    case "settings":
      return "Connection details and current session.";
    default:
      return "Operational view of the gateway control plane.";
  }
}

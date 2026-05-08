import {
  Bell,
  Boxes,
  Gauge,
  KeyRound,
  Layers,
  LayoutDashboard,
  Network,
  Route,
  ScrollText,
  Server,
  Settings,
  Share2,
  Users,
  Waypoints,
  type LucideIcon,
} from "lucide-react";

import { Separator } from "@/components/ui/Separator";
import { StatusDot } from "@/components/ui/StatusDot";
import { useGatewayHealth } from "@/hooks/useHealth";
import { useErrorCount } from "@/hooks/useLogs";
import { useRoutes } from "@/hooks/useRoutes";
import { cn } from "@/lib/cn";
import { useAppStore, type SectionId } from "@/store/appStore";

interface NavItem {
  id: SectionId;
  label: string;
  icon: LucideIcon;
}

interface NavGroup {
  label?: string;
  items: NavItem[];
}

const groups: NavGroup[] = [
  {
    items: [{ id: "dashboard", label: "Dashboard", icon: LayoutDashboard }],
  },
  {
    label: "Manage",
    items: [
      { id: "services", label: "Services", icon: Server },
      { id: "routes", label: "Routes", icon: Route },
      { id: "gateways", label: "Gateways", icon: Network },
      { id: "clusters", label: "Clusters", icon: Boxes },
      { id: "middlewares", label: "Middlewares", icon: Layers },
    ],
  },
  {
    label: "Observe",
    items: [
      { id: "metrics", label: "Metrics", icon: Gauge },
      { id: "logs", label: "Logs", icon: ScrollText },
      { id: "traces", label: "Traces", icon: Waypoints },
      { id: "alerts", label: "Alerts", icon: Bell },
    ],
  },
  {
    label: "Settings",
    items: [
      { id: "environments", label: "Environments", icon: Share2 },
      { id: "users", label: "Users", icon: Users },
      { id: "tokens", label: "API Tokens", icon: KeyRound },
      { id: "settings", label: "Settings", icon: Settings },
    ],
  },
];

export function Sidebar() {
  const section = useAppStore((s) => s.section);
  const setSection = useAppStore((s) => s.setSection);

  const gatewayHealth = useGatewayHealth();
  const { data: errorCount = 0 } = useErrorCount(60, 30_000);
  const { data: routes = [] } = useRoutes();
  const activeCount = routes.filter((r) => r.is_active).length;

  return (
    <aside className="surface-panel relative flex w-[240px] shrink-0 flex-col">
      <nav className="flex-1 overflow-auto px-3 py-4">
        {groups.map((group, idx) => (
          <div key={idx} className={cn(idx > 0 && "mt-6")}>
            {group.label && (
              <div className="mb-2 px-2 text-step text-text-muted/80">{group.label}</div>
            )}
            <ul className="space-y-1">
              {group.items.map((item) => {
                const active = section === item.id;
                const Icon = item.icon;
                const badge =
                  item.id === "alerts" && errorCount > 0
                    ? String(errorCount > 99 ? "99+" : errorCount)
                    : undefined;
                return (
                  <li key={item.id}>
                    <button
                      onClick={() => setSection(item.id)}
                      className={cn(
                        "group relative flex w-full items-center gap-3 rounded-md px-2.5 py-2 text-sm",
                        "transition-colors duration-150",
                        "text-text-secondary hover:bg-bg-glass/60 hover:text-text-primary",
                        active &&
                          "text-text-primary bg-bg-glass/80 border border-accent/40 shadow-[inset_0_0_0_1px_hsl(var(--accent)/0.18),0_0_24px_-6px_hsl(var(--accent)/0.4)]",
                      )}
                    >
                      {active && (
                        <span
                          aria-hidden
                          className="absolute inset-y-1 left-0 w-[2px] rounded-r-full bg-accent shadow-[0_0_12px_hsl(var(--accent)/0.85)]"
                        />
                      )}
                      <Icon
                        className={cn(
                          "h-4 w-4 shrink-0",
                          active ? "text-accent-soft" : "text-text-muted group-hover:text-text-primary",
                        )}
                      />
                      <span className="flex-1 text-left tracking-tight">{item.label}</span>
                      {badge && (
                        <span
                          className={cn(
                            "ml-auto inline-flex h-4 min-w-[16px] items-center justify-center rounded-full px-1 text-[10px] font-medium",
                            active
                              ? "bg-accent/30 text-accent-soft"
                              : "bg-status-err/20 text-status-err",
                          )}
                        >
                          {badge}
                        </span>
                      )}
                    </button>
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </nav>

      <div className="px-3 pb-3 pt-2">
        <Separator className="mb-3 opacity-70" />
        <div className="rounded-md border border-line-subtle bg-bg-glass/60 px-2.5 py-2">
          <div className="flex items-center gap-2">
            <StatusDot
              status={gatewayHealth.ok ? "active" : gatewayHealth.live ? "degraded" : "inactive"}
            />
            <div className="flex flex-1 flex-col text-left leading-tight">
              <span className="text-[10px] uppercase tracking-[0.18em] text-text-muted">
                API Gateway
              </span>
              <span className="text-sm text-text-primary">
                {gatewayHealth.ok ? "Ready" : gatewayHealth.live ? "Starting" : "Unreachable"}
              </span>
            </div>
          </div>
          <div className="mt-1 grid grid-cols-2 gap-1 text-[10px] text-text-muted">
            <span>routes</span>
            <span className="text-right text-text-primary">
              {activeCount}/{routes.length}
            </span>
            <span>5xx (1h)</span>
            <span className="text-right text-text-primary">{errorCount}</span>
          </div>
        </div>
      </div>
    </aside>
  );
}

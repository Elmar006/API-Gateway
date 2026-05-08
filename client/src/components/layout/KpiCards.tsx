import { Activity, AlertTriangle, Boxes, GitBranch } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useMemo } from "react";

import { useErrorCount } from "@/hooks/useLogs";
import { useMetricsSummary } from "@/hooks/useMetrics";
import { useRoutes } from "@/hooks/useRoutes";
import { deriveServices } from "@/lib/derive";

interface Stat {
  label: string;
  value: string;
  icon: LucideIcon;
  trend?: string;
  accent?: "default" | "ok" | "warn";
}

export function KpiCards() {
  const { data: routes = [] } = useRoutes();
  const services = useMemo(() => deriveServices(routes), [routes]);
  const period = "day" as const;
  const { data: summary } = useMetricsSummary(period);
  const { data: errorCount = 0 } = useErrorCount(60, 30_000);

  const activeRoutes = routes.filter((r) => r.is_active).length;

  // Compute uptime as 1 - 5xx/total over the summary window if available.
  const uptime = useMemo(() => {
    if (!summary || !summary.total_requests || summary.total_requests <= 0) return null;
    let errors = 0;
    for (const [code, count] of Object.entries(summary.status_counts ?? {})) {
      const num = Number(code);
      if (Number.isFinite(num) && num >= 500) errors += count;
    }
    const success = summary.total_requests - errors;
    if (success < 0) return null;
    return (success / summary.total_requests) * 100;
  }, [summary]);

  const stats: Stat[] = [
    {
      label: "Services",
      value: String(services.length),
      icon: Boxes,
      trend: "Derived from upstream hosts",
    },
    {
      label: "Routes",
      value: String(routes.length),
      icon: GitBranch,
      trend: `${activeRoutes} active`,
    },
    {
      label: "Errors (1h)",
      value: String(errorCount),
      icon: AlertTriangle,
      trend: errorCount > 0 ? "5xx responses" : "all clear",
      accent: errorCount > 0 ? "warn" : "ok",
    },
    {
      label: "Uptime",
      value: uptime === null ? "—" : `${uptime.toFixed(2)}%`,
      icon: Activity,
      trend: summary ? `Last ${period}` : "Awaiting traffic",
      accent: uptime !== null && uptime >= 99 ? "ok" : "default",
    },
  ];

  return (
    <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
      {stats.map((stat) => (
        <div
          key={stat.label}
          className="surface-glass group relative flex items-center gap-3 rounded-lg p-4"
        >
          <div className="grid h-10 w-10 place-items-center rounded-md border border-line-subtle bg-bg-chip/60 text-accent-soft transition-transform duration-200 group-hover:-translate-y-0.5">
            <stat.icon className="h-4 w-4" />
          </div>
          <div className="flex flex-1 flex-col">
            <div className="text-2xl font-semibold leading-none tracking-tight">
              <span
                className={
                  stat.accent === "ok"
                    ? "text-status-ok"
                    : stat.accent === "warn"
                      ? "text-status-warn"
                      : ""
                }
              >
                {stat.value}
              </span>
            </div>
            <div className="mt-1 text-xs text-text-muted">{stat.label}</div>
          </div>
          {stat.trend && (
            <div className="hidden text-right text-[11px] text-text-muted lg:block">
              {stat.trend}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

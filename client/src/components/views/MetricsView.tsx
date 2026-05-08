import { Activity, Gauge, Timer, TrendingUp } from "lucide-react";
import { useMemo, useState } from "react";

import { useMetricsSummary } from "@/hooks/useMetrics";
import type { MetricsPeriod } from "@/types/domain";

const PERIODS: { id: MetricsPeriod; label: string }[] = [
  { id: "hour", label: "1h" },
  { id: "day", label: "24h" },
  { id: "week", label: "7d" },
  { id: "month", label: "30d" },
];

export function MetricsView() {
  const [period, setPeriod] = useState<MetricsPeriod>("day");
  const { data, error, isLoading } = useMetricsSummary(period);

  const grouped = useMemo(() => {
    const buckets = { ok: 0, warn: 0, err: 0, other: 0 };
    if (!data) return buckets;
    for (const [code, count] of Object.entries(data.status_counts ?? {})) {
      const num = Number(code);
      if (!Number.isFinite(num)) continue;
      if (num >= 500) buckets.err += count;
      else if (num >= 400) buckets.warn += count;
      else if (num >= 200 && num < 400) buckets.ok += count;
      else buckets.other += count;
    }
    return buckets;
  }, [data]);

  const total = data?.total_requests ?? 0;

  return (
    <div className="surface-glass flex h-full flex-col gap-4 overflow-auto rounded-xl p-4">
      <header className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Gauge className="h-4 w-4 text-accent-soft" />
          <span className="font-medium text-text-primary">Metrics summary</span>
          {isLoading && <span className="text-xs text-text-muted">loading…</span>}
        </div>
        <div className="flex items-center gap-1 rounded-md border border-line-subtle bg-bg-glass/60 p-0.5">
          {PERIODS.map((p) => (
            <button
              key={p.id}
              onClick={() => setPeriod(p.id)}
              className={`rounded px-2 py-1 text-xs ${
                period === p.id
                  ? "bg-accent/30 text-text-primary"
                  : "text-text-muted hover:text-text-primary"
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </header>

      {error && (
        <div className="rounded-md border border-status-err/40 bg-status-err/10 px-3 py-2 text-sm text-status-err">
          Failed to load metrics: {(error as Error).message}
        </div>
      )}

      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <Stat
          icon={<TrendingUp className="h-4 w-4" />}
          label="Total requests"
          value={total.toLocaleString()}
        />
        <Stat
          icon={<Activity className="h-4 w-4" />}
          label="Throughput"
          value={`${(data?.rps ?? 0).toFixed(2)} rps`}
        />
        <Stat
          icon={<Timer className="h-4 w-4" />}
          label="Avg latency"
          value={`${(data?.avg_response_time_ms ?? 0).toFixed(1)} ms`}
        />
        <Stat
          icon={<Gauge className="h-4 w-4" />}
          label="Error rate"
          value={
            total === 0
              ? "—"
              : `${(((grouped.warn + grouped.err) / total) * 100).toFixed(2)}%`
          }
          accent={grouped.err > 0 ? "warn" : "ok"}
        />
      </div>

      <div className="grid flex-1 grid-cols-1 gap-3 lg:grid-cols-2">
        <section className="surface-glass rounded-md p-3">
          <div className="mb-2 text-step text-text-muted">Status distribution</div>
          <StatusBar counts={grouped} total={total} />
          <ul className="mt-3 grid grid-cols-2 gap-1 text-xs text-text-muted">
            <li><span className="text-status-ok">●</span> 2xx/3xx — {grouped.ok.toLocaleString()}</li>
            <li><span className="text-status-warn">●</span> 4xx — {grouped.warn.toLocaleString()}</li>
            <li><span className="text-status-err">●</span> 5xx — {grouped.err.toLocaleString()}</li>
            <li><span className="text-text-muted">●</span> other — {grouped.other.toLocaleString()}</li>
          </ul>
        </section>
        <section className="surface-glass rounded-md p-3">
          <div className="mb-2 text-step text-text-muted">Slowest routes</div>
          {(data?.slowest_routes?.length ?? 0) === 0 ? (
            <div className="rounded-md border border-dashed border-line-subtle px-3 py-6 text-center text-xs text-text-muted">
              No traffic recorded for this window yet.
            </div>
          ) : (
            <ul className="space-y-1.5">
              {data?.slowest_routes.slice(0, 8).map((r, idx) => (
                <li
                  key={`${r.path}:${idx}`}
                  className="flex items-center gap-2 rounded-md border border-line-subtle bg-bg-glass/40 px-2.5 py-1.5"
                >
                  <span className="flex-1 truncate font-mono text-xs text-text-primary">{r.path}</span>
                  <span className="font-mono text-[11px] text-text-muted">
                    {r.count.toLocaleString()} req
                  </span>
                  <span className="font-mono text-[11px] text-status-warn">
                    {r.avg_time_ms.toFixed(1)} ms
                  </span>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
  );
}

function Stat({
  icon,
  label,
  value,
  accent,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  accent?: "ok" | "warn";
}) {
  return (
    <div className="surface-glass flex items-center gap-3 rounded-md p-3">
      <div className="grid h-9 w-9 place-items-center rounded-md border border-line-subtle bg-bg-chip/60 text-accent-soft">
        {icon}
      </div>
      <div className="flex flex-1 flex-col leading-tight">
        <span
          className={`text-lg font-semibold tracking-tight ${
            accent === "warn"
              ? "text-status-warn"
              : accent === "ok"
                ? "text-status-ok"
                : "text-text-primary"
          }`}
        >
          {value}
        </span>
        <span className="text-[11px] text-text-muted">{label}</span>
      </div>
    </div>
  );
}

function StatusBar({
  counts,
  total,
}: {
  counts: { ok: number; warn: number; err: number; other: number };
  total: number;
}) {
  if (total === 0) {
    return (
      <div className="h-2.5 w-full rounded-full border border-dashed border-line-subtle" />
    );
  }
  const pct = (n: number) => `${(n / total) * 100}%`;
  return (
    <div className="flex h-2.5 w-full overflow-hidden rounded-full">
      <div className="bg-status-ok" style={{ width: pct(counts.ok) }} />
      <div className="bg-status-warn" style={{ width: pct(counts.warn) }} />
      <div className="bg-status-err" style={{ width: pct(counts.err) }} />
      <div className="bg-text-muted/40" style={{ width: pct(counts.other) }} />
    </div>
  );
}

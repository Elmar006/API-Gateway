import { ExternalLink, Waypoints } from "lucide-react";
import { useMemo } from "react";

import { useLogs } from "@/hooks/useLogs";
import type { RequestLog } from "@/api";

const TRACE_VIEWER_BASE = (
  import.meta.env.VITE_TRACE_VIEWER_URL ?? ""
).toString().trim().replace(/\/+$/, "");

export function TracesView() {
  const { data, isLoading, error } = useLogs(
    { limit: 200, offset: 0 },
    { refetchInterval: 5_000 },
  );

  const traces = useMemo(() => {
    const seen = new Set<string>();
    const out: RequestLog[] = [];
    for (const log of data?.logs ?? []) {
      const id = log.trace_id ?? "";
      if (!id || seen.has(id)) continue;
      seen.add(id);
      out.push(log);
    }
    return out;
  }, [data]);

  return (
    <div className="surface-glass flex h-full flex-col rounded-xl">
      <header className="flex items-center justify-between gap-3 border-b border-line-subtle/70 px-4 py-3">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Waypoints className="h-4 w-4 text-accent-soft" />
          <span className="font-medium text-text-primary">Traces</span>
          <span className="text-xs text-text-muted">
            {traces.length} unique trace IDs from recent logs
          </span>
        </div>
        {!TRACE_VIEWER_BASE && (
          <span className="text-[11px] text-text-muted">
            Set VITE_TRACE_VIEWER_URL to enable deep links to Jaeger / Tempo.
          </span>
        )}
      </header>

      {error ? (
        <div className="grid flex-1 place-items-center text-sm text-status-err">
          Failed to load logs: {(error as Error).message}
        </div>
      ) : isLoading ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">Loading…</div>
      ) : traces.length === 0 ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">
          No traces in the recent log window. Enable OpenTelemetry on the
          backend (OTEL_TRACES_ENABLED=true).
        </div>
      ) : (
        <ul className="flex-1 divide-y divide-line-subtle/40 overflow-auto">
          {traces.map((log) => (
            <TraceRow key={`${log.trace_id}-${log.id}`} log={log} />
          ))}
        </ul>
      )}
    </div>
  );
}

function TraceRow({ log }: { log: RequestLog }) {
  const id = log.trace_id ?? "";
  const url = TRACE_VIEWER_BASE ? `${TRACE_VIEWER_BASE}/${id}` : "";
  return (
    <li className="grid grid-cols-[200px_60px_56px_1fr_auto] items-center gap-2 px-4 py-2 text-xs">
      <code className="truncate font-mono text-text-primary" title={id}>
        {id}
      </code>
      <span className="font-mono uppercase text-text-secondary">{log.method}</span>
      <span
        className={`font-mono ${
          log.status_code >= 500
            ? "text-status-err"
            : log.status_code >= 400
              ? "text-status-warn"
              : "text-status-ok"
        }`}
      >
        {log.status_code}
      </span>
      <span className="truncate font-mono text-text-muted" title={log.path}>
        {log.path}
      </span>
      {url ? (
        <a
          href={url}
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1 text-accent-soft hover:underline"
        >
          Open <ExternalLink className="h-3 w-3" />
        </a>
      ) : (
        <span className="text-text-muted">—</span>
      )}
    </li>
  );
}

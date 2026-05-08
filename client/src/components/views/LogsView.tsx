import { useVirtualizer } from "@tanstack/react-virtual";
import { Pause, Play, RotateCw, ScrollText } from "lucide-react";
import { useMemo, useRef, useState } from "react";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { useLogs } from "@/hooks/useLogs";
import type { RequestLog } from "@/types/domain";

const STATUS_OPTIONS = [
  { label: "Any status", value: "" },
  { label: "2xx", value: "200" },
  { label: "4xx", value: "400" },
  { label: "5xx", value: "500" },
];

const PAGE_SIZE = 200;

export function LogsView() {
  const [path, setPath] = useState("");
  const [status, setStatus] = useState<string>("");
  const [live, setLive] = useState(true);

  const filters = useMemo(
    () => ({
      path: path.trim() || undefined,
      status: status ? Number(status) : undefined,
      limit: PAGE_SIZE,
      offset: 0,
    }),
    [path, status],
  );

  const { data, refetch, isFetching, error } = useLogs(filters, {
    refetchInterval: live ? 3_000 : false,
  });

  const logs = data?.logs ?? [];

  const parentRef = useRef<HTMLDivElement | null>(null);
  const virtualiser = useVirtualizer({
    count: logs.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 36,
    overscan: 12,
  });

  return (
    <div className="surface-glass flex h-full flex-col rounded-xl">
      <header className="flex items-center justify-between gap-3 border-b border-line-subtle/70 px-4 py-3">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <ScrollText className="h-4 w-4 text-accent-soft" />
          <span className="font-medium text-text-primary">Request log</span>
          <span className="text-xs text-text-muted">
            {data ? `${logs.length} of ${data.total} shown` : "loading…"}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Input
            value={path}
            onChange={(e) => setPath(e.target.value)}
            placeholder="Filter by path…"
            className="h-8 w-56 font-mono text-xs"
          />
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="h-8 rounded-md border border-line-subtle bg-bg-glass/60 px-2 text-xs"
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value} className="bg-bg-panel">
                {opt.label}
              </option>
            ))}
          </select>
          <Button
            variant="ghost"
            size="icon-sm"
            title={live ? "Pause auto-refresh" : "Resume auto-refresh"}
            onClick={() => setLive((v) => !v)}
          >
            {live ? <Pause className="h-3.5 w-3.5" /> : <Play className="h-3.5 w-3.5" />}
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            title="Refresh now"
            onClick={() => refetch()}
            disabled={isFetching}
          >
            <RotateCw className={`h-3.5 w-3.5 ${isFetching ? "animate-spin" : ""}`} />
          </Button>
        </div>
      </header>

      <div className="grid grid-cols-[160px_60px_56px_1fr_88px_72px_84px] items-center gap-2 border-b border-line-subtle/60 px-4 py-2 text-[10px] uppercase tracking-[0.18em] text-text-muted">
        <span>Time</span>
        <span>Method</span>
        <span>Status</span>
        <span>Path</span>
        <span className="text-right">Latency</span>
        <span>Client</span>
        <span>Trace</span>
      </div>

      {error ? (
        <div className="grid flex-1 place-items-center text-sm text-status-err">
          Failed to load logs: {(error as Error).message}
        </div>
      ) : logs.length === 0 ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">
          No logs matched these filters.
        </div>
      ) : (
        <div ref={parentRef} className="relative flex-1 overflow-auto">
          <div style={{ height: virtualiser.getTotalSize(), position: "relative" }}>
            {virtualiser.getVirtualItems().map((virtualRow) => {
              const log = logs[virtualRow.index];
              return (
                <LogRow
                  key={log.id}
                  log={log}
                  style={{
                    position: "absolute",
                    top: 0,
                    left: 0,
                    width: "100%",
                    height: virtualRow.size,
                    transform: `translateY(${virtualRow.start}px)`,
                  }}
                />
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

const TRACE_VIEWER_BASE = (
  import.meta.env.VITE_TRACE_VIEWER_URL ?? ""
).toString().trim().replace(/\/+$/, "");

function LogRow({ log, style }: { log: RequestLog; style: React.CSSProperties }) {
  const isErr = log.status_code >= 500;
  const isWarn = log.status_code >= 400 && log.status_code < 500;
  const traceId = log.trace_id ?? "";
  const traceShort = traceId ? `${traceId.slice(0, 8)}…` : "—";
  const traceUrl = traceId && TRACE_VIEWER_BASE ? `${TRACE_VIEWER_BASE}/${traceId}` : "";
  return (
    <div
      style={style}
      className={`grid grid-cols-[160px_60px_56px_1fr_88px_72px_84px] items-center gap-2 border-b border-line-subtle/30 px-4 text-xs ${
        isErr ? "bg-status-err/5" : isWarn ? "bg-status-warn/5" : ""
      }`}
    >
      <span className="font-mono text-text-muted">
        {new Date(log.created_at).toLocaleTimeString()}
      </span>
      <span className="font-mono uppercase text-text-secondary">{log.method}</span>
      <span
        className={`font-mono ${
          isErr ? "text-status-err" : isWarn ? "text-status-warn" : "text-status-ok"
        }`}
      >
        {log.status_code}
      </span>
      <span className="truncate font-mono text-text-primary" title={log.path}>
        {log.path}
      </span>
      <span className="text-right font-mono text-text-muted">
        {log.response_time_ms.toFixed(1)} ms
      </span>
      <span className="truncate font-mono text-text-muted" title={log.client_ip}>
        {log.client_ip}
      </span>
      {traceUrl ? (
        <a
          href={traceUrl}
          target="_blank"
          rel="noreferrer"
          title={traceId}
          className="truncate font-mono text-accent-soft hover:underline"
        >
          {traceShort}
        </a>
      ) : (
        <span title={traceId} className="truncate font-mono text-text-muted">
          {traceShort}
        </span>
      )}
    </div>
  );
}

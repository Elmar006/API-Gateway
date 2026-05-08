import { AlertTriangle, RotateCw } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { Button } from "@/components/ui/Button";
import { useLogs } from "@/hooks/useLogs";

// Anchor the window to a 60-second boundary so the query key is stable
// across renders. Without this, `Date.now()` would change on every render
// and produce a different ISO string per render → new query key → fetch
// loop bounded only by network round-trip time.
const WINDOW_BUCKET_MS = 60_000;

function currentBucket(): number {
  return Math.floor(Date.now() / WINDOW_BUCKET_MS);
}

export function AlertsView() {
  const [bucket, setBucket] = useState(currentBucket);

  // Advance the bucket once per minute so the 24h window slides forward
  // without producing a new query key on every render.
  useEffect(() => {
    const id = window.setInterval(() => {
      const next = currentBucket();
      setBucket((prev) => (prev === next ? prev : next));
    }, WINDOW_BUCKET_MS);
    return () => window.clearInterval(id);
  }, []);

  const fromIso = useMemo(
    () => new Date(bucket * WINDOW_BUCKET_MS - 24 * 3600_000).toISOString(),
    [bucket],
  );

  const { data, isFetching, refetch, error } = useLogs(
    { status: 500, from: fromIso, limit: 100, offset: 0 },
    { refetchInterval: 30_000 },
  );
  const logs = data?.logs ?? [];

  return (
    <div className="surface-glass flex h-full flex-col rounded-xl">
      <header className="flex items-center justify-between gap-3 border-b border-line-subtle/70 px-4 py-3">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <AlertTriangle className="h-4 w-4 text-status-warn" />
          <span className="font-medium text-text-primary">Server errors (last 24h)</span>
          <span className="text-xs text-text-muted">
            {data ? `${logs.length} of ${data.total}` : "loading…"}
          </span>
        </div>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={() => refetch()}
          disabled={isFetching}
          title="Refresh"
        >
          <RotateCw className={`h-3.5 w-3.5 ${isFetching ? "animate-spin" : ""}`} />
        </Button>
      </header>

      {error ? (
        <div className="grid flex-1 place-items-center text-sm text-status-err">
          Failed to load alerts: {(error as Error).message}
        </div>
      ) : logs.length === 0 ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">
          No 5xx responses in the past 24 hours. The gateway is healthy.
        </div>
      ) : (
        <ul className="flex-1 overflow-auto">
          {logs.map((log) => (
            <li
              key={log.id}
              className="grid grid-cols-[160px_70px_60px_1fr_90px] items-center gap-2 border-b border-line-subtle/30 px-4 py-2 text-xs"
            >
              <span className="font-mono text-text-muted">
                {new Date(log.created_at).toLocaleString()}
              </span>
              <span className="font-mono uppercase text-text-secondary">{log.method}</span>
              <span className="font-mono text-status-err">{log.status_code}</span>
              <span className="truncate font-mono text-text-primary" title={log.path}>
                {log.path}
              </span>
              <span className="text-right font-mono text-text-muted">
                {log.response_time_ms.toFixed(1)} ms
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

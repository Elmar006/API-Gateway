import { useQuery } from "@tanstack/react-query";

import { logsApi } from "@/api";
import type { LogFilters } from "@/types/domain";

export function useLogs(filters: LogFilters, opts?: { enabled?: boolean; refetchInterval?: number | false }) {
  return useQuery({
    queryKey: ["logs", filters],
    queryFn: ({ signal }) => logsApi.list(filters, signal),
    placeholderData: (previous) => previous,
    refetchInterval: opts?.refetchInterval ?? false,
    enabled: opts?.enabled ?? true,
  });
}

/**
 * Counts how many 5xx events occurred during the last `windowMinutes` minutes.
 * Used to populate the alert badge in the title bar — a real number, not "3".
 */
export function useErrorCount(windowMinutes = 60, refetchInterval: number | false = 30_000) {
  return useQuery({
    queryKey: ["logs:error-count", windowMinutes],
    // Compute `from` inside queryFn so each refetch uses an up-to-date
    // window boundary even when the consuming component has not re-rendered.
    // Including `from` in the key would create a brand-new query on every
    // render; excluding it from the key but reading it from a stale
    // closure would silently expand the window. Computing here gives us
    // a stable key with always-fresh boundaries.
    queryFn: async ({ signal }) => {
      const from = new Date(Date.now() - windowMinutes * 60_000).toISOString();
      const page = await logsApi.list({ status: 500, from, limit: 1, offset: 0 }, signal);
      return page.total;
    },
    refetchInterval,
    refetchOnWindowFocus: false,
    staleTime: 15_000,
  });
}

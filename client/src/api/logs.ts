import { http } from "./client";
import type { LogFilters, LogsPage } from "./types";

function buildQuery(filters: LogFilters): string {
  const params = new URLSearchParams();
  if (filters.path) params.set("path", filters.path);
  if (filters.status) params.set("status", String(filters.status));
  if (filters.from) params.set("from", filters.from);
  if (filters.to) params.set("to", filters.to);
  if (filters.limit !== undefined) params.set("limit", String(filters.limit));
  if (filters.offset !== undefined) params.set("offset", String(filters.offset));
  const qs = params.toString();
  return qs ? `?${qs}` : "";
}

export const logsApi = {
  list: (filters: LogFilters = {}, signal?: AbortSignal) =>
    http.get<LogsPage>(`/logs${buildQuery(filters)}`, { signal }),
};

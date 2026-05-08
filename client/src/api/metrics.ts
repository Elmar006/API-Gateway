import { http } from "./client";
import type { MetricsPeriod, MetricsSummary } from "./types";

export const metricsApi = {
  summary: (period: MetricsPeriod = "day", signal?: AbortSignal) =>
    http.get<MetricsSummary>(`/metrics-summary?period=${period}`, { signal }),
};

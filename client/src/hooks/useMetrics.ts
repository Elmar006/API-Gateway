import { useQuery } from "@tanstack/react-query";

import { metricsApi } from "@/api";
import type { MetricsPeriod } from "@/types/domain";

export function useMetricsSummary(period: MetricsPeriod = "day", refetchInterval: number | false = 5_000) {
  return useQuery({
    queryKey: ["metrics:summary", period],
    queryFn: ({ signal }) => metricsApi.summary(period, signal),
    refetchInterval,
    refetchOnWindowFocus: false,
    staleTime: 4_000,
  });
}

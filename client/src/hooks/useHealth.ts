import { useQuery } from "@tanstack/react-query";

import { ApiError, healthApi } from "@/api";

export interface HealthState {
  ready: boolean;
  live: boolean;
  /** True only while we're confidently online; used by the gateway node colour. */
  ok: boolean;
}

export function useGatewayHealth(refetchInterval: number | false = 10_000): HealthState {
  const { data, error } = useQuery({
    queryKey: ["health:ready"],
    queryFn: async ({ signal }) => {
      try {
        await healthApi.ready(signal);
        return { ready: true, live: true } as const;
      } catch (err) {
        if (err instanceof ApiError && err.status === 503) {
          return { ready: false, live: true } as const;
        }
        throw err;
      }
    },
    refetchInterval,
    refetchOnWindowFocus: false,
    retry: 1,
  });

  if (error) return { ready: false, live: false, ok: false };
  return { ready: data?.ready ?? false, live: data?.live ?? false, ok: data?.ready === true };
}

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { routesApi } from "@/api";
import type { Route, RouteInput } from "@/types/domain";

const ROUTES_KEY = ["routes"] as const;

export function useRoutes() {
  return useQuery({
    queryKey: ROUTES_KEY,
    queryFn: ({ signal }) => routesApi.list(signal),
  });
}

export function useCreateRoute() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: RouteInput) => routesApi.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ROUTES_KEY }),
  });
}

interface UpdateArgs {
  id: number;
  input: RouteInput;
}

export function useUpdateRoute() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: UpdateArgs) => routesApi.update(id, input),
    onMutate: async ({ id, input }) => {
      await qc.cancelQueries({ queryKey: ROUTES_KEY });
      const previous = qc.getQueryData<Route[]>(ROUTES_KEY);
      if (previous) {
        qc.setQueryData<Route[]>(
          ROUTES_KEY,
          previous.map((r) =>
            r.id === id
              ? {
                  ...r,
                  ...input,
                  updated_at: new Date().toISOString(),
                }
              : r,
          ),
        );
      }
      return { previous };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.previous) qc.setQueryData(ROUTES_KEY, ctx.previous);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ROUTES_KEY }),
  });
}

export function useDeleteRoute() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => routesApi.remove(id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: ROUTES_KEY });
      const previous = qc.getQueryData<Route[]>(ROUTES_KEY);
      if (previous) qc.setQueryData<Route[]>(ROUTES_KEY, previous.filter((r) => r.id !== id));
      return { previous };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.previous) qc.setQueryData(ROUTES_KEY, ctx.previous);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ROUTES_KEY }),
  });
}

export function useToggleRoute() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => routesApi.toggle(id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: ROUTES_KEY });
      const previous = qc.getQueryData<Route[]>(ROUTES_KEY);
      if (previous) {
        qc.setQueryData<Route[]>(
          ROUTES_KEY,
          previous.map((r) => (r.id === id ? { ...r, is_active: !r.is_active } : r)),
        );
      }
      return { previous };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.previous) qc.setQueryData(ROUTES_KEY, ctx.previous);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ROUTES_KEY }),
  });
}

export const ROUTES_QUERY_KEY = ROUTES_KEY;

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { middlewaresApi } from "@/api";
import type {
  AttachMiddlewareInput,
  Middleware,
  MiddlewareInput,
} from "@/api";

const KEY = ["middlewares"] as const;

export function useMiddlewares() {
  return useQuery({
    queryKey: KEY,
    queryFn: ({ signal }) => middlewaresApi.list(signal),
  });
}

export function useRouteMiddlewares(routeId: number | null) {
  return useQuery({
    queryKey: ["routes", routeId, "middlewares"],
    queryFn: ({ signal }) => middlewaresApi.listForRoute(routeId as number, signal),
    enabled: typeof routeId === "number" && routeId > 0,
  });
}

export function useCreateMiddleware() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: MiddlewareInput) => middlewaresApi.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export function useUpdateMiddleware() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: MiddlewareInput }) =>
      middlewaresApi.update(id, input),
    onMutate: async ({ id, input }) => {
      await qc.cancelQueries({ queryKey: KEY });
      const previous = qc.getQueryData<Middleware[]>(KEY);
      if (previous) {
        qc.setQueryData<Middleware[]>(
          KEY,
          previous.map((m) =>
            m.id === id ? { ...m, ...input, updated_at: new Date().toISOString() } : m,
          ),
        );
      }
      return { previous };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.previous) qc.setQueryData(KEY, ctx.previous);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export function useDeleteMiddleware() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => middlewaresApi.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export function useAttachMiddleware() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ routeId, input }: { routeId: number; input: AttachMiddlewareInput }) =>
      middlewaresApi.attach(routeId, input),
    onSuccess: (_data, vars) =>
      qc.invalidateQueries({ queryKey: ["routes", vars.routeId, "middlewares"] }),
  });
}

export function useDetachMiddleware() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ routeId, middlewareId }: { routeId: number; middlewareId: number }) =>
      middlewaresApi.detach(routeId, middlewareId),
    onSuccess: (_data, vars) =>
      qc.invalidateQueries({ queryKey: ["routes", vars.routeId, "middlewares"] }),
  });
}

export const MIDDLEWARES_QUERY_KEY = KEY;

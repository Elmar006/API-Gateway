import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { environmentsApi } from "@/api";
import type { Environment, EnvironmentInput } from "@/api";

const KEY = ["environments"] as const;

export function useEnvironments() {
  return useQuery({
    queryKey: KEY,
    queryFn: ({ signal }) => environmentsApi.list(signal),
  });
}

export function useCreateEnvironment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: EnvironmentInput) => environmentsApi.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

interface UpdateArgs {
  id: number;
  input: EnvironmentInput;
}

export function useUpdateEnvironment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: UpdateArgs) => environmentsApi.update(id, input),
    onMutate: async ({ id, input }) => {
      await qc.cancelQueries({ queryKey: KEY });
      const previous = qc.getQueryData<Environment[]>(KEY);
      if (previous) {
        qc.setQueryData<Environment[]>(
          KEY,
          previous.map((e) =>
            e.id === id ? { ...e, ...input, updated_at: new Date().toISOString() } : e,
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

export function useDeleteEnvironment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => environmentsApi.remove(id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: KEY });
      const previous = qc.getQueryData<Environment[]>(KEY);
      if (previous)
        qc.setQueryData<Environment[]>(
          KEY,
          previous.filter((e) => e.id !== id),
        );
      return { previous };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.previous) qc.setQueryData(KEY, ctx.previous);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export const ENVIRONMENTS_QUERY_KEY = KEY;

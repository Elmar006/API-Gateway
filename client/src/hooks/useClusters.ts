import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { clustersApi } from "@/api";
import type { Cluster, ClusterInput, ClusterTargetInput } from "@/api";

const KEY = ["clusters"] as const;

export function useClusters() {
  return useQuery({
    queryKey: KEY,
    queryFn: ({ signal }) => clustersApi.list(signal),
  });
}

export function useCluster(id: number | null) {
  return useQuery({
    queryKey: [...KEY, id],
    queryFn: ({ signal }) => clustersApi.get(id as number, signal),
    enabled: typeof id === "number" && id > 0,
  });
}

export function useCreateCluster() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: ClusterInput) => clustersApi.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export function useUpdateCluster() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: ClusterInput }) =>
      clustersApi.update(id, input),
    onMutate: async ({ id, input }) => {
      await qc.cancelQueries({ queryKey: KEY });
      const previous = qc.getQueryData<Cluster[]>(KEY);
      if (previous) {
        qc.setQueryData<Cluster[]>(
          KEY,
          previous.map((c) =>
            c.id === id ? { ...c, ...input, updated_at: new Date().toISOString() } : c,
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

export function useDeleteCluster() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => clustersApi.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export function useAddClusterTarget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      clusterId,
      input,
    }: {
      clusterId: number;
      input: ClusterTargetInput;
    }) => clustersApi.addTarget(clusterId, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export function useRemoveClusterTarget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ clusterId, targetId }: { clusterId: number; targetId: number }) =>
      clustersApi.removeTarget(clusterId, targetId),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export const CLUSTERS_QUERY_KEY = KEY;

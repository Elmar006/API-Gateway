import { http } from "./client";
import type {
  Cluster,
  ClusterInput,
  ClusterTargetInput,
} from "./types";

export const clustersApi = {
  list: (signal?: AbortSignal) =>
    http.get<Cluster[]>("/clusters", { signal }),

  get: (id: number, signal?: AbortSignal) =>
    http.get<Cluster>(`/clusters/${id}`, { signal }),

  create: (input: ClusterInput, signal?: AbortSignal) =>
    http.post<{ id: number }>("/clusters", input, { signal }),

  update: (id: number, input: ClusterInput, signal?: AbortSignal) =>
    http.put<{ status: string }>(`/clusters/${id}`, input, { signal }),

  remove: (id: number, signal?: AbortSignal) =>
    http.delete<{ status: string }>(`/clusters/${id}`, { signal }),

  addTarget: (clusterId: number, target: ClusterTargetInput, signal?: AbortSignal) =>
    http.post<{ id: number }>(`/clusters/${clusterId}/targets`, target, { signal }),

  updateTarget: (
    clusterId: number,
    targetId: number,
    target: ClusterTargetInput,
    signal?: AbortSignal,
  ) =>
    http.put<{ status: string }>(
      `/clusters/${clusterId}/targets/${targetId}`,
      target,
      { signal },
    ),

  removeTarget: (clusterId: number, targetId: number, signal?: AbortSignal) =>
    http.delete<{ status: string }>(
      `/clusters/${clusterId}/targets/${targetId}`,
      { signal },
    ),
};

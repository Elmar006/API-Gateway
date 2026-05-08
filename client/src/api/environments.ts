import { http } from "./client";
import type { Environment, EnvironmentInput } from "./types";

export const environmentsApi = {
  list: (signal?: AbortSignal) =>
    http.get<Environment[]>("/environments", { signal }),

  get: (id: number, signal?: AbortSignal) =>
    http.get<Environment>(`/environments/${id}`, { signal }),

  create: (input: EnvironmentInput, signal?: AbortSignal) =>
    http.post<{ id: number }>("/environments", input, { signal }),

  update: (id: number, input: EnvironmentInput, signal?: AbortSignal) =>
    http.put<{ status: string }>(`/environments/${id}`, input, { signal }),

  remove: (id: number, signal?: AbortSignal) =>
    http.delete<{ status: string }>(`/environments/${id}`, { signal }),
};

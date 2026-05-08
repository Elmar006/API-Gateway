import { http } from "./client";
import type { Route, RouteInput } from "./types";

export const routesApi = {
  list: (signal?: AbortSignal) => http.get<Route[]>("/routes", { signal }),

  create: (input: RouteInput, signal?: AbortSignal) =>
    http.post<{ id: number }>("/routes", input, { signal }),

  update: (id: number, input: RouteInput, signal?: AbortSignal) =>
    http.put<{ status: string }>(`/routes/${id}`, input, { signal }),

  remove: (id: number, signal?: AbortSignal) =>
    http.delete<{ status: string }>(`/routes/${id}`, { signal }),

  toggle: (id: number, signal?: AbortSignal) =>
    http.patch<{ status: string }>(`/routes/${id}/toggle`, undefined, { signal }),
};

import { http } from "./client";
import type {
  AttachMiddlewareInput,
  Middleware,
  MiddlewareInput,
} from "./types";

export const middlewaresApi = {
  list: (signal?: AbortSignal) =>
    http.get<Middleware[]>("/middlewares", { signal }),

  get: (id: number, signal?: AbortSignal) =>
    http.get<Middleware>(`/middlewares/${id}`, { signal }),

  create: (input: MiddlewareInput, signal?: AbortSignal) =>
    http.post<{ id: number }>("/middlewares", input, { signal }),

  update: (id: number, input: MiddlewareInput, signal?: AbortSignal) =>
    http.put<{ status: string }>(`/middlewares/${id}`, input, { signal }),

  remove: (id: number, signal?: AbortSignal) =>
    http.delete<{ status: string }>(`/middlewares/${id}`, { signal }),

  listForRoute: (routeId: number, signal?: AbortSignal) =>
    http.get<Middleware[]>(`/routes/${routeId}/middlewares`, { signal }),

  attach: (routeId: number, input: AttachMiddlewareInput, signal?: AbortSignal) =>
    http.post<{ status: string }>(`/routes/${routeId}/middlewares`, input, { signal }),

  detach: (routeId: number, middlewareId: number, signal?: AbortSignal) =>
    http.delete<{ status: string }>(
      `/routes/${routeId}/middlewares/${middlewareId}`,
      { signal },
    ),
};

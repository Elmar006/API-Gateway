import { http } from "./client";
import type { AdminUser, UserInput } from "./types";

export const usersApi = {
  list: (signal?: AbortSignal) =>
    http.get<AdminUser[]>("/users", { signal }),

  get: (id: number, signal?: AbortSignal) =>
    http.get<AdminUser>(`/users/${id}`, { signal }),

  create: (input: UserInput, signal?: AbortSignal) =>
    http.post<{ id: number }>("/users", input, { signal }),

  update: (id: number, input: UserInput, signal?: AbortSignal) =>
    http.put<{ status: string }>(`/users/${id}`, input, { signal }),

  remove: (id: number, signal?: AbortSignal) =>
    http.delete<{ status: string }>(`/users/${id}`, { signal }),
};

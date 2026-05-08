import { http } from "./client";
import type {
  ApiToken,
  CreateTokenInput,
  CreateTokenResponse,
} from "./types";

export const tokensApi = {
  list: (userId: number, signal?: AbortSignal) =>
    http.get<ApiToken[]>(`/users/${userId}/tokens`, { signal }),

  create: (userId: number, input: CreateTokenInput, signal?: AbortSignal) =>
    http.post<CreateTokenResponse>(`/users/${userId}/tokens`, input, { signal }),

  revoke: (id: number, signal?: AbortSignal) =>
    http.delete<{ status: string }>(`/tokens/${id}`, { signal }),
};

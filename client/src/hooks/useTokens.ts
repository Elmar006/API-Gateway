import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { tokensApi } from "@/api";
import type { CreateTokenInput } from "@/api";

const tokensKey = (userId: number) => ["users", userId, "tokens"] as const;

export function useUserTokens(userId: number | null) {
  return useQuery({
    queryKey: ["users", userId ?? 0, "tokens"],
    queryFn: ({ signal }) => tokensApi.list(userId as number, signal),
    enabled: typeof userId === "number" && userId > 0,
  });
}

export function useCreateToken(userId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateTokenInput) => tokensApi.create(userId, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: tokensKey(userId) }),
  });
}

export function useRevokeToken(userId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => tokensApi.revoke(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: tokensKey(userId) }),
  });
}

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { usersApi } from "@/api";
import type { AdminUser, UserInput } from "@/api";

const KEY = ["users"] as const;

export function useUsers() {
  return useQuery({
    queryKey: KEY,
    queryFn: ({ signal }) => usersApi.list(signal),
  });
}

export function useCreateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UserInput) => usersApi.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export function useUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: UserInput }) =>
      usersApi.update(id, input),
    onMutate: async ({ id, input }) => {
      await qc.cancelQueries({ queryKey: KEY });
      const previous = qc.getQueryData<AdminUser[]>(KEY);
      if (previous) {
        qc.setQueryData<AdminUser[]>(
          KEY,
          previous.map((u) =>
            u.id === id
              ? {
                  ...u,
                  username: input.username,
                  role: input.role,
                  updated_at: new Date().toISOString(),
                }
              : u,
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

export function useDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => usersApi.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export const USERS_QUERY_KEY = KEY;

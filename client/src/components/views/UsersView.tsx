import { Plus, Trash2, Users } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import {
  useCreateUser,
  useDeleteUser,
  useUpdateUser,
  useUsers,
} from "@/hooks/useUsers";
import { USER_ROLES } from "@/api";
import type { AdminUser, UserRole } from "@/api";
import { useAppStore } from "@/store/appStore";

export function UsersView() {
  const { data: users = [], isLoading, error } = useUsers();
  const create = useCreateUser();
  const remove = useDeleteUser();
  const setSelectedUser = useAppStore((s) => s.setSelectedUser);

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<UserRole>("editor");
  const canCreate =
    username.trim().length > 0 && password.length >= 8 && !create.isPending;

  return (
    <div className="surface-glass flex h-full flex-col rounded-xl">
      <header className="flex items-center justify-between gap-3 border-b border-line-subtle/70 px-4 py-3">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Users className="h-4 w-4 text-accent-soft" />
          <span className="font-medium text-text-primary">Users</span>
          <span className="text-xs text-text-muted">{users.length} total</span>
        </div>
      </header>

      <form
        className="grid grid-cols-[1fr_1fr_140px_auto] items-center gap-2 border-b border-line-subtle/40 px-4 py-3"
        onSubmit={async (e) => {
          e.preventDefault();
          if (!canCreate) return;
          await create.mutateAsync({ username: username.trim(), password, role });
          setUsername("");
          setPassword("");
        }}
      >
        <Input
          placeholder="alice"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
        <Input
          type="password"
          placeholder="min 8 chars"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <select
          className="h-9 rounded-md border border-line-subtle bg-bg-glass/60 px-2 text-sm"
          value={role}
          onChange={(e) => setRole(e.target.value as UserRole)}
        >
          {USER_ROLES.map((r) => (
            <option key={r} value={r} className="bg-bg-panel">
              {r}
            </option>
          ))}
        </select>
        <Button type="submit" variant="primary" size="md" disabled={!canCreate}>
          <Plus className="h-3.5 w-3.5" /> Add
        </Button>
      </form>

      {error ? (
        <div className="grid flex-1 place-items-center text-sm text-status-err">
          Failed to load users: {(error as Error).message}
        </div>
      ) : isLoading ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">Loading…</div>
      ) : users.length === 0 ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">No users yet.</div>
      ) : (
        <ul className="flex-1 divide-y divide-line-subtle/40 overflow-auto">
          {users.map((u) => (
            <UserRow
              key={u.id}
              user={u}
              onTokens={() => {
                setSelectedUser(u.id);
                useAppStore.getState().setSection("tokens");
              }}
              onDelete={() => remove.mutate(u.id)}
            />
          ))}
        </ul>
      )}
    </div>
  );
}

function UserRow({
  user,
  onTokens,
  onDelete,
}: {
  user: AdminUser;
  onTokens: () => void;
  onDelete: () => void;
}) {
  const update = useUpdateUser();
  const [role, setRole] = useState<UserRole>(user.role);
  const dirty = role !== user.role;

  return (
    <li className="grid grid-cols-[1fr_140px_auto_auto_auto] items-center gap-3 px-4 py-2.5 text-sm">
      <span className="font-medium text-text-primary">{user.username}</span>
      <select
        className="h-9 rounded-md border border-line-subtle bg-bg-glass/60 px-2 text-sm"
        value={role}
        onChange={(e) => setRole(e.target.value as UserRole)}
      >
        {USER_ROLES.map((r) => (
          <option key={r} value={r} className="bg-bg-panel">
            {r}
          </option>
        ))}
      </select>
      <Button
        variant="ghost"
        size="md"
        disabled={!dirty || update.isPending}
        onClick={() =>
          update.mutate({
            id: user.id,
            input: { username: user.username, role },
          })
        }
      >
        Save
      </Button>
      <Button variant="ghost" size="md" onClick={onTokens}>
        Tokens
      </Button>
      <Button variant="ghost" size="icon-sm" aria-label="Delete user" onClick={onDelete}>
        <Trash2 className="h-3.5 w-3.5" />
      </Button>
    </li>
  );
}

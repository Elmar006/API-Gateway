import { Copy, KeyRound, Plus, Trash2 } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { useUsers } from "@/hooks/useUsers";
import {
  useCreateToken,
  useRevokeToken,
  useUserTokens,
} from "@/hooks/useTokens";
import { ALL_SCOPES } from "@/api";
import type { ApiToken, CreateTokenResponse, Scope } from "@/api";
import { useAppStore } from "@/store/appStore";

export function TokensView() {
  const { data: users = [] } = useUsers();
  const selectedUserId = useAppStore((s) => s.selectedUserId);
  const setSelectedUser = useAppStore((s) => s.setSelectedUser);

  const userId = selectedUserId ?? users[0]?.id ?? null;
  const { data: tokens = [], isLoading, error } = useUserTokens(userId);

  return (
    <div className="surface-glass flex h-full flex-col rounded-xl">
      <header className="flex items-center justify-between gap-3 border-b border-line-subtle/70 px-4 py-3">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <KeyRound className="h-4 w-4 text-accent-soft" />
          <span className="font-medium text-text-primary">API tokens</span>
          <span className="text-xs text-text-muted">{tokens.length} active for selected user</span>
        </div>
        <select
          className="h-8 rounded-md border border-line-subtle bg-bg-glass/60 px-2 text-xs"
          value={userId ?? ""}
          onChange={(e) => setSelectedUser(Number(e.target.value) || null)}
        >
          <option value="" className="bg-bg-panel">
            — pick user —
          </option>
          {users.map((u) => (
            <option key={u.id} value={u.id} className="bg-bg-panel">
              {u.username} ({u.role})
            </option>
          ))}
        </select>
      </header>

      {userId ? (
        <CreateTokenForm userId={userId} />
      ) : (
        <div className="border-b border-line-subtle/40 px-4 py-3 text-xs text-text-muted">
          Select a user to issue tokens.
        </div>
      )}

      {error ? (
        <div className="grid flex-1 place-items-center text-sm text-status-err">
          Failed to load tokens: {(error as Error).message}
        </div>
      ) : !userId ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">
          No user selected.
        </div>
      ) : isLoading ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">Loading…</div>
      ) : tokens.length === 0 ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">
          No tokens for this user yet.
        </div>
      ) : (
        <ul className="flex-1 divide-y divide-line-subtle/40 overflow-auto">
          {tokens.map((t) => (
            <TokenRow key={t.id} token={t} userId={userId} />
          ))}
        </ul>
      )}
    </div>
  );
}

function CreateTokenForm({ userId }: { userId: number }) {
  const create = useCreateToken(userId);
  const [name, setName] = useState("");
  const [scopes, setScopes] = useState<Set<Scope>>(new Set(["routes:read"]));
  const [issued, setIssued] = useState<CreateTokenResponse | null>(null);

  return (
    <div className="border-b border-line-subtle/40 px-4 py-3">
      <form
        className="grid grid-cols-[1fr_auto] items-start gap-3"
        onSubmit={async (e) => {
          e.preventDefault();
          if (!name.trim() || scopes.size === 0) return;
          const res = await create.mutateAsync({
            name: name.trim(),
            scopes: Array.from(scopes),
          });
          setIssued(res);
          setName("");
          setScopes(new Set(["routes:read"]));
        }}
      >
        <div className="flex flex-col gap-2">
          <Input
            placeholder="ci-bot"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <div className="flex flex-wrap gap-1">
            {ALL_SCOPES.map((s) => {
              const on = scopes.has(s);
              return (
                <button
                  key={s}
                  type="button"
                  onClick={() => {
                    setScopes((prev) => {
                      const next = new Set(prev);
                      if (on) next.delete(s);
                      else next.add(s);
                      return next;
                    });
                  }}
                  className={`rounded-md border px-1.5 py-0.5 font-mono text-[10px] ${
                    on
                      ? "border-accent-ring bg-accent/15 text-text-primary"
                      : "border-line-subtle bg-bg-glass/40 text-text-muted hover:text-text-primary"
                  }`}
                >
                  {s}
                </button>
              );
            })}
          </div>
        </div>
        <Button
          type="submit"
          variant="primary"
          size="md"
          disabled={!name.trim() || scopes.size === 0 || create.isPending}
        >
          <Plus className="h-3.5 w-3.5" /> Issue
        </Button>
      </form>

      {issued && (
        <div className="mt-3 flex items-center gap-2 rounded-md border border-accent-ring/40 bg-accent/10 px-3 py-2 text-xs">
          <span className="text-text-muted">Secret (shown once):</span>
          <code className="flex-1 truncate font-mono text-text-primary">{issued.secret}</code>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Copy"
            onClick={() => navigator.clipboard.writeText(issued.secret)}
          >
            <Copy className="h-3.5 w-3.5" />
          </Button>
          <Button variant="ghost" size="md" onClick={() => setIssued(null)}>
            Dismiss
          </Button>
        </div>
      )}
    </div>
  );
}

function TokenRow({ token, userId }: { token: ApiToken; userId: number }) {
  const revoke = useRevokeToken(userId);
  const revoked = token.revoked_at !== null && token.revoked_at !== undefined;
  return (
    <li className="grid grid-cols-[1fr_2fr_140px_120px_auto] items-center gap-3 px-4 py-2.5 text-sm">
      <span className="font-medium text-text-primary">{token.name}</span>
      <span className="truncate font-mono text-xs text-text-muted">{token.scopes}</span>
      <span className="font-mono text-xs text-text-muted">{token.prefix}…</span>
      <span className={revoked ? "text-status-err" : "text-status-ok"}>
        {revoked ? "revoked" : "active"}
      </span>
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label="Revoke"
        disabled={revoked || revoke.isPending}
        onClick={() => revoke.mutate(token.id)}
      >
        <Trash2 className="h-3.5 w-3.5" />
      </Button>
    </li>
  );
}

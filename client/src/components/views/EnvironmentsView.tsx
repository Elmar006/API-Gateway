import { Plus, Share2, Trash2 } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import {
  useCreateEnvironment,
  useDeleteEnvironment,
  useEnvironments,
  useUpdateEnvironment,
} from "@/hooks/useEnvironments";
import type { Environment } from "@/api";

export function EnvironmentsView() {
  const { data: envs = [], isLoading, error } = useEnvironments();
  const create = useCreateEnvironment();
  const remove = useDeleteEnvironment();

  const [name, setName] = useState("");
  const [domain, setDomain] = useState("");
  const [color, setColor] = useState("#3b82f6");
  const canCreate = name.trim().length > 0 && !create.isPending;

  return (
    <div className="surface-glass flex h-full flex-col rounded-xl">
      <header className="flex items-center justify-between gap-3 border-b border-line-subtle/70 px-4 py-3">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Share2 className="h-4 w-4 text-accent-soft" />
          <span className="font-medium text-text-primary">Environments</span>
          <span className="text-xs text-text-muted">{envs.length} total</span>
        </div>
      </header>

      <form
        className="grid grid-cols-[1fr_1fr_120px_auto] items-center gap-2 border-b border-line-subtle/40 px-4 py-3"
        onSubmit={async (e) => {
          e.preventDefault();
          if (!canCreate) return;
          await create.mutateAsync({
            name: name.trim(),
            base_domain: domain.trim(),
            color,
          });
          setName("");
          setDomain("");
        }}
      >
        <Input
          placeholder="staging"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <Input
          placeholder="staging.example.com"
          value={domain}
          onChange={(e) => setDomain(e.target.value)}
          className="font-mono text-xs"
        />
        <input
          type="color"
          value={color}
          onChange={(e) => setColor(e.target.value)}
          className="h-9 w-full rounded-md border border-line-subtle bg-bg-glass/60"
          aria-label="Environment color"
        />
        <Button type="submit" variant="primary" size="md" disabled={!canCreate}>
          <Plus className="h-3.5 w-3.5" /> Add
        </Button>
      </form>

      {error ? (
        <div className="grid flex-1 place-items-center text-sm text-status-err">
          Failed to load environments: {(error as Error).message}
        </div>
      ) : isLoading ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">Loading…</div>
      ) : envs.length === 0 ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">
          No environments yet.
        </div>
      ) : (
        <div className="flex-1 overflow-auto">
          <ul className="divide-y divide-line-subtle/40">
            {envs.map((env) => (
              <EnvRow key={env.id} env={env} onDelete={() => remove.mutate(env.id)} />
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

function EnvRow({ env, onDelete }: { env: Environment; onDelete: () => void }) {
  const update = useUpdateEnvironment();
  const [name, setName] = useState(env.name);
  const [domain, setDomain] = useState(env.base_domain);
  const dirty = name !== env.name || domain !== env.base_domain;

  return (
    <li className="grid grid-cols-[12px_1fr_1fr_88px_auto] items-center gap-3 px-4 py-2.5 text-sm">
      <span
        className="h-3 w-3 rounded-full border border-line-subtle"
        style={{ background: env.color }}
        title={env.color}
      />
      <Input value={name} onChange={(e) => setName(e.target.value)} />
      <Input
        value={domain}
        onChange={(e) => setDomain(e.target.value)}
        className="font-mono text-xs"
      />
      <Button
        variant="ghost"
        size="md"
        disabled={!dirty || update.isPending}
        onClick={() =>
          update.mutate({
            id: env.id,
            input: { name, base_domain: domain, color: env.color, description: env.description },
          })
        }
      >
        Save
      </Button>
      <Button variant="ghost" size="icon-sm" aria-label="Delete" onClick={onDelete}>
        <Trash2 className="h-3.5 w-3.5" />
      </Button>
    </li>
  );
}

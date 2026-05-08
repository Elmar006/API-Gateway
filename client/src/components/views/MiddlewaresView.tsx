import { Layers, Plus, Trash2 } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Switch } from "@/components/ui/Switch";
import {
  useCreateMiddleware,
  useDeleteMiddleware,
  useMiddlewares,
  useUpdateMiddleware,
} from "@/hooks/useMiddlewares";
import { MIDDLEWARE_KINDS } from "@/api";
import type { Middleware, MiddlewareKind } from "@/api";

const KIND_DEFAULTS: Record<MiddlewareKind, Record<string, unknown>> = {
  cors: { origins: ["*"], methods: ["GET", "POST"], headers: ["*"] },
  rate_limit_override: { rps: 100, burst: 50 },
  header_rewrite: { set: { "X-Edge": "routeflow" } },
  request_id: { header: "X-Request-Id" },
  strip_prefix: { prefix: "/api" },
};

export function MiddlewaresView() {
  const { data: mws = [], isLoading, error } = useMiddlewares();
  const create = useCreateMiddleware();
  const remove = useDeleteMiddleware();

  const [name, setName] = useState("");
  const [kind, setKind] = useState<MiddlewareKind>("cors");
  const canCreate = name.trim().length > 0 && !create.isPending;

  return (
    <div className="surface-glass flex h-full flex-col rounded-xl">
      <header className="flex items-center justify-between gap-3 border-b border-line-subtle/70 px-4 py-3">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Layers className="h-4 w-4 text-accent-soft" />
          <span className="font-medium text-text-primary">Middlewares</span>
          <span className="text-xs text-text-muted">{mws.length} total</span>
        </div>
      </header>

      <form
        className="grid grid-cols-[1fr_180px_auto] items-center gap-2 border-b border-line-subtle/40 px-4 py-3"
        onSubmit={async (e) => {
          e.preventDefault();
          if (!canCreate) return;
          await create.mutateAsync({
            name: name.trim(),
            kind,
            config: KIND_DEFAULTS[kind],
            is_active: true,
          });
          setName("");
        }}
      >
        <Input
          placeholder="cors-default"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <select
          className="h-9 rounded-md border border-line-subtle bg-bg-glass/60 px-2 text-sm"
          value={kind}
          onChange={(e) => setKind(e.target.value as MiddlewareKind)}
        >
          {MIDDLEWARE_KINDS.map((k) => (
            <option key={k} value={k} className="bg-bg-panel">
              {k}
            </option>
          ))}
        </select>
        <Button type="submit" variant="primary" size="md" disabled={!canCreate}>
          <Plus className="h-3.5 w-3.5" /> Add
        </Button>
      </form>

      {error ? (
        <div className="grid flex-1 place-items-center text-sm text-status-err">
          Failed to load middlewares: {(error as Error).message}
        </div>
      ) : isLoading ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">Loading…</div>
      ) : mws.length === 0 ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">
          No middlewares yet.
        </div>
      ) : (
        <ul className="flex-1 divide-y divide-line-subtle/40 overflow-auto">
          {mws.map((m) => (
            <MiddlewareRow key={m.id} mw={m} onDelete={() => remove.mutate(m.id)} />
          ))}
        </ul>
      )}
    </div>
  );
}

function MiddlewareRow({ mw, onDelete }: { mw: Middleware; onDelete: () => void }) {
  const update = useUpdateMiddleware();
  const [active, setActive] = useState(mw.is_active);
  const [text, setText] = useState(JSON.stringify(mw.config, null, 2));
  const [parseErr, setParseErr] = useState<string | null>(null);

  return (
    <li className="grid grid-cols-[1fr_2fr_88px_auto_auto] items-start gap-3 px-4 py-3 text-sm">
      <div className="flex flex-col gap-1">
        <span className="font-medium text-text-primary">{mw.name}</span>
        <span className="rounded-md border border-line-subtle bg-bg-glass/40 px-1.5 py-0 text-[10px] uppercase tracking-wider text-text-muted">
          {mw.kind}
        </span>
      </div>
      <div className="flex flex-col gap-1">
        <textarea
          className="h-24 w-full rounded-md border border-line-subtle bg-bg-glass/40 px-2 py-1.5 font-mono text-[11px] text-text-primary"
          value={text}
          onChange={(e) => {
            setText(e.target.value);
            try {
              JSON.parse(e.target.value);
              setParseErr(null);
            } catch (err) {
              setParseErr((err as Error).message);
            }
          }}
        />
        {parseErr && <span className="text-[10px] text-status-err">{parseErr}</span>}
      </div>
      <Switch checked={active} onCheckedChange={setActive} />
      <Button
        variant="ghost"
        size="md"
        disabled={parseErr !== null || update.isPending}
        onClick={() => {
          const cfg = (() => {
            try {
              return JSON.parse(text) as unknown;
            } catch {
              return null;
            }
          })();
          if (cfg === null) return;
          update.mutate({
            id: mw.id,
            input: {
              name: mw.name,
              kind: mw.kind,
              config: cfg,
              is_active: active,
              description: mw.description,
            },
          });
        }}
      >
        Save
      </Button>
      <Button variant="ghost" size="icon-sm" aria-label="Delete" onClick={onDelete}>
        <Trash2 className="h-3.5 w-3.5" />
      </Button>
    </li>
  );
}

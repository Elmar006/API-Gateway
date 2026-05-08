import { Boxes, Plus, Trash2 } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import {
  useAddClusterTarget,
  useClusters,
  useCreateCluster,
  useDeleteCluster,
  useRemoveClusterTarget,
} from "@/hooks/useClusters";
import { CLUSTER_STRATEGIES } from "@/api";
import type { Cluster, ClusterStrategy } from "@/api";

export function ClustersView() {
  const { data: clusters = [], isLoading, error } = useClusters();
  const create = useCreateCluster();
  const remove = useDeleteCluster();

  const [name, setName] = useState("");
  const [strategy, setStrategy] = useState<ClusterStrategy>("round_robin");
  const canCreate = name.trim().length > 0 && !create.isPending;

  return (
    <div className="surface-glass flex h-full flex-col rounded-xl">
      <header className="flex items-center justify-between gap-3 border-b border-line-subtle/70 px-4 py-3">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Boxes className="h-4 w-4 text-accent-soft" />
          <span className="font-medium text-text-primary">Clusters</span>
          <span className="text-xs text-text-muted">{clusters.length} total</span>
        </div>
      </header>

      <form
        className="grid grid-cols-[1fr_180px_auto] items-center gap-2 border-b border-line-subtle/40 px-4 py-3"
        onSubmit={async (e) => {
          e.preventDefault();
          if (!canCreate) return;
          await create.mutateAsync({ name: name.trim(), strategy });
          setName("");
        }}
      >
        <Input
          placeholder="primary"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <select
          className="h-9 rounded-md border border-line-subtle bg-bg-glass/60 px-2 text-sm"
          value={strategy}
          onChange={(e) => setStrategy(e.target.value as ClusterStrategy)}
        >
          {CLUSTER_STRATEGIES.map((s) => (
            <option key={s} value={s} className="bg-bg-panel">
              {s}
            </option>
          ))}
        </select>
        <Button type="submit" variant="primary" size="md" disabled={!canCreate}>
          <Plus className="h-3.5 w-3.5" /> Add
        </Button>
      </form>

      {error ? (
        <div className="grid flex-1 place-items-center text-sm text-status-err">
          Failed to load clusters: {(error as Error).message}
        </div>
      ) : isLoading ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">Loading…</div>
      ) : clusters.length === 0 ? (
        <div className="grid flex-1 place-items-center text-sm text-text-muted">
          No clusters yet.
        </div>
      ) : (
        <ul className="flex-1 divide-y divide-line-subtle/40 overflow-auto">
          {clusters.map((c) => (
            <ClusterCard key={c.id} cluster={c} onDelete={() => remove.mutate(c.id)} />
          ))}
        </ul>
      )}
    </div>
  );
}

function ClusterCard({ cluster, onDelete }: { cluster: Cluster; onDelete: () => void }) {
  const addTarget = useAddClusterTarget();
  const removeTarget = useRemoveClusterTarget();
  const [url, setUrl] = useState("");
  const [weight, setWeight] = useState(1);

  return (
    <li className="px-4 py-3">
      <div className="flex items-center gap-3">
        <span className="rounded-md border border-line-subtle bg-bg-glass/40 px-2 py-0.5 text-xs uppercase tracking-wider text-text-muted">
          {cluster.strategy}
        </span>
        <span className="text-sm font-medium text-text-primary">{cluster.name}</span>
        <span className="ml-auto text-xs text-text-muted">
          {cluster.targets?.length ?? 0} targets
        </span>
        <Button variant="ghost" size="icon-sm" aria-label="Delete cluster" onClick={onDelete}>
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>

      <ul className="mt-2 space-y-1">
        {(cluster.targets ?? []).map((t) => (
          <li
            key={t.id}
            className="grid grid-cols-[1fr_60px_88px_auto] items-center gap-2 rounded-md border border-line-subtle/40 bg-bg-glass/30 px-2.5 py-1.5 text-xs"
          >
            <span className="truncate font-mono">{t.url}</span>
            <span className="text-right text-text-muted">w={t.weight}</span>
            <span
              className={
                t.is_healthy ? "text-status-ok" : "text-status-err"
              }
            >
              {t.is_healthy ? "healthy" : "unhealthy"}
            </span>
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label="Remove target"
              onClick={() =>
                removeTarget.mutate({ clusterId: cluster.id, targetId: t.id })
              }
            >
              <Trash2 className="h-3 w-3" />
            </Button>
          </li>
        ))}
      </ul>

      <form
        className="mt-2 grid grid-cols-[1fr_80px_auto] items-center gap-2"
        onSubmit={async (e) => {
          e.preventDefault();
          if (!url.trim()) return;
          await addTarget.mutateAsync({
            clusterId: cluster.id,
            input: { url: url.trim(), weight: Number(weight) || 1 },
          });
          setUrl("");
          setWeight(1);
        }}
      >
        <Input
          placeholder="http://upstream:8080"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          className="font-mono text-xs"
        />
        <Input
          type="number"
          min={1}
          value={weight}
          onChange={(e) => setWeight(Number(e.target.value))}
        />
        <Button type="submit" variant="ghost" size="md" disabled={addTarget.isPending}>
          <Plus className="h-3.5 w-3.5" /> Target
        </Button>
      </form>
    </li>
  );
}

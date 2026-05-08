import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Network } from "lucide-react";
import { memo } from "react";

import { NodeShell } from "@/components/nodes/NodeShell";
import { StatusDot } from "@/components/ui/StatusDot";
import type { GatewayNodeInfo } from "@/types/domain";

export interface GatewayNodeData extends Record<string, unknown> {
  gateway: GatewayNodeInfo;
  selected?: boolean;
  highlighted?: boolean;
  onSelect: () => void;
  onHover: (id: string | null) => void;
}

function GatewayNodeImpl({ data }: NodeProps) {
  const d = data as unknown as GatewayNodeData;
  const { gateway } = d;
  return (
    <NodeShell
      tone="accent"
      size="lg"
      selected={d.selected}
      highlighted={d.highlighted}
      onClick={d.onSelect}
      onMouseEnter={() => d.onHover(gateway.id)}
      onMouseLeave={() => d.onHover(null)}
      className="min-w-[208px]"
    >
      <Handle type="target" position={Position.Left} className="!opacity-60" />
      <Handle type="source" position={Position.Right} className="!opacity-60" />
      <div className="relative flex flex-col items-center gap-2 py-1">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-x-3 top-1 h-12 rounded-full bg-accent/30 blur-xl opacity-60"
        />
        <div className="relative grid h-10 w-10 place-items-center rounded-md border border-accent/40 bg-bg-glass/70 text-accent-soft shadow-[inset_0_0_0_1px_hsl(var(--glass-highlight)/0.12)]">
          <Network className="h-4 w-4" />
        </div>
        <div className="relative flex items-center gap-2 text-sm font-medium text-text-primary">
          {gateway.label}
          <StatusDot status={gateway.status} className="h-1.5 w-1.5" />
        </div>
        <div className="relative text-[11px] uppercase tracking-[0.18em] text-text-muted">
          {gateway.endpoint}
        </div>
      </div>
    </NodeShell>
  );
}

export const GatewayNode = memo(GatewayNodeImpl);

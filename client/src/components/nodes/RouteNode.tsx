import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Route as RouteIcon } from "lucide-react";
import { memo } from "react";

import { NodeShell } from "@/components/nodes/NodeShell";
import { StatusDot } from "@/components/ui/StatusDot";
import { statusForRoute } from "@/lib/derive";
import type { Route } from "@/types/domain";

export interface RouteNodeData extends Record<string, unknown> {
  route: Route;
  selected?: boolean;
  highlighted?: boolean;
  onSelect: () => void;
  onHover: (id: string | null) => void;
}

function RouteNodeImpl({ data }: NodeProps) {
  const d = data as unknown as RouteNodeData;
  const { route } = d;
  const status = statusForRoute(route);
  const id = `route:${route.id}`;
  return (
    <NodeShell
      selected={d.selected}
      highlighted={d.highlighted}
      onClick={d.onSelect}
      onMouseEnter={() => d.onHover(id)}
      onMouseLeave={() => d.onHover(null)}
    >
      <Handle type="target" position={Position.Left} className="!opacity-60" />
      <Handle type="source" position={Position.Right} className="!opacity-60" />
      <div className="flex items-center gap-2.5">
        <div className="grid h-7 w-7 place-items-center rounded-md border border-line-subtle bg-bg-chip/60 text-accent-soft">
          <RouteIcon className="h-3.5 w-3.5" />
        </div>
        <div className="flex flex-1 flex-col leading-tight">
          <div className="flex items-center gap-2 text-sm text-text-primary">
            <span className="rounded bg-bg-chip/70 px-1.5 py-0.5 font-mono text-[10px] uppercase text-text-secondary">
              {route.method}
            </span>
            <StatusDot status={status} className="h-1.5 w-1.5" />
          </div>
          <div className="truncate font-mono text-[11px] text-text-muted" title={route.path_pattern}>
            {route.path_pattern}
          </div>
        </div>
      </div>
    </NodeShell>
  );
}

export const RouteNode = memo(RouteNodeImpl);

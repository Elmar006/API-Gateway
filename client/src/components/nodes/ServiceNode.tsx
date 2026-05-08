import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Boxes } from "lucide-react";
import { memo } from "react";

import { NodeShell } from "@/components/nodes/NodeShell";
import { StatusDot } from "@/components/ui/StatusDot";
import { statusForService } from "@/lib/derive";
import type { DerivedService } from "@/types/domain";

export interface ServiceNodeData extends Record<string, unknown> {
  service: DerivedService;
  selected?: boolean;
  highlighted?: boolean;
  onSelect: () => void;
  onHover: (id: string | null) => void;
}

function ServiceNodeImpl({ data }: NodeProps) {
  const d = data as unknown as ServiceNodeData;
  const { service } = d;
  const status = statusForService(service);
  return (
    <NodeShell
      selected={d.selected}
      highlighted={d.highlighted}
      onClick={d.onSelect}
      onMouseEnter={() => d.onHover(service.id)}
      onMouseLeave={() => d.onHover(null)}
    >
      <Handle type="target" position={Position.Left} className="!opacity-60" />
      <div className="flex items-center gap-2.5">
        <div className="grid h-7 w-7 place-items-center rounded-md border border-line-subtle bg-bg-chip/60 text-text-secondary">
          <Boxes className="h-3.5 w-3.5" />
        </div>
        <div className="flex flex-1 flex-col leading-tight">
          <div className="flex items-center gap-2 text-sm text-text-primary">
            <span className="truncate" title={service.host}>{service.host}</span>
            <StatusDot status={status} className="h-1.5 w-1.5" />
          </div>
          <div className="text-[11px] text-text-muted">
            {service.scheme.toUpperCase()} • {service.routeCount} route
            {service.routeCount === 1 ? "" : "s"}
          </div>
        </div>
      </div>
    </NodeShell>
  );
}

export const ServiceNode = memo(ServiceNodeImpl);

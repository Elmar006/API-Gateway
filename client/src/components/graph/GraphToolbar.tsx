import { useReactFlow } from "@xyflow/react";
import {
  Crosshair,
  Filter,
  LocateFixed,
  Maximize,
  Minimize,
  RotateCcw,
  Search,
} from "lucide-react";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { useGraphStore } from "@/store/graphStore";

export function GraphToolbar() {
  const { fitView, zoomIn, zoomOut } = useReactFlow();
  const filter = useGraphStore((s) => s.graphFilter);
  const setFilter = useGraphStore((s) => s.setGraphFilter);

  return (
    <div className="absolute left-4 right-4 top-4 z-10 flex items-center justify-between gap-3">
      <div className="surface-glass flex items-center gap-1 rounded-md p-1">
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={() => fitView({ padding: 0.18, duration: 320 })}
          title="Fit to canvas"
          aria-label="Fit"
        >
          <LocateFixed className="h-3.5 w-3.5" />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={() => zoomIn({ duration: 220 })}
          title="Zoom in"
          aria-label="Zoom in"
        >
          <Maximize className="h-3.5 w-3.5" />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={() => zoomOut({ duration: 220 })}
          title="Zoom out"
          aria-label="Zoom out"
        >
          <Minimize className="h-3.5 w-3.5" />
        </Button>
        <span className="mx-1 h-4 w-px bg-line-subtle" />
        <Button
          variant="ghost"
          size="icon-sm"
          title="Center selection"
          onClick={() => fitView({ padding: 0.4, duration: 320, includeHiddenNodes: false })}
          aria-label="Center"
        >
          <Crosshair className="h-3.5 w-3.5" />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          title="Reset view"
          onClick={() => fitView({ padding: 0.2, duration: 380 })}
          aria-label="Reset"
        >
          <RotateCcw className="h-3.5 w-3.5" />
        </Button>
      </div>

      <div className="surface-glass relative flex h-9 items-center gap-2 rounded-md px-2.5">
        <Filter className="h-3.5 w-3.5 text-text-muted" />
        <Input
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          placeholder="Filter nodes…"
          className="h-7 w-44 border-none bg-transparent px-1 text-xs focus-visible:shadow-none"
        />
        <Search className="h-3.5 w-3.5 text-text-muted" />
      </div>
    </div>
  );
}

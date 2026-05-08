import { Maximize2, Minus, Square, X } from "lucide-react";

import { Button } from "@/components/ui/Button";
import { useTitleBar } from "@/hooks/useTitleBar";

export function WindowControls() {
  const { minimize, toggleMaximize, close, isMaximized, isTauri } = useTitleBar();
  if (!isTauri) return null;

  return (
    <div className="titlebar-no-drag flex items-center gap-1">
      <Button
        variant="ghost"
        size="icon-sm"
        onClick={() => minimize()}
        aria-label="Minimize"
        title="Minimize"
      >
        <Minus className="h-3.5 w-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon-sm"
        onClick={() => toggleMaximize()}
        aria-label={isMaximized ? "Restore" : "Maximize"}
        title={isMaximized ? "Restore" : "Maximize"}
      >
        {isMaximized ? <Square className="h-3 w-3" /> : <Maximize2 className="h-3 w-3" />}
      </Button>
      <Button
        variant="ghost"
        size="icon-sm"
        onClick={() => close()}
        aria-label="Close"
        title="Close"
        className="hover:bg-status-err/30 hover:text-status-err hover:border-status-err/40"
      >
        <X className="h-3.5 w-3.5" />
      </Button>
    </div>
  );
}

import { Bell } from "lucide-react";

import { Logo } from "@/components/common/Logo";
import { Button } from "@/components/ui/Button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/Tooltip";
import { useErrorCount } from "@/hooks/useLogs";
import { useTitleBar } from "@/hooks/useTitleBar";
import { useAppStore } from "@/store/appStore";

import { AccountChip } from "./AccountChip";
import { GlobalSearch } from "./GlobalSearch";
import { WindowControls } from "./WindowControls";

export function TitleBar() {
  const setSection = useAppStore((s) => s.setSection);
  const { startDrag } = useTitleBar();

  // Real "unread alerts" badge — count of 5xx responses in the past hour.
  const { data: errorCount = 0 } = useErrorCount(60, 30_000);

  return (
    <header
      className="titlebar-drag relative z-30 flex h-14 shrink-0 items-center gap-4 border-b border-line-subtle/80 bg-bg-panel/70 px-4 backdrop-blur-md"
      onMouseDown={(e) => {
        if (e.button !== 0) return;
        const target = e.target as HTMLElement;
        if (target.closest("[data-no-drag], .titlebar-no-drag")) return;
        void startDrag();
      }}
    >
      <div className="flex items-center gap-2.5">
        <Logo />
        <div className="leading-tight">
          <div className="text-sm font-semibold text-text-primary tracking-tight">RouteFlow</div>
          <div className="text-[10px] uppercase tracking-[0.2em] text-text-muted">Control Plane</div>
        </div>
      </div>

      <div className="flex flex-1 items-center justify-center">
        <GlobalSearch />
      </div>

      <div className="flex items-center gap-2">
        <TooltipProvider delayDuration={200}>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setSection("logs")}
                aria-label="Notifications"
                className="titlebar-no-drag relative"
              >
                <Bell className="h-4 w-4" />
                {errorCount > 0 && (
                  <span className="absolute right-2 top-2 inline-flex h-1.5 w-1.5 rounded-full bg-status-err shadow-[0_0_6px_hsl(var(--status-err)/0.7)]" />
                )}
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              {errorCount > 0
                ? `${errorCount} server error${errorCount === 1 ? "" : "s"} in the last hour`
                : "No errors in the last hour"}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>

        <AccountChip />

        <span className="mx-1 hidden h-5 w-px bg-line-subtle/80 md:block" />
        <WindowControls />
      </div>
    </header>
  );
}

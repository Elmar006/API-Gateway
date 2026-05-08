import { cn } from "@/lib/cn";
import type { Status } from "@/types/domain";

const tone: Record<Status, string> = {
  active: "bg-status-ok shadow-[0_0_8px_hsl(var(--status-ok)/0.6)]",
  degraded: "bg-status-warn shadow-[0_0_8px_hsl(var(--status-warn)/0.6)]",
  inactive: "bg-text-muted",
};

export function StatusDot({ status, className }: { status: Status; className?: string }) {
  return (
    <span
      aria-hidden
      className={cn(
        "relative inline-flex h-2 w-2 shrink-0 rounded-full",
        tone[status],
        status === "active" && "before:absolute before:inset-0 before:rounded-full before:animate-pulse-soft before:bg-status-ok/60",
        className,
      )}
    />
  );
}

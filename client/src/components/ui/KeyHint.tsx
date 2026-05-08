import type { ReactNode } from "react";

import { cn } from "@/lib/cn";

export function KeyHint({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <kbd
      className={cn(
        "inline-flex items-center gap-0.5 rounded-md border border-line-subtle bg-bg-chip/70",
        "px-1.5 py-0.5 text-[10px] font-medium text-text-muted shadow-edge",
        "tracking-wider",
        className,
      )}
    >
      {children}
    </kbd>
  );
}

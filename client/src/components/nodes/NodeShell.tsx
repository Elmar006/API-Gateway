import { type ReactNode } from "react";

import { cn } from "@/lib/cn";

export interface NodeShellProps {
  selected?: boolean;
  highlighted?: boolean;
  size?: "md" | "lg";
  tone?: "default" | "accent";
  className?: string;
  children: ReactNode;
  onClick?: () => void;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
}

export function NodeShell({
  selected,
  highlighted,
  size = "md",
  tone = "default",
  className,
  children,
  ...rest
}: NodeShellProps) {
  return (
    <div
      {...rest}
      className={cn(
        "group relative isolate cursor-pointer rounded-xl",
        "border border-line-subtle/90",
        "bg-[linear-gradient(160deg,hsl(var(--bg-glass)/0.85)_0%,hsl(var(--bg-glass)/0.45)_100%)]",
        "backdrop-blur-md",
        "shadow-[inset_0_1px_0_hsl(var(--glass-highlight)/0.06),0_18px_36px_-22px_rgb(8_4_32/0.7)]",
        "transition-[transform,border-color,box-shadow] duration-200",
        "hover:-translate-y-[1px] hover:border-accent/60",
        size === "md" && "min-w-[180px] px-3 py-2.5",
        size === "lg" && "min-w-[180px] px-4 py-3.5",
        tone === "accent" &&
          "border-accent/55 bg-[linear-gradient(160deg,hsl(263_60%_22%/0.86)_0%,hsl(263_50%_18%/0.6)_100%)]",
        selected &&
          "border-accent shadow-[inset_0_1px_0_hsl(var(--glass-highlight)/0.12),0_0_0_1px_hsl(var(--accent)/0.55),0_0_28px_-2px_hsl(var(--accent)/0.55)]",
        highlighted && !selected && "border-accent/60",
        className,
      )}
    >
      <span
        aria-hidden
        className="pointer-events-none absolute inset-x-3 top-0 h-px bg-gradient-to-r from-transparent via-white/35 to-transparent opacity-50"
      />
      {children}
    </div>
  );
}

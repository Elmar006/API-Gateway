import { cva, type VariantProps } from "class-variance-authority";
import type { HTMLAttributes } from "react";

import { cn } from "@/lib/cn";

const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[11px] font-medium tracking-wide",
  {
    variants: {
      tone: {
        ok: "border-status-ok/40 bg-status-ok/12 text-status-ok",
        warn: "border-status-warn/40 bg-status-warn/12 text-status-warn",
        err: "border-status-err/40 bg-status-err/12 text-status-err",
        info: "border-status-info/40 bg-status-info/14 text-status-info",
        muted: "border-line-subtle bg-bg-chip/70 text-text-muted",
        accent: "border-accent/40 bg-accent/16 text-accent-soft",
      },
    },
    defaultVariants: { tone: "muted" },
  },
);

export interface BadgeProps
  extends HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {
  dot?: boolean;
}

export function Badge({ className, tone, dot, children, ...props }: BadgeProps) {
  return (
    <span className={cn(badgeVariants({ tone }), className)} {...props}>
      {dot && (
        <span
          className={cn("h-1.5 w-1.5 rounded-full", {
            "bg-status-ok": tone === "ok",
            "bg-status-warn": tone === "warn",
            "bg-status-err": tone === "err",
            "bg-status-info": tone === "info",
            "bg-text-muted": tone === "muted" || tone === undefined,
            "bg-accent": tone === "accent",
          })}
        />
      )}
      {children}
    </span>
  );
}

import { cva } from "class-variance-authority";

export const buttonVariants = cva(
  [
    "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium",
    "transition-[transform,background-color,border-color,box-shadow,color] duration-150",
    "focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-accent-ring focus-visible:shadow-ring",
    "disabled:pointer-events-none disabled:opacity-50",
    "select-none active:translate-y-[0.5px]",
  ].join(" "),
  {
    variants: {
      variant: {
        primary:
          "bg-accent-gradient text-text-inverse shadow-[0_8px_24px_-12px_hsl(var(--accent-ring)/0.55)] hover:shadow-[0_10px_24px_-10px_hsl(var(--accent-ring)/0.65)] hover:saturate-110",
        ghost:
          "bg-transparent text-text-secondary hover:bg-bg-glass/80 hover:text-text-primary border border-transparent hover:border-line-subtle",
        outline:
          "bg-transparent text-text-primary border border-line-subtle hover:bg-bg-glass/60 hover:border-line-strong",
        secondary:
          "bg-bg-glass/70 text-text-primary border border-line-subtle hover:bg-bg-glass hover:border-line-strong",
        danger:
          "bg-status-err/15 text-status-err border border-status-err/30 hover:bg-status-err/25 hover:border-status-err/50",
        chip: "bg-bg-chip/70 text-text-secondary border border-line-subtle hover:text-text-primary hover:border-line-strong",
        glass:
          "surface-glass text-text-primary hover:text-text-primary hover:border-line-strong",
      },
      size: {
        sm: "h-8 px-3 text-xs",
        md: "h-9 px-3.5 text-sm",
        lg: "h-10 px-4 text-sm",
        icon: "h-9 w-9 p-0",
        "icon-sm": "h-7 w-7 p-0",
      },
    },
    defaultVariants: { variant: "secondary", size: "md" },
  },
);

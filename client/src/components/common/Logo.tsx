import { cn } from "@/lib/cn";

export function Logo({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "relative grid h-7 w-7 place-items-center rounded-md",
        "bg-accent-gradient text-text-inverse shadow-[0_4px_16px_-4px_hsl(var(--accent)/0.55)]",
        className,
      )}
      aria-hidden
    >
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
        <circle cx="8" cy="3" r="1.7" fill="currentColor" />
        <circle cx="3" cy="13" r="1.7" fill="currentColor" />
        <circle cx="13" cy="13" r="1.7" fill="currentColor" />
        <path
          d="M8 4.5L4 11.5M8 4.5L12 11.5M4.5 13H11.5"
          stroke="currentColor"
          strokeWidth="1.2"
          strokeLinecap="round"
        />
      </svg>
      <span className="absolute inset-0 rounded-md ring-1 ring-inset ring-white/15" />
    </div>
  );
}

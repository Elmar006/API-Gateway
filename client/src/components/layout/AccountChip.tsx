import { ChevronDown, LogOut } from "lucide-react";

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/Popover";
import { Separator } from "@/components/ui/Separator";
import { useAuthStore } from "@/store/authStore";

const initials = (name: string) =>
  name
    .split(/\s+|[._-]/)
    .map((p) => p[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase();

export function AccountChip() {
  const claims = useAuthStore((s) => s.claims);
  const signOut = useAuthStore((s) => s.signOut);

  const name = claims?.username ?? claims?.sub ?? "Unknown";
  const role = claims?.role ?? "user";

  return (
    <Popover>
      <PopoverTrigger className="titlebar-no-drag group inline-flex items-center gap-2 rounded-md border border-transparent px-1.5 py-1 hover:border-line-subtle hover:bg-bg-glass/60 focus:outline-none focus-visible:shadow-ring">
        <div
          aria-hidden
          className="grid h-8 w-8 place-items-center rounded-full bg-accent-gradient text-text-inverse text-xs font-semibold shadow-[0_4px_16px_-6px_hsl(var(--accent)/0.55)]"
        >
          {initials(name)}
        </div>
        <div className="hidden text-left leading-tight md:block">
          <div className="text-sm font-medium text-text-primary">{name}</div>
          <div className="text-[11px] text-text-muted">{role}</div>
        </div>
        <ChevronDown className="hidden h-3.5 w-3.5 text-text-muted md:block" />
      </PopoverTrigger>
      <PopoverContent align="end" className="w-60">
        <div className="px-2 pb-2 pt-1">
          <div className="text-sm font-medium">{name}</div>
          <div className="text-xs text-text-muted">role: {role}</div>
          {claims?.exp && (
            <div className="mt-1 text-[10px] uppercase tracking-wider text-text-muted">
              session expires {new Date(claims.exp * 1000).toLocaleTimeString()}
            </div>
          )}
        </div>
        <Separator className="my-1" />
        <button
          onClick={signOut}
          className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-status-err/90 hover:bg-status-err/15"
        >
          <LogOut className="h-4 w-4" /> Sign out
        </button>
      </PopoverContent>
    </Popover>
  );
}

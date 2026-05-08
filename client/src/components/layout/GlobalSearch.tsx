import { Search } from "lucide-react";
import { useEffect, useMemo, useRef } from "react";

import { Input } from "@/components/ui/Input";
import { KeyHint } from "@/components/ui/KeyHint";
import { useRoutes } from "@/hooks/useRoutes";
import { useHotkey, platformModSymbol } from "@/hooks/useHotkey";
import { deriveServices } from "@/lib/derive";
import { cn } from "@/lib/cn";
import { useAppStore } from "@/store/appStore";
import { useGraphStore } from "@/store/graphStore";

interface SearchHit {
  key: string;
  label: string;
  sub: string;
  kind: "route" | "service";
  /** value to pass into `select` */
  selectId: string;
}

export function GlobalSearch() {
  const inputRef = useRef<HTMLInputElement>(null);
  const search = useAppStore((s) => s.search);
  const setSearch = useAppStore((s) => s.setSearch);
  const select = useGraphStore((s) => s.select);
  const setSection = useAppStore((s) => s.setSection);
  const { data: routes = [] } = useRoutes();
  const services = useMemo(() => deriveServices(routes), [routes]);

  useHotkey("mod+k", () => inputRef.current?.focus());

  const results = useMemo<SearchHit[]>(() => {
    if (!search.trim()) return [];
    const q = search.toLowerCase();
    const routeHits: SearchHit[] = routes
      .filter(
        (r) =>
          r.path_pattern.toLowerCase().includes(q) ||
          r.target_url.toLowerCase().includes(q) ||
          r.method.toLowerCase().includes(q),
      )
      .slice(0, 6)
      .map((r) => ({
        key: `r:${r.id}`,
        label: `${r.method} ${r.path_pattern}`,
        sub: r.target_url,
        kind: "route",
        selectId: String(r.id),
      }));
    const serviceHits: SearchHit[] = services
      .filter((s) => s.host.toLowerCase().includes(q))
      .slice(0, 4)
      .map((s) => ({
        key: `s:${s.id}`,
        label: s.host,
        sub: `${s.scheme}:// • ${s.routeCount} route${s.routeCount === 1 ? "" : "s"}`,
        kind: "service",
        selectId: s.id,
      }));
    return [...routeHits, ...serviceHits].slice(0, 8);
  }, [search, routes, services]);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") inputRef.current?.blur();
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  return (
    <div className="titlebar-no-drag relative w-full max-w-[520px]">
      <div className="group relative flex h-9 items-center">
        <Search className="pointer-events-none absolute left-3 h-4 w-4 text-text-muted" />
        <Input
          ref={inputRef}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search routes or upstream hosts…"
          className={cn(
            "h-9 w-full border-line-subtle/80 bg-bg-glass/40 pl-9 pr-16 text-sm",
            "shadow-[inset_0_1px_0_0_hsl(var(--glass-highlight)/0.04)]",
            "group-hover:bg-bg-glass/60",
          )}
        />
        <KeyHint className="absolute right-2 top-1/2 -translate-y-1/2">{platformModSymbol}K</KeyHint>
      </div>

      {results.length > 0 && (
        <div className="absolute left-0 right-0 top-[calc(100%+6px)] rounded-md border border-line-subtle bg-bg-panel/95 shadow-floating backdrop-blur-md z-50 animate-fade-in">
          <ul className="py-1.5">
            {results.map((r) => (
              <li key={r.key}>
                <button
                  className="flex w-full items-center justify-between gap-3 px-3 py-2 text-left text-sm hover:bg-bg-glass/70"
                  onClick={() => {
                    select({ id: r.selectId, kind: r.kind });
                    setSection("routes");
                    setSearch("");
                  }}
                >
                  <div className="flex flex-col">
                    <span className="truncate text-text-primary">{r.label}</span>
                    <span className="truncate text-xs text-text-muted">{r.sub}</span>
                  </div>
                  <span className="text-step text-text-muted">{r.kind}</span>
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

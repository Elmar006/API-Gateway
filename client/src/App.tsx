import { TooltipProvider } from "@/components/ui/Tooltip";
import { Inspector } from "@/components/panels/Inspector";
import { Sidebar } from "@/components/layout/Sidebar";
import { TitleBar } from "@/components/layout/TitleBar";
import { Workspace } from "@/components/layout/Workspace";
import { useAppStore } from "@/store/appStore";
import { useGraphStore } from "@/store/graphStore";
import { useHotkey } from "@/hooks/useHotkey";

export function App() {
  const inspectorOpen = useAppStore((s) => s.inspectorOpen);
  const setInspectorOpen = useAppStore((s) => s.setInspectorOpen);
  const select = useGraphStore((s) => s.select);

  useHotkey("esc", () => select(null));
  useHotkey("mod+b", () => setInspectorOpen(!inspectorOpen));

  return (
    <TooltipProvider delayDuration={250}>
      <div className="relative flex h-screen w-screen flex-col overflow-hidden bg-bg-base">
        <div className="pointer-events-none absolute inset-0 bg-hero-gradient opacity-90" />
        <TitleBar />
        <main className="relative z-10 flex flex-1 overflow-hidden">
          <Sidebar />
          <Workspace />
          {inspectorOpen && <Inspector />}
        </main>
      </div>
    </TooltipProvider>
  );
}

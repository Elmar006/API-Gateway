import { lazy, Suspense } from "react";

import { Canvas } from "@/components/graph/Canvas";
import { KpiCards } from "@/components/layout/KpiCards";
import { WorkspaceHeader } from "@/components/layout/WorkspaceHeader";
import { useAppStore, type SectionId } from "@/store/appStore";

// Lazy-load heavy data views; canvas + KPI keep the initial bundle lean.
const LogsView = lazy(() =>
  import("@/components/views/LogsView").then((m) => ({ default: m.LogsView })),
);
const MetricsView = lazy(() =>
  import("@/components/views/MetricsView").then((m) => ({ default: m.MetricsView })),
);
const AlertsView = lazy(() =>
  import("@/components/views/AlertsView").then((m) => ({ default: m.AlertsView })),
);
const SettingsView = lazy(() =>
  import("@/components/views/SettingsView").then((m) => ({ default: m.SettingsView })),
);
const EnvironmentsView = lazy(() =>
  import("@/components/views/EnvironmentsView").then((m) => ({ default: m.EnvironmentsView })),
);
const ClustersView = lazy(() =>
  import("@/components/views/ClustersView").then((m) => ({ default: m.ClustersView })),
);
const MiddlewaresView = lazy(() =>
  import("@/components/views/MiddlewaresView").then((m) => ({ default: m.MiddlewaresView })),
);
const UsersView = lazy(() =>
  import("@/components/views/UsersView").then((m) => ({ default: m.UsersView })),
);
const TokensView = lazy(() =>
  import("@/components/views/TokensView").then((m) => ({ default: m.TokensView })),
);
const TracesView = lazy(() =>
  import("@/components/views/TracesView").then((m) => ({ default: m.TracesView })),
);

const CANVAS_SECTIONS: ReadonlySet<SectionId> = new Set([
  "dashboard",
  "routes",
  "services",
  "gateways",
]);

export function Workspace() {
  const section = useAppStore((s) => s.section);

  let body: React.ReactNode;
  if (CANVAS_SECTIONS.has(section)) {
    body = (
      <>
        <KpiCards />
        <div className="relative min-h-0 flex-1">
          <Canvas />
        </div>
      </>
    );
  } else if (section === "logs") {
    body = <DataPane><LogsView /></DataPane>;
  } else if (section === "metrics") {
    body = <DataPane><MetricsView /></DataPane>;
  } else if (section === "alerts") {
    body = <DataPane><AlertsView /></DataPane>;
  } else if (section === "settings") {
    body = <DataPane><SettingsView /></DataPane>;
  } else if (section === "environments") {
    body = <DataPane><EnvironmentsView /></DataPane>;
  } else if (section === "clusters") {
    body = <DataPane><ClustersView /></DataPane>;
  } else if (section === "middlewares") {
    body = <DataPane><MiddlewaresView /></DataPane>;
  } else if (section === "users") {
    body = <DataPane><UsersView /></DataPane>;
  } else if (section === "tokens") {
    body = <DataPane><TokensView /></DataPane>;
  } else if (section === "traces") {
    body = <DataPane><TracesView /></DataPane>;
  } else {
    body = <DataPane><SettingsView /></DataPane>;
  }

  return (
    <section className="flex flex-1 flex-col gap-4 px-6 py-5">
      <WorkspaceHeader />
      {body}
    </section>
  );
}

function DataPane({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative min-h-0 flex-1">
      <Suspense
        fallback={
          <div className="surface-glass grid h-full place-items-center rounded-xl text-sm text-text-muted">
            Loading…
          </div>
        }
      >
        {children}
      </Suspense>
    </div>
  );
}

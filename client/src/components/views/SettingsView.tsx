import { Globe2, KeyRound, ShieldAlert } from "lucide-react";

import { API_BASE_URL } from "@/api";
import { useGatewayHealth } from "@/hooks/useHealth";
import { useAuthStore } from "@/store/authStore";

export function SettingsView() {
  const claims = useAuthStore((s) => s.claims);
  const health = useGatewayHealth();

  return (
    <div className="surface-glass flex h-full flex-col gap-4 overflow-auto rounded-xl p-4">
      <header className="flex items-center gap-2 text-sm text-text-secondary">
        <ShieldAlert className="h-4 w-4 text-accent-soft" />
        <span className="font-medium text-text-primary">Workspace settings</span>
      </header>

      <section className="surface-glass rounded-md p-3">
        <div className="mb-2 flex items-center gap-2 text-step text-text-muted">
          <Globe2 className="h-3.5 w-3.5" /> Connection
        </div>
        <ul className="space-y-1.5 text-xs">
          <Row label="API base URL" value={API_BASE_URL} mono />
          <Row label="Health check" value={health.ok ? "ok" : health.live ? "starting" : "unreachable"} />
        </ul>
      </section>

      <section className="surface-glass rounded-md p-3">
        <div className="mb-2 flex items-center gap-2 text-step text-text-muted">
          <KeyRound className="h-3.5 w-3.5" /> Session
        </div>
        <ul className="space-y-1.5 text-xs">
          <Row label="User" value={claims?.username ?? claims?.sub ?? "—"} />
          <Row label="Role" value={claims?.role ?? "—"} />
          <Row
            label="Expires"
            value={claims?.exp ? new Date(claims.exp * 1000).toLocaleString() : "—"}
          />
        </ul>
      </section>

      <p className="text-xs text-text-muted">
        Authentication tokens are stored in the browser session storage and
        cleared automatically when the tab is closed. Use the account menu to
        sign out manually.
      </p>
    </div>
  );
}

function Row({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <li className="flex items-center justify-between gap-2 rounded-md border border-line-subtle bg-bg-glass/40 px-2.5 py-1.5">
      <span className="text-text-muted">{label}</span>
      <span className={`text-text-primary ${mono ? "font-mono" : ""}`}>{value}</span>
    </li>
  );
}

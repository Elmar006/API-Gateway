import { motion } from "framer-motion";
import { Loader2, ShieldCheck } from "lucide-react";
import { useState, type FormEvent } from "react";

import { ApiError, auth } from "@/api";
import { Logo } from "@/components/common/Logo";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { useAuthStore } from "@/store/authStore";

export function LoginScreen() {
  const setToken = useAuthStore((s) => s.setToken);
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (pending) return;
    setError(null);
    setPending(true);
    try {
      const { token } = await auth.login({ username: username.trim(), password });
      setToken(token);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.status === 401 ? "Invalid username or password" : err.message);
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Unknown error");
      }
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="relative grid h-screen w-screen place-items-center bg-bg-base text-text-primary">
      <div aria-hidden className="pointer-events-none absolute inset-0 bg-hero-gradient opacity-90" />
      <motion.form
        onSubmit={onSubmit}
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.32, ease: "easeOut" }}
        className="surface-panel relative z-10 flex w-[380px] flex-col gap-5 rounded-xl border border-line-subtle p-7 shadow-floating"
      >
        <div className="flex items-center gap-3">
          <Logo />
          <div>
            <div className="text-base font-semibold tracking-tight">RouteFlow</div>
            <div className="text-[11px] uppercase tracking-[0.22em] text-text-muted">
              Control plane sign-in
            </div>
          </div>
        </div>

        <div className="flex flex-col gap-3">
          <label className="flex flex-col gap-1.5">
            <span className="text-[11px] uppercase tracking-wider text-text-muted">Username</span>
            <Input
              autoFocus
              autoComplete="username"
              name="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-[11px] uppercase tracking-wider text-text-muted">Password</span>
            <Input
              autoComplete="current-password"
              name="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </label>
        </div>

        {error && (
          <div
            role="alert"
            className="rounded-md border border-status-err/40 bg-status-err/10 px-3 py-2 text-xs text-status-err"
          >
            {error}
          </div>
        )}

        <Button type="submit" variant="primary" size="md" disabled={pending} className="w-full justify-center">
          {pending ? <Loader2 className="h-4 w-4 animate-spin" /> : <ShieldCheck className="h-4 w-4" />}
          {pending ? "Signing in…" : "Sign in"}
        </Button>

        <div className="text-[11px] leading-relaxed text-text-muted">
          Use the <code className="font-mono text-text-secondary">ADMIN_USERNAME</code> /
          <code className="font-mono text-text-secondary"> ADMIN_PASSWORD</code> credentials
          configured for the gateway. The admin API runs on port <span className="font-mono">9090</span> by default; set
          <code className="font-mono text-text-secondary"> VITE_API_BASE_URL</code> to point elsewhere.
        </div>
      </motion.form>
    </div>
  );
}

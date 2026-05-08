import { Trash2 } from "lucide-react";
import { useEffect, useRef } from "react";
import { Controller, useForm } from "react-hook-form";

import { Button } from "@/components/ui/Button";
import { Card, CardSection, CardTitle } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { Switch } from "@/components/ui/Switch";
import { useDeleteRoute, useUpdateRoute } from "@/hooks/useRoutes";
import { parseTargetURL } from "@/lib/derive";
import { useGraphStore } from "@/store/graphStore";
import { HTTP_METHODS_DISPLAY } from "@/types/domain";
import type { HTTPMethod, Route, RouteInput } from "@/types/domain";

interface FormShape {
  method: HTTPMethod;
  path_pattern: string;
  target_url: string;
  is_active: boolean;
  priority: number;
  rate_limit: number;
  require_auth: boolean;
}

function toInput(form: FormShape): RouteInput {
  return {
    method: form.method,
    path_pattern: form.path_pattern.trim(),
    target_url: form.target_url.trim(),
    is_active: !!form.is_active,
    priority: Number.isFinite(form.priority) ? Math.max(0, Math.floor(form.priority)) : 0,
    rate_limit: Number.isFinite(form.rate_limit) ? Math.max(0, Math.floor(form.rate_limit)) : 0,
    require_auth: !!form.require_auth,
  };
}

function inputsEqual(a: RouteInput, b: RouteInput): boolean {
  return (
    a.method === b.method &&
    a.path_pattern === b.path_pattern &&
    a.target_url === b.target_url &&
    a.is_active === b.is_active &&
    a.priority === b.priority &&
    a.rate_limit === b.rate_limit &&
    a.require_auth === b.require_auth
  );
}

export function RouteInspector({ route }: { route: Route }) {
  const updateMutation = useUpdateRoute();
  const deleteMutation = useDeleteRoute();
  const select = useGraphStore((s) => s.select);

  // Stable ref to .mutate so the auto-save effect's dep array does not flip
  // every time the mutation enters/leaves a pending state. Without this, the
  // effect cleanup would cancel an in-flight debounce timer whenever a save
  // started (isPending: true) or completed (isPending: false), and any edit
  // made while another save was in-flight could be silently dropped.
  const mutateRef = useRef(updateMutation.mutate);
  useEffect(() => {
    mutateRef.current = updateMutation.mutate;
  }, [updateMutation.mutate]);

  const { register, watch, control, reset, getValues } = useForm<FormShape>({
    defaultValues: {
      method: route.method,
      path_pattern: route.path_pattern,
      target_url: route.target_url,
      is_active: route.is_active,
      priority: route.priority,
      rate_limit: route.rate_limit,
      require_auth: route.require_auth,
    },
  });

  const lastSent = useRef<RouteInput>({
    method: route.method,
    path_pattern: route.path_pattern,
    target_url: route.target_url,
    is_active: route.is_active,
    priority: route.priority,
    rate_limit: route.rate_limit,
    require_auth: route.require_auth,
  });

  // Reset form when navigating to a different route.
  const routeKey = route.id;
  useEffect(() => {
    reset({
      method: route.method,
      path_pattern: route.path_pattern,
      target_url: route.target_url,
      is_active: route.is_active,
      priority: route.priority,
      rate_limit: route.rate_limit,
      require_auth: route.require_auth,
    });
    lastSent.current = {
      method: route.method,
      path_pattern: route.path_pattern,
      target_url: route.target_url,
      is_active: route.is_active,
      priority: route.priority,
      rate_limit: route.rate_limit,
      require_auth: route.require_auth,
    };
    // Only when the *identity* of the route changes — not when its server fields refresh.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [routeKey]);

  // Debounced auto-save: 350 ms after the user stops typing.
  // Deps intentionally exclude `updateMutation` (which changes reference on
  // every state transition) — we read .mutate via a stable ref instead.
  useEffect(() => {
    let timer: ReturnType<typeof setTimeout> | null = null;
    const sub = watch(() => {
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => {
        const candidate = toInput(getValues());
        if (!candidate.path_pattern.startsWith("/")) return;
        const parsed = parseTargetURL(candidate.target_url);
        if (!parsed) return;
        if (inputsEqual(candidate, lastSent.current)) return;
        lastSent.current = candidate;
        mutateRef.current({ id: routeKey, input: candidate });
      }, 350);
    });
    return () => {
      sub.unsubscribe();
      if (timer) clearTimeout(timer);
    };
  }, [watch, getValues, routeKey]);

  return (
    <div className="space-y-3">
      <Card>
        <CardSection className="space-y-3">
          <CardTitle>Routing</CardTitle>
          <div className="grid grid-cols-[80px_1fr] items-center gap-x-3 gap-y-2">
            <label className="text-xs text-text-muted">Method</label>
            <Controller
              control={control}
              name="method"
              render={({ field }) => (
                <select
                  className="h-9 w-full rounded-md border border-line-subtle bg-bg-glass/60 px-2.5 text-sm text-text-primary"
                  value={field.value}
                  onChange={field.onChange}
                >
                  {HTTP_METHODS_DISPLAY.map((m) => (
                    <option key={m} value={m} className="bg-bg-panel">
                      {m}
                    </option>
                  ))}
                </select>
              )}
            />
            <label className="text-xs text-text-muted">Pattern</label>
            <Input
              className="font-mono"
              placeholder="/api/users/**"
              {...register("path_pattern")}
            />
            <label className="text-xs text-text-muted">Target</label>
            <Input
              className="font-mono"
              placeholder="http://users.svc.cluster.local:8080"
              {...register("target_url")}
            />
          </div>
        </CardSection>

        <CardSection className="space-y-3">
          <CardTitle>Behaviour</CardTitle>
          <div className="grid grid-cols-2 gap-2">
            <KV label="Priority">
              <Input type="number" min={0} {...register("priority", { valueAsNumber: true })} />
            </KV>
            <KV label="Rate limit (rps)">
              <Input type="number" min={0} {...register("rate_limit", { valueAsNumber: true })} />
            </KV>
          </div>
          <div className="flex items-center justify-between rounded-md border border-line-subtle bg-bg-glass/40 px-3 py-2">
            <div>
              <div className="text-sm">Require auth</div>
              <div className="text-[11px] text-text-muted">Reject unauthenticated requests at the edge</div>
            </div>
            <Controller
              control={control}
              name="require_auth"
              render={({ field }) => (
                <Switch checked={field.value} onCheckedChange={field.onChange} />
              )}
            />
          </div>
          <div className="flex items-center justify-between rounded-md border border-line-subtle bg-bg-glass/40 px-3 py-2">
            <div>
              <div className="text-sm">Active</div>
              <div className="text-[11px] text-text-muted">If off, the route is bypassed</div>
            </div>
            <Controller
              control={control}
              name="is_active"
              render={({ field }) => (
                <Switch checked={field.value} onCheckedChange={field.onChange} />
              )}
            />
          </div>
        </CardSection>

        <CardSection className="space-y-2">
          <CardTitle>Danger zone</CardTitle>
          <Button
            variant="ghost"
            size="md"
            onClick={() => {
              if (deleteMutation.isPending) return;
              const ok = window.confirm(
                `Delete route ${route.method} ${route.path_pattern}? This cannot be undone.`,
              );
              if (!ok) return;
              deleteMutation.mutate(route.id, {
                onSuccess: () => select(null),
              });
            }}
            className="w-full justify-center border border-status-err/30 text-status-err hover:bg-status-err/10"
          >
            <Trash2 className="h-4 w-4" /> Delete route
          </Button>
        </CardSection>
      </Card>
    </div>
  );
}

function KV({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-[11px] uppercase tracking-wider text-text-muted">{label}</span>
      {children}
    </label>
  );
}

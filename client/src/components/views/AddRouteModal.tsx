import * as Dialog from "@radix-ui/react-dialog";
import { Plus, X } from "lucide-react";
import { useEffect, useState } from "react";
import { Controller, useForm } from "react-hook-form";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Switch } from "@/components/ui/Switch";
import { useCreateRoute } from "@/hooks/useRoutes";
import { parseTargetURL } from "@/lib/derive";
import { HTTP_METHODS_DISPLAY } from "@/types/domain";
import type { HTTPMethod, RouteInput } from "@/types/domain";

const DEFAULT: RouteInput = {
  method: "GET",
  path_pattern: "/api/example",
  target_url: "http://example.svc.cluster.local:8080",
  is_active: true,
  priority: 0,
  rate_limit: 0,
  require_auth: false,
};

export function AddRouteButton() {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button variant="primary" size="md" onClick={() => setOpen(true)}>
        <Plus className="h-4 w-4" /> Add route
      </Button>
      <AddRouteModal open={open} onOpenChange={setOpen} />
    </>
  );
}

interface ModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function AddRouteModal({ open, onOpenChange }: ModalProps) {
  const create = useCreateRoute();
  const { register, handleSubmit, reset, control, watch, formState } =
    useForm<RouteInput>({ defaultValues: DEFAULT });

  useEffect(() => {
    if (!open) reset(DEFAULT);
  }, [open, reset]);

  const targetUrl = watch("target_url");
  const path = watch("path_pattern");
  const targetParsed = parseTargetURL(targetUrl);
  const pathOk = typeof path === "string" && path.startsWith("/");

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-bg-deep/70 backdrop-blur-sm" />
        <Dialog.Content className="fixed left-1/2 top-1/2 z-50 w-[min(560px,90vw)] -translate-x-1/2 -translate-y-1/2 rounded-xl border border-line-subtle bg-bg-panel p-5 shadow-floating">
          <div className="mb-3 flex items-start justify-between gap-3">
            <div>
              <Dialog.Title className="text-lg font-semibold tracking-tight text-text-primary">
                New route
              </Dialog.Title>
              <Dialog.Description className="mt-0.5 text-xs text-text-muted">
                The route is registered immediately and traffic begins flowing if it is active.
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <Button variant="ghost" size="icon-sm" aria-label="Close">
                <X className="h-3.5 w-3.5" />
              </Button>
            </Dialog.Close>
          </div>

          <form
            onSubmit={handleSubmit(async (values) => {
              if (!pathOk || !targetParsed) return;
              await create.mutateAsync({
                ...values,
                priority: Number(values.priority) || 0,
                rate_limit: Number(values.rate_limit) || 0,
              });
              onOpenChange(false);
            })}
            className="space-y-3"
          >
            <div className="grid grid-cols-[100px_1fr] items-center gap-x-3 gap-y-2">
              <label className="text-xs text-text-muted">Method</label>
              <Controller
                control={control}
                name="method"
                render={({ field }) => (
                  <select
                    className="h-9 w-full rounded-md border border-line-subtle bg-bg-glass/60 px-2.5 text-sm text-text-primary"
                    value={field.value}
                    onChange={(e) => field.onChange(e.target.value as HTTPMethod)}
                  >
                    {HTTP_METHODS_DISPLAY.map((m) => (
                      <option key={m} value={m} className="bg-bg-panel">
                        {m}
                      </option>
                    ))}
                  </select>
                )}
              />
              <label className="text-xs text-text-muted">Path pattern</label>
              <Input
                className="font-mono"
                placeholder="/api/users/**"
                {...register("path_pattern", { required: true })}
              />
              <label className="text-xs text-text-muted">Target URL</label>
              <Input
                className="font-mono"
                placeholder="http://users.svc.cluster.local:8080"
                {...register("target_url", { required: true })}
              />
              <label className="text-xs text-text-muted">Priority</label>
              <Input type="number" min={0} {...register("priority", { valueAsNumber: true })} />
              <label className="text-xs text-text-muted">Rate limit (rps)</label>
              <Input type="number" min={0} {...register("rate_limit", { valueAsNumber: true })} />
            </div>

            <div className="grid grid-cols-2 gap-2">
              <label className="flex items-center justify-between rounded-md border border-line-subtle bg-bg-glass/40 px-3 py-2 text-sm">
                <span>Active</span>
                <Controller
                  control={control}
                  name="is_active"
                  render={({ field }) => (
                    <Switch checked={!!field.value} onCheckedChange={field.onChange} />
                  )}
                />
              </label>
              <label className="flex items-center justify-between rounded-md border border-line-subtle bg-bg-glass/40 px-3 py-2 text-sm">
                <span>Require auth</span>
                <Controller
                  control={control}
                  name="require_auth"
                  render={({ field }) => (
                    <Switch checked={!!field.value} onCheckedChange={field.onChange} />
                  )}
                />
              </label>
            </div>

            {!pathOk && path && (
              <div className="rounded-md border border-status-warn/40 bg-status-warn/10 px-3 py-2 text-xs text-status-warn">
                Path pattern must start with “/”.
              </div>
            )}
            {!targetParsed && targetUrl && (
              <div className="rounded-md border border-status-warn/40 bg-status-warn/10 px-3 py-2 text-xs text-status-warn">
                Target URL must be a valid http(s) URL.
              </div>
            )}
            {create.error && (
              <div className="rounded-md border border-status-err/40 bg-status-err/10 px-3 py-2 text-xs text-status-err">
                {(create.error as Error).message}
              </div>
            )}

            <div className="flex justify-end gap-2 pt-1">
              <Dialog.Close asChild>
                <Button type="button" variant="ghost" size="md">
                  Cancel
                </Button>
              </Dialog.Close>
              <Button
                type="submit"
                variant="primary"
                size="md"
                disabled={create.isPending || !pathOk || !targetParsed || !formState.isValid}
              >
                {create.isPending ? "Creating…" : "Create route"}
              </Button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

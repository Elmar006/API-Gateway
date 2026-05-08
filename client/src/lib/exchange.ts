import { parse as parseYaml, stringify as stringifyYaml } from "yaml";

import { isTauri, pickOpenFile, pickSaveFile, readTextFile, writeTextFile } from "@/lib/tauri";
import type { Route, RouteInput } from "@/types/domain";

const filters = [
  { name: "Route configs", extensions: ["json", "yaml", "yml"] },
  { name: "All files", extensions: ["*"] },
];

/**
 * Import a YAML/JSON route file. Returns RouteInput[] (no id/timestamps) so the
 * caller can decide which routes to upsert via the backend.
 */
export async function importRoutes(): Promise<RouteInput[] | null> {
  if (!isTauri()) {
    return browserImport();
  }
  const path = await pickOpenFile({ filters });
  if (!path) return null;
  const text = await readTextFile(path);
  return decode(text, path);
}

/**
 * Export the current set of backend routes to YAML/JSON. We strip timestamps
 * (they're server-owned), keep id for round-tripping/auditing.
 */
export async function exportRoutes(routes: readonly Route[]): Promise<boolean> {
  const exportable = routes.map(stripExportShape);
  if (!isTauri()) {
    browserExport(exportable);
    return true;
  }
  const path = await pickSaveFile({
    defaultPath: "routeflow-routes.yaml",
    filters,
  });
  if (!path) return false;
  const yaml = path.endsWith(".json")
    ? JSON.stringify({ routes: exportable }, null, 2)
    : stringifyYaml({ routes: exportable });
  await writeTextFile(path, yaml);
  return true;
}

interface ExportShape extends RouteInput {
  id?: number;
}

function stripExportShape(r: Route): ExportShape {
  return {
    id: r.id,
    method: r.method,
    path_pattern: r.path_pattern,
    target_url: r.target_url,
    is_active: r.is_active,
    priority: r.priority,
    rate_limit: r.rate_limit,
    require_auth: r.require_auth,
  };
}

function coerceInput(raw: unknown): RouteInput | null {
  if (!raw || typeof raw !== "object") return null;
  const r = raw as Record<string, unknown>;
  const method = typeof r.method === "string" ? r.method.toUpperCase() : null;
  const path_pattern = typeof r.path_pattern === "string" ? r.path_pattern : null;
  const target_url = typeof r.target_url === "string" ? r.target_url : null;
  if (!method || !path_pattern || !target_url) return null;
  return {
    method: method as RouteInput["method"],
    path_pattern,
    target_url,
    is_active: typeof r.is_active === "boolean" ? r.is_active : true,
    priority: typeof r.priority === "number" ? r.priority : 0,
    rate_limit: typeof r.rate_limit === "number" ? r.rate_limit : 0,
    require_auth: typeof r.require_auth === "boolean" ? r.require_auth : false,
  };
}

function decode(text: string, path: string): RouteInput[] {
  const isJson = path.toLowerCase().endsWith(".json");
  const data = isJson ? JSON.parse(text) : parseYaml(text);
  const list = Array.isArray(data?.routes)
    ? (data.routes as unknown[])
    : Array.isArray(data)
      ? (data as unknown[])
      : null;
  if (!list) throw new Error("Unsupported route configuration format");
  const result: RouteInput[] = [];
  for (const item of list) {
    const r = coerceInput(item);
    if (r) result.push(r);
  }
  return result;
}

function browserExport(routes: readonly ExportShape[]) {
  const blob = new Blob([stringifyYaml({ routes })], { type: "text/yaml" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "routeflow-routes.yaml";
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

function browserImport(): Promise<RouteInput[] | null> {
  return new Promise((resolve) => {
    const input = document.createElement("input");
    input.type = "file";
    input.accept = ".json,.yaml,.yml";
    input.addEventListener("change", () => {
      const file = input.files?.[0];
      if (!file) return resolve(null);
      const reader = new FileReader();
      reader.onload = () => {
        try {
          const text = String(reader.result ?? "");
          resolve(decode(text, file.name));
        } catch {
          resolve(null);
        }
      };
      reader.readAsText(file);
    });
    input.click();
  });
}

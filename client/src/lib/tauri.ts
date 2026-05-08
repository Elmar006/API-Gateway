/**
 * Thin abstraction over the Tauri runtime. All calls degrade gracefully when
 * the app is being rendered in a plain browser (e.g. `npm run dev`) so the UI
 * keeps working during web-only development.
 */

type Maybe<T> = T | undefined;

let cachedTauri: Maybe<boolean>;

export function isTauri(): boolean {
  if (cachedTauri !== undefined) return cachedTauri;
  cachedTauri = typeof window !== "undefined" && "__TAURI_INTERNALS__" in window;
  return cachedTauri;
}

export async function notify(title: string, body?: string): Promise<void> {
  if (!isTauri()) {
    if (typeof window !== "undefined" && "Notification" in window) {
      try {
        if (Notification.permission === "default") {
          await Notification.requestPermission();
        }
        if (Notification.permission === "granted") {
          new Notification(title, { body });
        }
      } catch {
        // ignore – notifications are best-effort.
      }
    }
    return;
  }
  const { sendNotification, isPermissionGranted, requestPermission } = await import(
    "@tauri-apps/plugin-notification"
  );
  let granted = await isPermissionGranted();
  if (!granted) {
    const permission = await requestPermission();
    granted = permission === "granted";
  }
  if (granted) sendNotification({ title, body });
}

export async function pickOpenFile(opts?: {
  filters?: { name: string; extensions: string[] }[];
}): Promise<string | null> {
  if (!isTauri()) return browserPickFile();
  const { open } = await import("@tauri-apps/plugin-dialog");
  const result = await open({ multiple: false, directory: false, filters: opts?.filters });
  return typeof result === "string" ? result : null;
}

export async function pickSaveFile(opts: {
  defaultPath?: string;
  filters?: { name: string; extensions: string[] }[];
}): Promise<string | null> {
  if (!isTauri()) return null;
  const { save } = await import("@tauri-apps/plugin-dialog");
  const result = await save(opts);
  return typeof result === "string" ? result : null;
}

export async function readTextFile(path: string): Promise<string> {
  if (!isTauri()) throw new Error("readTextFile is only available inside Tauri");
  const { readTextFile: read } = await import("@tauri-apps/plugin-fs");
  return read(path);
}

export async function writeTextFile(path: string, contents: string): Promise<void> {
  if (!isTauri()) throw new Error("writeTextFile is only available inside Tauri");
  const { writeTextFile: write } = await import("@tauri-apps/plugin-fs");
  await write(path, contents);
}

export async function appWindow() {
  if (!isTauri()) return null;
  const mod = await import("@tauri-apps/api/window");
  return mod.getCurrentWindow();
}

async function browserPickFile(): Promise<string | null> {
  return new Promise((resolve) => {
    const input = document.createElement("input");
    input.type = "file";
    input.accept = ".json,.yaml,.yml";
    input.addEventListener("change", () => {
      const file = input.files?.[0];
      if (!file) return resolve(null);
      const reader = new FileReader();
      reader.onload = () => {
        const data = reader.result;
        if (typeof data !== "string") return resolve(null);
        // Stash the raw payload on a global cache keyed by name so callers can
        // round-trip via readTextFile-like helpers in browser dev mode.
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (window as any).__routeflow_browser_files__ = {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          ...((window as any).__routeflow_browser_files__ ?? {}),
          [file.name]: data,
        };
        resolve(file.name);
      };
      reader.readAsText(file);
    });
    input.click();
  });
}

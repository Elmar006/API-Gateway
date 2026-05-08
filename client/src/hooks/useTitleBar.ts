import { useEffect, useState } from "react";
import { appWindow, isTauri } from "@/lib/tauri";

export interface TitleBarController {
  isMaximized: boolean;
  minimize: () => Promise<void>;
  toggleMaximize: () => Promise<void>;
  close: () => Promise<void>;
  startDrag: () => Promise<void>;
  isTauri: boolean;
}

export function useTitleBar(): TitleBarController {
  const [maximized, setMaximized] = useState(false);

  useEffect(() => {
    if (!isTauri()) return;
    let unlisten: (() => void) | undefined;
    (async () => {
      const w = await appWindow();
      if (!w) return;
      setMaximized(await w.isMaximized());
      unlisten = await w.onResized(async () => {
        setMaximized(await w.isMaximized());
      });
    })();
    return () => unlisten?.();
  }, []);

  return {
    isMaximized: maximized,
    isTauri: isTauri(),
    minimize: async () => {
      const w = await appWindow();
      await w?.minimize();
    },
    toggleMaximize: async () => {
      const w = await appWindow();
      await w?.toggleMaximize();
    },
    close: async () => {
      const w = await appWindow();
      await w?.close();
    },
    startDrag: async () => {
      const w = await appWindow();
      await w?.startDragging();
    },
  };
}

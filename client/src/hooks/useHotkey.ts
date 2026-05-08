import { useEffect } from "react";

interface HotkeyOptions {
  preventDefault?: boolean;
  enabled?: boolean;
}

const isMac =
  typeof navigator !== "undefined" && /(Macintosh|MacIntel|MacPPC|Mac68K)/i.test(navigator.platform);

/**
 * Registers a keyboard shortcut. The combo string accepts any number of
 * modifier prefixes ("mod" maps to ⌘ on macOS and Ctrl elsewhere) and a
 * single key name (case-insensitive). Examples: "mod+k", "shift+/", "esc".
 */
export function useHotkey(combo: string, handler: (event: KeyboardEvent) => void, opts: HotkeyOptions = {}) {
  const { preventDefault = true, enabled = true } = opts;

  useEffect(() => {
    if (!enabled) return;
    const tokens = combo.toLowerCase().split("+").map((t) => t.trim());
    const key = tokens.pop()!;
    const wantsMod = tokens.includes("mod");
    const wantsShift = tokens.includes("shift");
    const wantsAlt = tokens.includes("alt") || tokens.includes("opt");

    function listener(event: KeyboardEvent) {
      const eventKey = event.key.toLowerCase();
      if (eventKey !== key) return;
      if (wantsMod) {
        const modPressed = isMac ? event.metaKey : event.ctrlKey;
        if (!modPressed) return;
      } else if (event.metaKey || event.ctrlKey) {
        return;
      }
      if (wantsShift !== event.shiftKey) return;
      if (wantsAlt !== event.altKey) return;

      if (preventDefault) event.preventDefault();
      handler(event);
    }
    window.addEventListener("keydown", listener);
    return () => window.removeEventListener("keydown", listener);
  }, [combo, handler, preventDefault, enabled]);
}

export const platformModSymbol = isMac ? "⌘" : "Ctrl";

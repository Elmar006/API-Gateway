import type { Config } from "tailwindcss";
import animate from "tailwindcss-animate";

const config: Config = {
  darkMode: ["class"],
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: { "2xl": "1400px" },
    },
    extend: {
      colors: {
        bg: {
          base: "hsl(var(--bg-base) / <alpha-value>)",
          deep: "hsl(var(--bg-deep) / <alpha-value>)",
          panel: "hsl(var(--bg-panel) / <alpha-value>)",
          glass: "hsl(var(--bg-glass) / <alpha-value>)",
          chip: "hsl(var(--bg-chip) / <alpha-value>)",
        },
        line: {
          subtle: "hsl(var(--line-subtle) / <alpha-value>)",
          strong: "hsl(var(--line-strong) / <alpha-value>)",
          glow: "hsl(var(--line-glow) / <alpha-value>)",
        },
        text: {
          primary: "hsl(var(--text-primary) / <alpha-value>)",
          secondary: "hsl(var(--text-secondary) / <alpha-value>)",
          muted: "hsl(var(--text-muted) / <alpha-value>)",
          inverse: "hsl(var(--text-inverse) / <alpha-value>)",
        },
        accent: {
          DEFAULT: "hsl(var(--accent) / <alpha-value>)",
          muted: "hsl(var(--accent-muted) / <alpha-value>)",
          soft: "hsl(var(--accent-soft) / <alpha-value>)",
          ring: "hsl(var(--accent-ring) / <alpha-value>)",
        },
        status: {
          ok: "hsl(var(--status-ok) / <alpha-value>)",
          warn: "hsl(var(--status-warn) / <alpha-value>)",
          err: "hsl(var(--status-err) / <alpha-value>)",
          info: "hsl(var(--status-info) / <alpha-value>)",
        },
      },
      borderRadius: {
        lg: "14px",
        md: "10px",
        sm: "6px",
      },
      fontFamily: {
        sans: [
          "InterVariable",
          "Inter",
          "ui-sans-serif",
          "system-ui",
          "-apple-system",
          "Segoe UI",
          "Roboto",
          "Helvetica Neue",
          "sans-serif",
        ],
        mono: [
          "JetBrains Mono",
          "ui-monospace",
          "SFMono-Regular",
          "Menlo",
          "Consolas",
          "monospace",
        ],
      },
      boxShadow: {
        glass:
          "inset 0 1px 0 0 hsl(var(--glass-highlight) / 0.6), inset 0 -1px 0 0 hsl(var(--glass-shadow) / 0.4), 0 8px 24px -8px hsl(var(--accent-ring) / 0.18)",
        ring: "0 0 0 1px hsl(var(--accent-ring) / 0.55), 0 0 24px -4px hsl(var(--accent) / 0.45)",
        floating: "0 24px 48px -16px rgb(8 4 32 / 0.6), 0 8px 16px -8px rgb(8 4 32 / 0.4)",
        edge: "0 1px 0 0 hsl(var(--line-subtle) / 1)",
      },
      backgroundImage: {
        "panel-gradient":
          "linear-gradient(180deg, hsl(var(--bg-panel) / 0.86) 0%, hsl(var(--bg-panel) / 0.7) 100%)",
        "glass-gradient":
          "linear-gradient(160deg, hsl(var(--bg-glass) / 0.76) 0%, hsl(var(--bg-glass) / 0.45) 100%)",
        "accent-gradient":
          "linear-gradient(135deg, hsl(var(--accent) / 1) 0%, hsl(var(--accent-soft) / 1) 100%)",
        "hero-gradient":
          "radial-gradient(120% 80% at 0% 0%, hsl(263 80% 24% / 0.55) 0%, transparent 60%), radial-gradient(80% 80% at 100% 0%, hsl(284 70% 32% / 0.4) 0%, transparent 60%), radial-gradient(60% 60% at 50% 100%, hsl(252 60% 18% / 0.6) 0%, transparent 60%)",
        "grid-fade":
          "linear-gradient(180deg, hsl(var(--bg-base) / 0) 0%, hsl(var(--bg-base) / 1) 80%)",
      },
      keyframes: {
        "fade-in": {
          from: { opacity: "0" },
          to: { opacity: "1" },
        },
        "scale-in": {
          from: { opacity: "0", transform: "scale(0.96)" },
          to: { opacity: "1", transform: "scale(1)" },
        },
        "pulse-soft": {
          "0%, 100%": { opacity: "0.55" },
          "50%": { opacity: "1" },
        },
        "edge-flow": {
          to: { strokeDashoffset: "-32" },
        },
      },
      animation: {
        "fade-in": "fade-in 220ms ease-out",
        "scale-in": "scale-in 180ms ease-out",
        "pulse-soft": "pulse-soft 2.4s ease-in-out infinite",
        "edge-flow": "edge-flow 1.6s linear infinite",
      },
    },
  },
  plugins: [animate],
};

export default config;

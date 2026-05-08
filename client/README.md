# RouteFlow

> Premium dark-glass desktop control plane for API gateway routing.

RouteFlow is a native desktop application that gives platform engineers a fast, opinionated UI for inspecting and managing the topology of their API gateway: traffic sources, gateways, routes, downstream services, middlewares, rate limits, retries, timeouts, and more. It ships as a single installable program for **macOS, Windows, and Linux** powered by [Tauri 2](https://v2.tauri.app/).

The application is intentionally **not** a browser SPA. It runs as a real OS window with custom title-bar chrome, native menus, native file dialogs, native notifications, and native filesystem access for importing and exporting routing configuration in JSON or YAML.

---

## Tech stack

| Layer            | Choice                                              |
| ---------------- | --------------------------------------------------- |
| Native shell     | **Tauri 2** (Rust)                                  |
| UI               | **React 18 + TypeScript** (Vite)                    |
| Styling          | **Tailwind CSS** + custom design tokens             |
| Primitives       | **shadcn/ui-style** components on **Radix UI**      |
| Graph            | **@xyflow/react** (React Flow v12) with custom nodes/edges |
| State            | **Zustand** (slice stores)                          |
| Forms            | **React Hook Form**                                 |
| Animation        | **Framer Motion**                                   |
| Icons            | **Lucide**                                          |
| File / dialog    | `@tauri-apps/plugin-fs`, `@tauri-apps/plugin-dialog`|
| Notifications    | `@tauri-apps/plugin-notification`                   |

Every visual primitive is hand-built for this product. The dark-purple glass system, gradients, glow, mirror-glint, status palette, and typography hierarchy are all defined in `src/styles/globals.css` and exposed as CSS variables consumed by Tailwind.

---

## Project layout

```
RouteFlow/
├── src/                       # React UI
│   ├── App.tsx                # Top-level shell
│   ├── main.tsx               # Entry point
│   ├── components/
│   │   ├── ui/                # shadcn-style primitives (Button, Tabs, Tooltip, …)
│   │   ├── layout/            # TitleBar, Sidebar, Workspace, header, KPI cards
│   │   ├── common/            # Logo, AccountChip
│   │   ├── graph/             # Canvas, edges, buildGraph, GraphToolbar
│   │   ├── nodes/             # Custom glass nodes (Source, Gateway, Route, Service)
│   │   └── panels/            # Inspector + per-entity inspectors
│   ├── store/                 # Zustand stores (app, graph)
│   ├── hooks/                 # useHotkey, useTitleBar
│   ├── lib/                   # cn, format, tauri bridge, exchange (import/export)
│   ├── data/                  # Realistic mock dataset
│   ├── types/                 # Shared domain types
│   └── styles/                # Tailwind + theme tokens
├── src-tauri/                 # Rust shell
│   ├── src/                   # main.rs, lib.rs, commands
│   ├── icons/                 # App icons (auto-generated)
│   ├── capabilities/          # Tauri 2 capability manifests
│   ├── Cargo.toml
│   ├── build.rs
│   └── tauri.conf.json
├── index.html
├── vite.config.ts
├── tailwind.config.ts
├── tsconfig.json
├── eslint.config.js
└── package.json
```

---

## Prerequisites

* **Node.js ≥ 20** and **npm ≥ 10**
* **Rust stable ≥ 1.85** (`rustup install stable && rustup default stable`)
* On **Linux** you also need:
  ```bash
  sudo apt install pkg-config libwebkit2gtk-4.1-dev libgtk-3-dev \
      libayatana-appindicator3-dev librsvg2-dev libssl-dev libsoup-3.0-dev
  ```
* On **macOS** Xcode Command Line Tools (`xcode-select --install`).
* On **Windows** Microsoft C++ Build Tools and the WebView2 runtime (preinstalled on Windows 10/11).

---

## Getting started

```bash
# Install JS dependencies
npm install

# Run as a desktop app (recommended — Tauri shell + Vite HMR)
npm run tauri:dev

# Or run only the web frontend in a browser tab for quick iteration
npm run dev          # http://localhost:1420

# Lint and format
npm run lint
npm run format
```

The first `npm run tauri:dev` will compile the Rust side. Allow a few minutes the first time — subsequent runs are incremental and fast.

### Production build

```bash
npm run tauri:build
```

This produces installable bundles per platform under `src-tauri/target/release/bundle/`:

| OS      | Output                                                                  |
| ------- | ----------------------------------------------------------------------- |
| macOS   | `.app` bundle and `.dmg` installer                                       |
| Windows | `.exe` (NSIS) and `.msi` installer                                       |
| Linux   | `.deb`, `.rpm`, and `.AppImage`                                          |

---

## Application surface

The app opens with a custom borderless window split into three columns:

1. **Sidebar** — grouped navigation (Manage / Observe / Settings) with a status-aware environment switcher pinned to the bottom.
2. **Workspace** — KPI cards plus the interactive routing canvas.
3. **Inspector** — context-sensitive editor for the selected node.

### Title bar

A custom header replaces the OS title bar:

* draggable region for window movement
* product mark and name
* global search (services, routes, nodes) with `⌘K` / `Ctrl+K` hotkey hint
* notifications badge
* account chip (avatar, name, role)
* native window controls (minimize, maximize, close) wired to Tauri APIs

### Routing canvas

Powered by React Flow v12 with all visuals replaced. Layout is column-based:

```
sources   →   gateway   →   routes   →   services
```

* nodes are custom **glass cards** with status dots, icons, monospaced technical lines, and a soft top mirror glint;
* the gateway is visually larger, accent-coloured, and pulses softly;
* edges are smooth, accent-glowing curves; the selected/hovered chain is emphasised, related edges are softly highlighted, unrelated ones fade;
* the floating top toolbar exposes **fit / zoom in / zoom out / center / reset**;
* the bottom-left controls and bottom-right minimap are fully restyled to match the dark glass palette;
* a search box in the toolbar filters nodes live (`buildGraph` rebuilds and dims non-matches).

### Inspector

The right panel changes content based on selection:

* **Routes** — full editable form (path, method, target, load-balancing strategy, sticky sessions, timeouts, retries, health checks, rate limits, headers, policies). Powered by React Hook Form so the structure is ready to be wired to a real backend.
* **Services** — deployment metadata (cluster, region, protocol, version, pod count, owner team, saturation).
* **Gateways** — flavour, region, uptime, endpoints.
* **Sources** — origin metadata.

A second-tab **Rules** view lists matchers per route, **Middleware** lists attached middlewares with enable state, and **Metrics** renders custom SVG sparklines for RPS, error rate, and p95 latency.

### Tauri integration

The frontend talks to the OS through a thin bridge in `src/lib/tauri.ts`:

* `pickOpenFile` / `pickSaveFile` — native dialogs, fall back to a download blob in browser dev mode;
* `readTextFile` / `writeTextFile` — filesystem access through the `fs` plugin;
* `notify(title, body)` — system notifications;
* `appWindow()` — drag, minimize, toggle-maximize, close.

`src/lib/exchange.ts` uses these primitives to **import** and **export** routing configuration in JSON or YAML. The `Import` / `Export` buttons in the workspace header are fully wired to native dialogs, parse the file with `yaml`, and surface the result via a system notification.

---

## Keyboard shortcuts

| Shortcut       | Action                              |
| -------------- | ----------------------------------- |
| `⌘K` / `Ctrl+K`| Focus the global search             |
| `Esc`          | Clear the current canvas selection  |
| `⌘B` / `Ctrl+B`| Toggle the inspector panel          |

---

## Available scripts

| Script               | Purpose                                         |
| -------------------- | ----------------------------------------------- |
| `npm run dev`        | Run the Vite frontend in a browser tab          |
| `npm run build`      | Type-check + bundle the frontend                |
| `npm run preview`    | Preview the production frontend bundle          |
| `npm run lint`       | Run ESLint over the codebase                    |
| `npm run format`     | Run Prettier on `src/**/*.{ts,tsx,css}`         |
| `npm run tauri`      | Pass-through to the Tauri CLI                   |
| `npm run tauri:dev`  | Run the desktop application in development mode |
| `npm run tauri:build`| Produce installable platform bundles            |

---

## Notes on the data layer

All entities currently live in `src/data/mockData.ts` with realistic, opinionated values (paths, RPS, retries, healthcheck thresholds, owner teams, etc.). The forms and stores are already shaped against the entity types in `src/types/domain.ts`, so wiring the app to a real gateway control plane API is a matter of replacing the static dataset with calls to your backend (the existing Go API gateway in this repository tree, for example).

---

## License

Proprietary. Internal use only.

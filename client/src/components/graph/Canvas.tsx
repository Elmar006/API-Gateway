import {
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  ReactFlow,
  ReactFlowProvider,
  useReactFlow,
  type EdgeTypes,
  type Node,
  type NodeChange,
  type NodeTypes,
} from "@xyflow/react";
import { useCallback, useEffect, useLayoutEffect, useMemo, useRef } from "react";

import { GatewayNode } from "@/components/nodes/GatewayNode";
import { RouteNode } from "@/components/nodes/RouteNode";
import { ServiceNode } from "@/components/nodes/ServiceNode";
import { useGatewayHealth } from "@/hooks/useHealth";
import { useRoutes } from "@/hooks/useRoutes";
import { defaultGatewayInfo, deriveServices } from "@/lib/derive";
import { useGraphStore } from "@/store/graphStore";
import type { InspectorTarget } from "@/types/domain";

import {
  GATEWAY_NODE_ID,
  ROUTE_NODE_ID_PREFIX,
  buildFlowGraph,
  buildSimulationGraph,
} from "./buildGraph";
import { FlowEdge } from "./edges";
import { createSimulation, type SimulationHandle } from "./forces";
import { GraphToolbar } from "./GraphToolbar";

const nodeTypes: NodeTypes = {
  gateway: GatewayNode,
  route: RouteNode,
  service: ServiceNode,
};

const edgeTypes: EdgeTypes = {
  flow: FlowEdge,
};

const PIN_STORAGE_KEY = "routeflow.graph.pins.v1";

interface PinMap {
  [id: string]: { x: number; y: number };
}

function loadPins(): PinMap {
  try {
    const raw = sessionStorage.getItem(PIN_STORAGE_KEY);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as PinMap;
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch {
    return {};
  }
}

function savePins(pins: PinMap) {
  try {
    sessionStorage.setItem(PIN_STORAGE_KEY, JSON.stringify(pins));
  } catch {
    /* sessionStorage unavailable (e.g. private mode) — skip */
  }
}

function targetForNodeId(id: string): InspectorTarget | null {
  if (id === GATEWAY_NODE_ID) return { id, kind: "gateway" };
  if (id.startsWith(ROUTE_NODE_ID_PREFIX)) {
    return { id: id.slice(ROUTE_NODE_ID_PREFIX.length), kind: "route" };
  }
  if (id.startsWith("svc:")) return { id, kind: "service" };
  return null;
}

function CanvasInner() {
  const { fitView } = useReactFlow();

  const { data: routes = [], isLoading, error } = useRoutes();

  const filter = useGraphStore((s) => s.graphFilter);
  const selection = useGraphStore((s) => s.selection);
  const hoveredId = useGraphStore((s) => s.hoveredId);
  const select = useGraphStore((s) => s.select);
  const hover = useGraphStore((s) => s.hover);

  const gatewayHealth = useGatewayHealth();
  const gateway = useMemo(
    () => defaultGatewayInfo(gatewayHealth.ok ? "active" : "degraded"),
    [gatewayHealth.ok],
  );

  const filteredRoutes = useMemo(() => {
    if (!filter.trim()) return routes;
    const q = filter.trim().toLowerCase();
    return routes.filter(
      (r) =>
        r.path_pattern.toLowerCase().includes(q) ||
        r.target_url.toLowerCase().includes(q) ||
        r.method.toLowerCase().includes(q),
    );
  }, [routes, filter]);
  const filteredServices = useMemo(() => deriveServices(filteredRoutes), [filteredRoutes]);

  const onSelectNode = useCallback(
    (id: string) => {
      const target = targetForNodeId(id);
      if (target) select(target);
    },
    [select],
  );

  const selectedNodeId = useMemo(() => {
    if (!selection) return null;
    if (selection.kind === "route") return `${ROUTE_NODE_ID_PREFIX}${selection.id}`;
    return selection.id;
  }, [selection]);

  // ── React Flow nodes/edges ─────────────────────────────────────────────
  const flowGraph = useMemo(
    () =>
      buildFlowGraph(
        gateway,
        filteredRoutes,
        filteredServices,
        { onSelect: onSelectNode, onHover: hover },
        selectedNodeId,
        hoveredId,
      ),
    [gateway, filteredRoutes, filteredServices, onSelectNode, hover, selectedNodeId, hoveredId],
  );

  // ── d3-force simulation ────────────────────────────────────────────────
  const containerRef = useRef<HTMLDivElement | null>(null);
  const simRef = useRef<SimulationHandle | null>(null);
  const nodesRef = useRef<Map<string, Node>>(new Map());
  const flowApiRef = useRef<{ setNodes: (n: Node[]) => void } | null>(null);
  const pinsRef = useRef<PinMap>(loadPins());
  const draggingRef = useRef<string | null>(null);
  const rafRef = useRef<number | null>(null);
  const pendingPositionsRef = useRef<Map<string, { x: number; y: number }> | null>(null);

  // Re-projection tick: turns position cache → React Flow node updates.
  // We coalesce d3 ticks with rAF so React only sees one update per frame
  // even if d3 settles faster than the display refresh rate.
  const applyPositions = useCallback((positions: Map<string, { x: number; y: number }>) => {
    pendingPositionsRef.current = positions;
    if (rafRef.current !== null) return;
    rafRef.current = requestAnimationFrame(() => {
      rafRef.current = null;
      const latest = pendingPositionsRef.current;
      if (!latest) return;
      pendingPositionsRef.current = null;

      const map = nodesRef.current;
      const next: Node[] = [];
      map.forEach((node, id) => {
        const pos = latest.get(id);
        if (!pos) {
          next.push(node);
          return;
        }
        next.push({ ...node, position: { x: pos.x, y: pos.y } });
      });
      nodesRef.current = new Map(next.map((n) => [n.id, n]));
      flowApiRef.current?.setNodes(next);
    });
  }, []);

  // (Re)build the simulation whenever the graph topology changes.
  useLayoutEffect(() => {
    const el = containerRef.current;
    const width = el?.clientWidth ?? 1200;
    const height = el?.clientHeight ?? 800;

    const { nodes: simNodes, links: simLinks } = buildSimulationGraph(
      filteredRoutes,
      filteredServices,
    );

    // Seed from saved pins so layout doesn't jump on reload.
    for (const n of simNodes) {
      const saved = pinsRef.current[n.id];
      if (saved) {
        n.x = saved.x;
        n.y = saved.y;
        n.fx = saved.x;
        n.fy = saved.y;
      }
    }

    if (!simRef.current) {
      simRef.current = createSimulation(
        simNodes,
        simLinks,
        { width, height },
        applyPositions,
      );
    } else {
      simRef.current.update(simNodes, simLinks);
    }

    return () => {
      // We intentionally keep the sim across rerenders — only stop on unmount.
    };
  }, [filteredRoutes, filteredServices, applyPositions]);

  // Permanent cleanup on unmount.
  useEffect(
    () => () => {
      simRef.current?.stop();
      if (rafRef.current !== null) cancelAnimationFrame(rafRef.current);
    },
    [],
  );

  // Keep nodesRef in sync with flowGraph; React Flow renders flowGraph.nodes,
  // and `applyPositions` mutates positions inside that map per tick.
  useLayoutEffect(() => {
    const map = new Map<string, Node>();
    for (const node of flowGraph.nodes) {
      // Merge with existing live position to avoid snapping back to (0,0).
      const previous = nodesRef.current.get(node.id);
      const merged = previous
        ? { ...node, position: previous.position }
        : node;
      map.set(node.id, merged);
    }
    nodesRef.current = map;
  }, [flowGraph.nodes]);

  // Drag handler — pins to cursor while dragging, releases on drop.
  const onNodesChange = useCallback(
    (changes: NodeChange[]) => {
      for (const change of changes) {
        if (change.type === "position" && "id" in change) {
          if (change.dragging) {
            draggingRef.current = change.id;
            if (change.position) simRef.current?.drag(change.id, change.position.x, change.position.y);
          } else if (draggingRef.current === change.id) {
            // Drag ended. Persist the pin location so it survives reloads.
            draggingRef.current = null;
            const node = nodesRef.current.get(change.id);
            if (node && change.position) {
              const pinned = { x: change.position.x, y: change.position.y };
              pinsRef.current[change.id] = pinned;
              savePins(pinsRef.current);
              simRef.current?.pin(change.id, pinned.x, pinned.y);
            }
          }
        } else if (change.type === "remove") {
          delete pinsRef.current[change.id];
          savePins(pinsRef.current);
        }
      }
    },
    [],
  );

  // Double-click a node = release its pin (let physics take over again).
  const onNodeDoubleClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      delete pinsRef.current[node.id];
      savePins(pinsRef.current);
      simRef.current?.release(node.id);
      simRef.current?.reheat(0.4);
    },
    [],
  );

  // After data loads, if the layout is fresh, fit view once it has settled.
  const didFit = useRef(false);
  useEffect(() => {
    if (didFit.current) return;
    if (filteredRoutes.length === 0) return;
    const t = window.setTimeout(() => {
      didFit.current = true;
      fitView({ padding: 0.2, duration: 320 });
    }, 280);
    return () => window.clearTimeout(t);
  }, [filteredRoutes.length, fitView]);

  return (
    <div
      ref={containerRef}
      className="relative h-full w-full overflow-hidden rounded-xl border border-line-subtle/80 bg-bg-deep/60 grid-bg shadow-floating"
    >
      <ReactFlow
        nodes={flowGraph.nodes}
        edges={flowGraph.edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        onNodesChange={onNodesChange}
        onNodeDoubleClick={onNodeDoubleClick}
        onPaneClick={() => select(null)}
        onInit={(instance) => {
          flowApiRef.current = { setNodes: instance.setNodes };
        }}
        fitView={false}
        minZoom={0.3}
        maxZoom={1.8}
        proOptions={{ hideAttribution: true }}
        nodesDraggable
        nodesConnectable={false}
        elementsSelectable
        zoomOnScroll
        panOnScroll
        nodesFocusable={false}
      >
        <Background variant={BackgroundVariant.Dots} gap={26} size={1} color="currentColor" />
        <MiniMap
          pannable
          zoomable
          bgColor="transparent"
          maskColor="hsl(252 50% 6% / 0.55)"
          maskStrokeColor="hsl(263 88% 66% / 0.55)"
          maskStrokeWidth={1}
          nodeColor={() => "hsl(263 88% 66% / 0.85)"}
          nodeStrokeColor={() => "hsl(274 88% 80% / 0.6)"}
          nodeStrokeWidth={0.6}
          nodeBorderRadius={3}
          position="bottom-right"
          style={{ width: 168, height: 96 }}
        />
        <Controls
          showInteractive={false}
          position="bottom-left"
          className="!gap-1 [&>button]:!h-8 [&>button]:!w-8 [&>button]:!rounded-md [&>button]:!border [&>button]:!border-line-subtle [&>button]:!bg-bg-glass/70 [&>button]:!text-text-secondary hover:[&>button]:!text-text-primary [&>button:hover]:!border-line-strong [&>button>svg]:!fill-current"
        />
      </ReactFlow>

      <GraphToolbar />

      {(isLoading || error || routes.length === 0) && (
        <div className="pointer-events-none absolute inset-0 grid place-items-center">
          <div className="surface-glass pointer-events-auto rounded-md px-4 py-2 text-xs text-text-secondary">
            {isLoading
              ? "Loading routes…"
              : error
                ? `Failed to load routes: ${(error as Error).message}`
                : "No routes yet — create one in the Routes section to see it here."}
          </div>
        </div>
      )}
    </div>
  );
}

export function Canvas() {
  return (
    <ReactFlowProvider>
      <CanvasInner />
    </ReactFlowProvider>
  );
}

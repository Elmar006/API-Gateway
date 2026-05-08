/**
 * Builds React Flow nodes & edges + d3-force topology from real backend data.
 *
 * - One static gateway node (the admin endpoint we're connected to).
 * - One node per `Route` returned by `/admin/routes`.
 * - One node per unique `target_url` host (DerivedService).
 * - Edges: gateway → route → service.
 *
 * No layout coordinates here — positions are owned by the d3-force simulation.
 */

import type { Edge, Node } from "@xyflow/react";

import type { GatewayNodeData } from "@/components/nodes/GatewayNode";
import type { RouteNodeData } from "@/components/nodes/RouteNode";
import type { ServiceNodeData } from "@/components/nodes/ServiceNode";
import { serviceIdFor } from "@/lib/derive";
import type {
  DerivedService,
  GatewayNodeInfo,
  Route,
} from "@/types/domain";

import type { SimLink, SimNode } from "./forces";
import type { FlowEdgeData } from "./edges";

export const ROUTE_NODE_ID_PREFIX = "route:";
export const GATEWAY_NODE_ID = "gateway";

export function routeNodeId(routeId: number): string {
  return `${ROUTE_NODE_ID_PREFIX}${routeId}`;
}

export interface GraphCallbacks {
  onSelect: (id: string) => void;
  onHover: (id: string | null) => void;
}

export interface FlowGraph {
  nodes: Node[];
  edges: Edge[];
}

export interface SimulationGraph {
  nodes: SimNode[];
  links: SimLink[];
}

/**
 * Build React Flow nodes/edges. `selectedId` and `hoveredId` are inlined into
 * `data` only for Flow's diffing; the visual highlight is also done via CSS so
 * memoised node components don't have to re-render on hover.
 */
export function buildFlowGraph(
  gateway: GatewayNodeInfo,
  routes: readonly Route[],
  services: readonly DerivedService[],
  callbacks: GraphCallbacks,
  selectedId: string | null,
  hoveredId: string | null,
): FlowGraph {
  const nodes: Node[] = [];
  const edges: Edge[] = [];

  const isHighlighted = (id: string): boolean => {
    if (!hoveredId) return false;
    if (hoveredId === id) return true;
    if (hoveredId === GATEWAY_NODE_ID) {
      // Gateway hover lights the whole graph.
      return true;
    }
    return false;
  };

  // Gateway node
  nodes.push({
    id: GATEWAY_NODE_ID,
    type: "gateway",
    position: { x: 0, y: 0 },
    draggable: true,
    selectable: true,
    data: {
      gateway,
      selected: selectedId === GATEWAY_NODE_ID,
      highlighted: isHighlighted(GATEWAY_NODE_ID),
      onSelect: () => callbacks.onSelect(GATEWAY_NODE_ID),
      onHover: callbacks.onHover,
    } satisfies GatewayNodeData,
  });

  // Route nodes
  for (const route of routes) {
    const id = routeNodeId(route.id);
    nodes.push({
      id,
      type: "route",
      position: { x: 0, y: 0 },
      draggable: true,
      selectable: true,
      data: {
        route,
        selected: selectedId === id,
        highlighted: isHighlighted(id),
        onSelect: () => callbacks.onSelect(id),
        onHover: callbacks.onHover,
      } satisfies RouteNodeData,
    });

    edges.push({
      id: `e:gw->${id}`,
      source: GATEWAY_NODE_ID,
      target: id,
      type: "flow",
      data: {
        highlighted: hoveredId === id || hoveredId === GATEWAY_NODE_ID,
        emphasised: selectedId === id || selectedId === GATEWAY_NODE_ID,
      } satisfies FlowEdgeData,
    });
  }

  // Service nodes
  for (const service of services) {
    nodes.push({
      id: service.id,
      type: "service",
      position: { x: 0, y: 0 },
      draggable: true,
      selectable: true,
      data: {
        service,
        selected: selectedId === service.id,
        highlighted: isHighlighted(service.id),
        onSelect: () => callbacks.onSelect(service.id),
        onHover: callbacks.onHover,
      } satisfies ServiceNodeData,
    });
  }

  // Route → service edges (only if the route's target has a parsable host)
  for (const route of routes) {
    const sid = serviceIdFor(route.target_url);
    if (!sid) continue;
    const id = routeNodeId(route.id);
    edges.push({
      id: `e:${id}->${sid}`,
      source: id,
      target: sid,
      type: "flow",
      data: {
        highlighted: hoveredId === id || hoveredId === sid,
        emphasised: selectedId === id || selectedId === sid,
      } satisfies FlowEdgeData,
    });
  }

  return { nodes, edges };
}

/**
 * Build the d3-force graph topology from the same data set. We bias spring
 * lengths and collision radii by node kind so the layout looks deliberate.
 */
export function buildSimulationGraph(
  routes: readonly Route[],
  services: readonly DerivedService[],
): SimulationGraph {
  const nodes: SimNode[] = [];
  const links: SimLink[] = [];

  nodes.push({ id: GATEWAY_NODE_ID, weight: 4, radius: 90 });

  for (const route of routes) {
    const id = routeNodeId(route.id);
    nodes.push({ id, weight: 1, radius: 56 });
    links.push({ source: GATEWAY_NODE_ID, target: id, distance: 220 });
  }

  for (const service of services) {
    nodes.push({ id: service.id, weight: 2, radius: 64 });
  }

  for (const route of routes) {
    const sid = serviceIdFor(route.target_url);
    if (!sid) continue;
    links.push({ source: routeNodeId(route.id), target: sid, distance: 180 });
  }

  return { nodes, links };
}

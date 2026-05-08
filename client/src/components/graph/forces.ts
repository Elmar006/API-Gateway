/**
 * d3-force simulation tuned for the RouteFlow graph.
 *
 * Goals:
 *  - Obsidian-style feel: nodes drift toward equilibrium, springs on edges,
 *    repulsion to prevent overlap, smooth drag-to-pin & release.
 *  - Idle = 0 CPU. We stop ticking as soon as the simulation cools down.
 *  - The hot loop never touches React: it mutates `node.x/y` in place and
 *    consumers (React Flow) read positions via a callback per RAF tick.
 */

import {
  forceCenter,
  forceCollide,
  forceLink,
  forceManyBody,
  forceSimulation,
  type Simulation,
  type SimulationNodeDatum,
} from "d3-force";

export interface SimNode extends SimulationNodeDatum {
  id: string;
  /** "weight class" – heavier nodes resist forces more (gateway > service > route). */
  weight?: number;
  /** node radius for collision (px). */
  radius?: number;
}

export interface SimLink {
  source: string;
  target: string;
  /** preferred edge length in px. */
  distance?: number;
}

export interface SimulationOptions {
  width: number;
  height: number;
  centerStrength?: number;
  chargeStrength?: number;
  linkDistance?: number;
  alphaDecay?: number;
  velocityDecay?: number;
}

export interface SimulationHandle {
  /** advance one tick (the simulation calls this internally; exposed for debugging). */
  tick: () => void;
  /** rebuild simulation with new node/link sets, preserving positions where possible. */
  update: (nodes: SimNode[], links: SimLink[]) => void;
  /** kick the system back into motion (useful after structural changes). */
  reheat: (alpha?: number) => void;
  /** pin a node to a position; release with `release(id)`. */
  pin: (id: string, x: number, y: number) => void;
  release: (id: string) => void;
  /** instantaneously move a node (for active drag). */
  drag: (id: string, x: number, y: number) => void;
  /** read positions snapshot. */
  positions: () => Map<string, { x: number; y: number }>;
  /** stop the simulation and free the RAF loop. */
  stop: () => void;
}

interface InternalNode extends SimNode {
  x: number;
  y: number;
}

const DEFAULTS: Required<Omit<SimulationOptions, "width" | "height">> = {
  centerStrength: 0.04,
  chargeStrength: -520,
  linkDistance: 200,
  alphaDecay: 0.025,
  velocityDecay: 0.34,
};

export function createSimulation(
  initialNodes: SimNode[],
  initialLinks: SimLink[],
  opts: SimulationOptions,
  onTick: (positions: Map<string, { x: number; y: number }>) => void,
): SimulationHandle {
  const settings = { ...DEFAULTS, ...opts };
  const nodes: InternalNode[] = initialNodes.map((n) => spreadInitial(n, settings));
  let sim: Simulation<InternalNode, undefined> = configure(nodes, initialLinks, settings);

  const positionsCache = new Map<string, { x: number; y: number }>();
  function syncPositions() {
    positionsCache.clear();
    for (const n of nodes) positionsCache.set(n.id, { x: n.x, y: n.y });
  }

  sim.on("tick", () => {
    syncPositions();
    onTick(positionsCache);
  });

  function findNode(id: string): InternalNode | undefined {
    return nodes.find((n) => n.id === id);
  }

  return {
    tick: () => sim.tick(),
    reheat: (alpha = 0.6) => sim.alpha(alpha).restart(),
    update: (newNodes, newLinks) => {
      // Preserve x/y from existing nodes by id so the layout doesn't jump.
      const map = new Map(nodes.map((n) => [n.id, n] as const));
      nodes.length = 0;
      for (const n of newNodes) {
        const existing = map.get(n.id);
        nodes.push(
          existing
            ? { ...existing, ...n, x: existing.x, y: existing.y }
            : spreadInitial(n, settings),
        );
      }
      sim.stop();
      sim = configure(nodes, newLinks, settings);
      sim.on("tick", () => {
        syncPositions();
        onTick(positionsCache);
      });
      sim.alpha(0.85).restart();
      syncPositions();
    },
    pin: (id, x, y) => {
      const n = findNode(id);
      if (!n) return;
      n.fx = x;
      n.fy = y;
      sim.alpha(0.4).restart();
    },
    release: (id) => {
      const n = findNode(id);
      if (!n) return;
      n.fx = null;
      n.fy = null;
      sim.alpha(0.3).restart();
    },
    drag: (id, x, y) => {
      const n = findNode(id);
      if (!n) return;
      n.fx = x;
      n.fy = y;
      // Mid-drag: keep simulation lukewarm so neighbours respond.
      if (sim.alpha() < 0.2) sim.alpha(0.2);
      sim.restart();
    },
    positions: () => positionsCache,
    stop: () => sim.stop(),
  };
}

function configure(
  nodes: InternalNode[],
  links: SimLink[],
  s: Required<Omit<SimulationOptions, "width" | "height">> & { width: number; height: number },
): Simulation<InternalNode, undefined> {
  const sim = forceSimulation<InternalNode>(nodes)
    .alphaDecay(s.alphaDecay)
    .velocityDecay(s.velocityDecay)
    .force(
      "link",
      forceLink<InternalNode, SimLink>(links.map((l) => ({ ...l })))
        .id((d) => d.id)
        .distance((l) => l.distance ?? s.linkDistance)
        .strength(0.55),
    )
    .force("charge", forceManyBody<InternalNode>().strength(s.chargeStrength).distanceMax(900))
    .force("center", forceCenter(s.width / 2, s.height / 2).strength(s.centerStrength))
    .force(
      "collide",
      forceCollide<InternalNode>()
        .radius((d) => (d.radius ?? 64) + 6)
        .iterations(2)
        .strength(0.85),
    );
  sim.alpha(1).restart();
  return sim;
}

function spreadInitial(
  node: SimNode,
  s: { width: number; height: number },
): InternalNode {
  // If a position is provided, respect it. Otherwise scatter inside the viewport.
  return {
    ...node,
    x: node.x ?? randomCenter(s.width),
    y: node.y ?? randomCenter(s.height),
  };
}

function randomCenter(extent: number): number {
  return extent / 2 + (Math.random() - 0.5) * Math.min(360, extent * 0.5);
}

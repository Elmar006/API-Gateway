import { create } from "zustand";

import type { InspectorTarget } from "@/types/domain";

/**
 * Pure UI state for the graph view: what is selected, hovered, filtered.
 *
 * Real data (routes / services / metrics) lives in TanStack Query — this
 * store only owns transient state that doesn't need a server round-trip.
 */
interface GraphState {
  selection: InspectorTarget | null;
  hoveredId: string | null;
  graphFilter: string;

  setGraphFilter: (value: string) => void;
  select: (target: InspectorTarget | null) => void;
  hover: (id: string | null) => void;
}

export const useGraphStore = create<GraphState>((set) => ({
  selection: null,
  hoveredId: null,
  graphFilter: "",

  setGraphFilter: (graphFilter) => set({ graphFilter }),
  select: (selection) => set({ selection }),
  hover: (hoveredId) => set({ hoveredId }),
}));

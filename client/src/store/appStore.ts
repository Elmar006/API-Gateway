import { create } from "zustand";

export type SectionId =
  | "dashboard"
  | "services"
  | "routes"
  | "gateways"
  | "clusters"
  | "middlewares"
  | "metrics"
  | "logs"
  | "traces"
  | "alerts"
  | "environments"
  | "users"
  | "tokens"
  | "settings";

interface AppState {
  section: SectionId;
  setSection: (section: SectionId) => void;

  search: string;
  setSearch: (value: string) => void;
  searchOpen: boolean;
  openSearch: () => void;
  closeSearch: () => void;

  inspectorOpen: boolean;
  setInspectorOpen: (open: boolean) => void;

  /** Currently focused admin user id (used by the Tokens view to scope queries). */
  selectedUserId: number | null;
  setSelectedUser: (id: number | null) => void;
}

export const useAppStore = create<AppState>((set) => ({
  section: "routes",
  setSection: (section) => set({ section }),

  search: "",
  setSearch: (value) => set({ search: value }),
  searchOpen: false,
  openSearch: () => set({ searchOpen: true }),
  closeSearch: () => set({ searchOpen: false }),

  inspectorOpen: true,
  setInspectorOpen: (inspectorOpen) => set({ inspectorOpen }),

  selectedUserId: null,
  setSelectedUser: (selectedUserId) => set({ selectedUserId }),
}));

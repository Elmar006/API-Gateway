import { create } from "zustand";

import { decodeJwtClaims, isJwtExpired, setTokenProvider, type JwtClaims } from "@/api";

const STORAGE_KEY = "routeflow.auth.token";

function readStoredToken(): string | null {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    if (isJwtExpired(raw)) {
      sessionStorage.removeItem(STORAGE_KEY);
      return null;
    }
    return raw;
  } catch {
    return null;
  }
}

function writeStoredToken(token: string | null): void {
  try {
    if (token) sessionStorage.setItem(STORAGE_KEY, token);
    else sessionStorage.removeItem(STORAGE_KEY);
  } catch {
    /* ignore – session storage may be unavailable in some embeddings */
  }
}

export interface AuthState {
  token: string | null;
  claims: JwtClaims | null;
  setToken: (token: string | null) => void;
  signOut: () => void;
}

export const useAuthStore = create<AuthState>((set) => {
  const initial = readStoredToken();
  return {
    token: initial,
    claims: initial ? decodeJwtClaims(initial) : null,
    setToken: (token) => {
      writeStoredToken(token);
      set({ token, claims: token ? decodeJwtClaims(token) : null });
    },
    signOut: () => {
      writeStoredToken(null);
      set({ token: null, claims: null });
    },
  };
});

// Wire the API client to read the token straight from the store so every
// outbound request gets a fresh value without a re-render dance.
setTokenProvider(() => useAuthStore.getState().token);

/** Convenience selector for components that only need an "is logged in" bit. */
export const selectIsAuthenticated = (s: AuthState): boolean => !!s.token;

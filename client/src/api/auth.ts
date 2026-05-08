import { http } from "./client";
import type { LoginRequest, LoginResponse } from "./types";

export const auth = {
  login: (creds: LoginRequest, signal?: AbortSignal) =>
    http.post<LoginResponse>("/login", creds, { anonymous: true, signal }),
};

export interface JwtClaims {
  sub?: string;
  username?: string;
  role?: string;
  exp?: number;
  iat?: number;
}

/** Parse a JWT body without verifying the signature (display-only). */
export function decodeJwtClaims(token: string): JwtClaims | null {
  const parts = token.split(".");
  if (parts.length < 2) return null;
  try {
    const payload = parts[1]
      .replace(/-/g, "+")
      .replace(/_/g, "/")
      .padEnd(parts[1].length + ((4 - (parts[1].length % 4)) % 4), "=");
    const json = atob(payload);
    return JSON.parse(json) as JwtClaims;
  } catch {
    return null;
  }
}

/** Returns true if the JWT exp claim has passed (with a 5s leeway). */
export function isJwtExpired(token: string): boolean {
  const claims = decodeJwtClaims(token);
  if (!claims?.exp) return false;
  const now = Math.floor(Date.now() / 1000);
  return claims.exp < now - 5;
}

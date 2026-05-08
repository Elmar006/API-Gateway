/**
 * Lean fetch-based HTTP client for the admin API.
 *
 * - JSON-only.
 * - Bearer-token injection from the auth store.
 * - 401 → emits an `apiAuthExpired` window event so the auth gate can react.
 * - Aborts via signal.
 * - Surfaces a typed `ApiError` carrying status + parsed message.
 */

import type { ApiErrorBody } from "./types";

const RAW_BASE = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim();
const DEFAULT_BASE = "/admin";
export const API_BASE_URL = RAW_BASE && RAW_BASE.length > 0 ? stripTrailingSlash(RAW_BASE) : DEFAULT_BASE;

function stripTrailingSlash(s: string): string {
  return s.endsWith("/") ? s.slice(0, -1) : s;
}

export class ApiError extends Error {
  readonly status: number;
  readonly body?: unknown;

  constructor(status: number, message: string, body?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }

  /** Network/CORS/abort failures (no HTTP response). */
  static network(message: string): ApiError {
    return new ApiError(0, message);
  }
}

let tokenProvider: () => string | null = () => null;
export function setTokenProvider(provider: () => string | null): void {
  tokenProvider = provider;
}

const AUTH_EXPIRED_EVENT = "apiAuthExpired";

/** Browser code can listen to this to reset to login on 401s. */
export function onAuthExpired(handler: () => void): () => void {
  const listener = () => handler();
  window.addEventListener(AUTH_EXPIRED_EVENT, listener);
  return () => window.removeEventListener(AUTH_EXPIRED_EVENT, listener);
}

interface RequestOptions extends Omit<RequestInit, "body" | "headers"> {
  body?: unknown;
  headers?: Record<string, string>;
  /** Skip the auth header for unauthenticated calls (e.g. /login). */
  anonymous?: boolean;
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const { body, headers = {}, anonymous, ...rest } = opts;

  const finalHeaders: Record<string, string> = {
    Accept: "application/json",
    ...headers,
  };

  let serializedBody: BodyInit | undefined;
  if (body !== undefined && body !== null) {
    serializedBody = JSON.stringify(body);
    finalHeaders["Content-Type"] = "application/json";
  }

  if (!anonymous) {
    const token = tokenProvider();
    if (token) finalHeaders["Authorization"] = `Bearer ${token}`;
  }

  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...rest,
      headers: finalHeaders,
      body: serializedBody,
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") throw err;
    const msg = err instanceof Error ? err.message : "Network error";
    throw ApiError.network(msg);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const text = await response.text();
  let parsed: unknown = undefined;
  if (text.length > 0) {
    try {
      parsed = JSON.parse(text);
    } catch {
      parsed = text;
    }
  }

  if (!response.ok) {
    if (response.status === 401 && !anonymous) {
      window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT));
    }
    const message =
      (parsed && typeof parsed === "object" && "error" in parsed
        ? (parsed as ApiErrorBody).error
        : undefined) ??
      (typeof parsed === "string" && parsed.length > 0 ? parsed : undefined) ??
      `Request failed with status ${response.status}`;
    throw new ApiError(response.status, message, parsed);
  }

  return parsed as T;
}

export const http = {
  get: <T>(path: string, opts?: RequestOptions) => request<T>(path, { ...opts, method: "GET" }),
  post: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "POST", body }),
  put: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "PUT", body }),
  patch: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "PATCH", body }),
  delete: <T>(path: string, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "DELETE" }),
};

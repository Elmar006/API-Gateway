/**
 * Backend API contract types — 1:1 with the Go entities exposed by the
 * admin server. Do NOT add fields that don't exist on the backend.
 *
 * Source: backend/internal/domain/entity/{route,requestLogs,metrics}.go
 */

export type HTTPMethod =
  | "ALL"
  | "GET"
  | "POST"
  | "PUT"
  | "PATCH"
  | "DELETE"
  | "HEAD"
  | "OPTIONS";

export const HTTP_METHODS: readonly HTTPMethod[] = [
  "ALL",
  "GET",
  "POST",
  "PUT",
  "PATCH",
  "DELETE",
  "HEAD",
  "OPTIONS",
] as const;

/** Mirrors backend `entity.Route` (snake_case, integer id). */
export interface Route {
  id: number;
  method: HTTPMethod;
  path_pattern: string;
  target_url: string;
  is_active: boolean;
  priority: number;
  rate_limit: number;
  require_auth: boolean;
  environment_id?: number | null;
  cluster_id?: number | null;
  created_at: string;
  updated_at: string;
}

/** Payload accepted by POST /routes and PUT /routes/:id. */
export interface RouteInput {
  method: HTTPMethod;
  path_pattern: string;
  target_url: string;
  is_active: boolean;
  priority: number;
  rate_limit: number;
  require_auth: boolean;
  environment_id?: number | null;
  cluster_id?: number | null;
}

/** Mirrors backend `entity.RequestLog`. */
export interface RequestLog {
  id: number;
  request_id: string;
  method: string;
  path: string;
  query?: string | null;
  route_id?: number | null;
  user_id?: number | null;
  client_ip: string;
  status_code: number;
  response_time_ms: number;
  target_url?: string | null;
  trace_id?: string | null;
  created_at: string;
}

export interface LogFilters {
  path?: string;
  status?: number;
  from?: string; // RFC3339
  to?: string;   // RFC3339
  limit?: number;
  offset?: number;
}

export interface LogsPage {
  logs: RequestLog[];
  total: number;
  limit: number;
  offset: number;
}

export type MetricsPeriod = "hour" | "day" | "week" | "month";

export interface RouteStat {
  path: string;
  avg_time_ms: number;
  count: number;
}

/** Mirrors backend `entity.Metrics`. status_counts keys arrive as strings via JSON. */
export interface MetricsSummary {
  total_requests: number;
  rps: number;
  avg_response_time_ms: number;
  status_counts: Record<string, number>;
  slowest_routes: RouteStat[];
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
}

export interface ApiErrorBody {
  error: string;
}

// ---------- Environments ----------

/** Mirrors backend `entity.Environment`. */
export interface Environment {
  id: number;
  name: string;
  base_domain: string;
  color: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface EnvironmentInput {
  name: string;
  base_domain?: string;
  color?: string;
  description?: string;
}

// ---------- Clusters ----------

export type ClusterStrategy = "round_robin" | "weighted" | "least_conn";

export const CLUSTER_STRATEGIES: readonly ClusterStrategy[] = [
  "round_robin",
  "weighted",
  "least_conn",
] as const;

/** Mirrors backend `entity.ClusterTarget`. */
export interface ClusterTarget {
  id: number;
  cluster_id: number;
  url: string;
  weight: number;
  is_healthy: boolean;
  last_check?: string | null;
  created_at: string;
  updated_at: string;
}

/** Mirrors backend `entity.Cluster`. */
export interface Cluster {
  id: number;
  name: string;
  strategy: ClusterStrategy;
  description: string;
  created_at: string;
  updated_at: string;
  targets?: ClusterTarget[];
}

export interface ClusterInput {
  name: string;
  strategy: ClusterStrategy;
  description?: string;
}

export interface ClusterTargetInput {
  url: string;
  weight: number;
}

// ---------- Middlewares ----------

export type MiddlewareKind =
  | "cors"
  | "header_rewrite"
  | "rate_limit_override"
  | "request_id"
  | "strip_prefix";

export const MIDDLEWARE_KINDS: readonly MiddlewareKind[] = [
  "cors",
  "header_rewrite",
  "rate_limit_override",
  "request_id",
  "strip_prefix",
] as const;

/** Mirrors backend `entity.Middleware`. */
export interface Middleware {
  id: number;
  name: string;
  kind: MiddlewareKind;
  config: unknown;
  is_active: boolean;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface MiddlewareInput {
  name: string;
  kind: MiddlewareKind;
  config: unknown;
  is_active: boolean;
  description?: string;
}

export interface AttachMiddlewareInput {
  middleware_id: number;
  sort_order: number;
}

// ---------- Users ----------

export type UserRole = "admin" | "editor" | "viewer";

export const USER_ROLES: readonly UserRole[] = ["admin", "editor", "viewer"] as const;

/** Mirrors backend `entity.AdminUser`. */
export interface AdminUser {
  id: number;
  username: string;
  role: UserRole;
  created_at: string;
  updated_at: string;
  last_login?: string | null;
}

export interface UserInput {
  username: string;
  role: UserRole;
  password?: string;
}

// ---------- API Tokens ----------

export type Scope =
  | "*"
  | "routes:read"
  | "routes:write"
  | "clusters:read"
  | "clusters:write"
  | "middlewares:read"
  | "middlewares:write"
  | "environments:read"
  | "environments:write"
  | "users:read"
  | "users:write"
  | "tokens:read"
  | "tokens:write"
  | "metrics:read"
  | "logs:read";

export const ALL_SCOPES: readonly Scope[] = [
  "routes:read",
  "routes:write",
  "clusters:read",
  "clusters:write",
  "middlewares:read",
  "middlewares:write",
  "environments:read",
  "environments:write",
  "users:read",
  "users:write",
  "tokens:read",
  "tokens:write",
  "metrics:read",
  "logs:read",
] as const;

/** Mirrors backend `entity.APIToken`. The plain-text secret is only ever
 *  returned by the create endpoint and never stored client-side. */
export interface ApiToken {
  id: number;
  user_id: number;
  name: string;
  prefix: string;
  scopes: string;
  last_used_at?: string | null;
  expires_at?: string | null;
  revoked_at?: string | null;
  created_at: string;
}

export interface CreateTokenInput {
  name: string;
  scopes: Scope[];
  expires_at?: string | null;
}

export interface CreateTokenResponse {
  id: number;
  prefix: string;
  scopes: string;
  /** Plaintext secret — shown to the user once and never persisted. */
  secret: string;
}

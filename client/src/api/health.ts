import { http } from "./client";

export interface HealthState {
  status: string;
}

export const healthApi = {
  health: (signal?: AbortSignal) => http.get<HealthState>("/health", { signal, anonymous: true }),
  ready: (signal?: AbortSignal) => http.get<HealthState>("/ready", { signal, anonymous: true }),
};

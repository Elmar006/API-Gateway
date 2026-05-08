/** Pluralise a count with a single label or a label/plural pair. */
export function plural(count: number, single: string, many?: string) {
  return `${count.toLocaleString()} ${count === 1 ? single : (many ?? `${single}s`)}`;
}

/** Compact decimal-suffix formatter (1.2k / 4.6m). */
export function compact(value: number): string {
  if (Math.abs(value) < 1_000) return value.toString();
  if (Math.abs(value) < 1_000_000) return `${(value / 1_000).toFixed(1)}k`;
  if (Math.abs(value) < 1_000_000_000) return `${(value / 1_000_000).toFixed(1)}m`;
  return `${(value / 1_000_000_000).toFixed(1)}b`;
}

/** Format a duration expressed in milliseconds. */
export function formatDuration(ms: number): string {
  if (ms < 1) return `${(ms * 1000).toFixed(0)}µs`;
  if (ms < 1000) return `${ms.toFixed(0)}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}

export function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value));
}

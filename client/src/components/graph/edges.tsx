import { BaseEdge, getBezierPath, type EdgeProps } from "@xyflow/react";

export interface FlowEdgeData extends Record<string, unknown> {
  highlighted?: boolean;
  emphasised?: boolean;
}

export function FlowEdge(props: EdgeProps) {
  const data = (props.data ?? {}) as FlowEdgeData;
  const [path] = getBezierPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    sourcePosition: props.sourcePosition,
    targetX: props.targetX,
    targetY: props.targetY,
    targetPosition: props.targetPosition,
    curvature: 0.3,
  });

  const stroke = data.emphasised
    ? "hsl(var(--accent))"
    : data.highlighted
      ? "hsl(var(--accent-soft))"
      : "hsl(var(--line-strong) / 0.7)";
  const filter = data.emphasised
    ? "drop-shadow(0 0 6px hsl(var(--accent)/0.9)) drop-shadow(0 0 14px hsl(var(--accent)/0.4))"
    : data.highlighted
      ? "drop-shadow(0 0 4px hsl(var(--accent)/0.6))"
      : undefined;

  return (
    <BaseEdge
      id={props.id}
      path={path}
      style={{
        stroke,
        strokeWidth: data.emphasised ? 1.6 : 1.1,
        filter,
        opacity: data.emphasised ? 1 : data.highlighted ? 0.95 : 0.5,
      }}
      markerEnd={props.markerEnd}
    />
  );
}

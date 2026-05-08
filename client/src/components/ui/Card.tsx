import { forwardRef, type HTMLAttributes } from "react";
import { cn } from "@/lib/cn";

export const Card = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        "surface-glass rounded-lg",
        "[&>*+*]:border-t [&>*+*]:border-line-subtle/60",
        className,
      )}
      {...props}
    />
  ),
);
Card.displayName = "Card";

export const CardSection = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn("p-4", className)} {...props} />
  ),
);
CardSection.displayName = "CardSection";

export const CardTitle = forwardRef<HTMLHeadingElement, HTMLAttributes<HTMLHeadingElement>>(
  ({ className, ...props }, ref) => (
    <h3
      ref={ref}
      className={cn("text-step text-text-muted", className)}
      {...props}
    />
  ),
);
CardTitle.displayName = "CardTitle";

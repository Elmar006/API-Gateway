import { forwardRef, type InputHTMLAttributes } from "react";
import { cn } from "@/lib/cn";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  fieldClassName?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(
        "h-9 w-full rounded-md border border-line-subtle bg-bg-glass/60",
        "px-3 text-sm text-text-primary placeholder:text-text-muted",
        "transition-colors duration-150",
        "focus-visible:outline-none focus-visible:border-accent-ring focus-visible:bg-bg-glass focus-visible:shadow-ring",
        "disabled:cursor-not-allowed disabled:opacity-60",
        className,
      )}
      {...props}
    />
  ),
);
Input.displayName = "Input";

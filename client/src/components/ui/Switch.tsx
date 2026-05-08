import * as SwitchPrimitive from "@radix-ui/react-switch";
import { forwardRef, type ComponentPropsWithoutRef, type ElementRef } from "react";

import { cn } from "@/lib/cn";

export const Switch = forwardRef<
  ElementRef<typeof SwitchPrimitive.Root>,
  ComponentPropsWithoutRef<typeof SwitchPrimitive.Root>
>(({ className, ...props }, ref) => (
  <SwitchPrimitive.Root
    ref={ref}
    className={cn(
      "peer inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full",
      "border border-line-subtle transition-colors duration-150",
      "data-[state=checked]:border-accent-ring data-[state=checked]:bg-accent/40",
      "data-[state=unchecked]:bg-bg-chip/70",
      "focus-visible:outline-none focus-visible:shadow-ring",
      className,
    )}
    {...props}
  >
    <SwitchPrimitive.Thumb
      className={cn(
        "pointer-events-none block h-4 w-4 rounded-full bg-text-primary shadow-md transition-transform duration-200",
        "translate-x-0.5 data-[state=checked]:translate-x-[18px]",
        "data-[state=checked]:bg-accent-soft",
      )}
    />
  </SwitchPrimitive.Root>
));
Switch.displayName = SwitchPrimitive.Root.displayName;

import { motion } from "framer-motion";
import { Sparkles } from "lucide-react";

import type { LucideIcon } from "lucide-react";

interface Props {
  title: string;
  description: string;
  icon: LucideIcon;
}

export function SectionPlaceholder({ title, description, icon: Icon }: Props) {
  return (
    <div className="surface-glass relative flex h-full flex-1 flex-col items-center justify-center overflow-hidden rounded-xl border-line-subtle/80 grid-bg">
      <div className="pointer-events-none absolute inset-0 bg-hero-gradient opacity-70" />
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, ease: "easeOut" }}
        className="relative z-10 flex w-full max-w-md flex-col items-center text-center"
      >
        <div className="grid h-14 w-14 place-items-center rounded-xl border border-accent/40 bg-accent/15 text-accent-soft shadow-[0_8px_24px_-12px_hsl(var(--accent)/0.6)]">
          <Icon className="h-6 w-6" />
        </div>
        <h2 className="mt-4 text-xl font-semibold tracking-tight text-text-primary">{title}</h2>
        <p className="mt-2 max-w-sm text-sm text-text-muted text-balance">{description}</p>
        <div className="mt-5 flex items-center gap-2 rounded-full border border-line-subtle bg-bg-glass/60 px-3 py-1 text-[11px] text-text-muted">
          <Sparkles className="h-3 w-3 text-accent-soft" /> Available in the next milestone
        </div>
      </motion.div>
    </div>
  );
}

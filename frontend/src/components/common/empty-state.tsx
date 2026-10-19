import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

export interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  description?: React.ReactNode;
  action?: React.ReactNode;
  className?: string;
  /** Heading element for the title; defaults to h2. */
  headingLevel?: "h1" | "h2" | "h3";
}

export function EmptyState({ icon: Icon, title, description, action, className, headingLevel: Heading = "h2" }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-2 py-12 text-center",
        className,
      )}
    >
      {Icon ? <Icon aria-hidden="true" className="size-8 text-muted-foreground" /> : null}
      <Heading className="font-heading text-lg tracking-[-0.01em]">{title}</Heading>
      {description ? (
        <p className="max-w-prose text-sm text-muted-foreground">{description}</p>
      ) : null}
      {action ? <div className="mt-2">{action}</div> : null}
    </div>
  );
}

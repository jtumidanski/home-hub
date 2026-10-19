import type { LucideIcon } from "lucide-react";
import { AlertCircle, AlertTriangle, CheckCircle2, Info } from "lucide-react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const statusBannerVariants = cva(
  "flex items-start gap-3 rounded-lg border px-4 py-3 text-sm text-foreground",
  {
    variants: {
      variant: {
        destructive: "border-destructive bg-destructive/10",
        success: "border-success bg-success/10",
        warning: "border-warning bg-warning/10",
        info: "border-info bg-info/10",
      },
    },
    defaultVariants: {
      variant: "info",
    },
  },
);

const iconClasses = {
  destructive: "text-destructive",
  success: "text-success",
  warning: "text-warning-text",
  info: "text-info",
} as const;

const defaultIcons: Record<StatusBannerVariant, LucideIcon> = {
  destructive: AlertCircle,
  success: CheckCircle2,
  warning: AlertTriangle,
  info: Info,
};

type StatusBannerVariant = NonNullable<
  VariantProps<typeof statusBannerVariants>["variant"]
>;

export interface StatusBannerProps extends VariantProps<typeof statusBannerVariants> {
  children: React.ReactNode;
  icon?: LucideIcon;
  className?: string;
}

/**
 * Inline status message. Status is conveyed by icon and text, never colour
 * alone. Destructive banners are assertive (role="alert"); the rest are polite
 * (role="status").
 */
function StatusBanner({
  variant = "info",
  icon,
  children,
  className,
}: StatusBannerProps) {
  const resolved: StatusBannerVariant = variant ?? "info";
  const Icon = icon ?? defaultIcons[resolved];
  return (
    <div
      role={resolved === "destructive" ? "alert" : "status"}
      data-slot="status-banner"
      data-variant={resolved}
      className={cn(statusBannerVariants({ variant: resolved }), className)}
    >
      <Icon
        aria-hidden="true"
        className={cn("mt-0.5 size-4 shrink-0", iconClasses[resolved])}
      />
      <div className="min-w-0 flex-1">{children}</div>
    </div>
  );
}

export { StatusBanner, statusBannerVariants };

import {
  Ban,
  CheckCircle2,
  CircleDashed,
  Archive,
  Clock,
  AlertTriangle,
  Truck,
  PackageCheck,
  type LucideIcon,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { STATUS_LABELS } from "@/types/models/package";

type BadgeVariant = "secondary" | "destructive" | "outline" | "success" | "warning" | "info";

const STATUS_VARIANTS: Record<string, BadgeVariant> = {
  pre_transit: "outline",
  in_transit: "info",
  out_for_delivery: "success",
  delivered: "success",
  exception: "destructive",
  stale: "warning",
  archived: "secondary",
};

const STATUS_ICONS: Record<string, LucideIcon> = {
  pre_transit: CircleDashed,
  in_transit: Truck,
  out_for_delivery: PackageCheck,
  delivered: CheckCircle2,
  exception: AlertTriangle,
  stale: Clock,
  archived: Archive,
};

interface StatusBadgeProps {
  status: string;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  const Icon = STATUS_ICONS[status] ?? Ban;
  return (
    <Badge variant={STATUS_VARIANTS[status] ?? "outline"}>
      <Icon aria-hidden="true" className="size-3" />
      {STATUS_LABELS[status] ?? status}
    </Badge>
  );
}

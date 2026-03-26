import { Badge } from "@/components/ui/badge";
import { STATUS_LABELS } from "@/types/models/package";

const STATUS_VARIANTS: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  pre_transit: "outline",
  in_transit: "default",
  out_for_delivery: "default",
  delivered: "secondary",
  exception: "destructive",
  stale: "outline",
  archived: "secondary",
};

interface StatusBadgeProps {
  status: string;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  return (
    <Badge variant={STATUS_VARIANTS[status] ?? "outline"}>
      {STATUS_LABELS[status] ?? status}
    </Badge>
  );
}

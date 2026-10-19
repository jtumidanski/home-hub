import { Truck } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { CARRIER_LABELS } from "@/types/models/package";

interface CarrierIconProps {
  carrier: string;
  className?: string;
}

/** Categorical chart tokens per carrier; the colour only tints the icon, never sits behind text. */
const CARRIER_ICON_COLORS: Record<string, string> = {
  usps: "text-chart-1",
  ups: "text-chart-2",
  fedex: "text-chart-3",
};

export function CarrierIcon({ carrier, className }: CarrierIconProps) {
  const label = CARRIER_LABELS[carrier] ?? carrier.toUpperCase();
  const iconColor = CARRIER_ICON_COLORS[carrier] ?? "text-muted-foreground";

  return (
    <Badge variant="outline" className={cn("text-xs font-semibold", className)}>
      <Truck aria-hidden="true" className={cn("size-3", iconColor)} />
      {label}
    </Badge>
  );
}

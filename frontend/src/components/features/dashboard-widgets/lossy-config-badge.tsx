import { Badge } from "@/components/ui/badge";
import { AlertTriangle } from "lucide-react";

/**
 * Small pill rendered on top of a widget whose persisted config failed
 * validation and was replaced with defaults. Hovering shows the reason.
 */
export function LossyConfigBadge() {
  return (
    <Badge
      variant="outline"
      className="absolute right-2 top-2 gap-1 bg-background/90 backdrop-blur"
      title="Saved configuration was invalid and reduced to defaults. Edit this widget to reconfigure."
    >
      <AlertTriangle className="h-3 w-3" />
      reduced to defaults
    </Badge>
  );
}

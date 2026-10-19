import { Badge } from "@/components/ui/badge";
import { AlertTriangle } from "lucide-react";
import { useId } from "react";

/**
 * Pill rendered on top of a widget whose persisted config failed validation
 * and was replaced with defaults. The explanation is real text (sr-only plus
 * a title hint), not title-only.
 */
export function LossyConfigBadge() {
  const explanationId = useId();
  return (
    <Badge
      variant="warning"
      className="absolute right-2 top-2 gap-1"
      aria-describedby={explanationId}
      title="Saved configuration was invalid and reduced to defaults. Edit this widget to reconfigure."
    >
      <AlertTriangle aria-hidden="true" className="size-3" />
      Defaults used
      <span id={explanationId} className="sr-only">
        Saved configuration was invalid and reduced to defaults. Edit this widget to reconfigure.
      </span>
    </Badge>
  );
}

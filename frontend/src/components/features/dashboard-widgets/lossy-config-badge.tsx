import { Badge } from "@/components/ui/badge";
import { AlertTriangle } from "lucide-react";

const EXPLANATION_ID = "lossy-config-explanation";

/**
 * Pill rendered on top of a widget whose persisted config failed validation
 * and was replaced with defaults. The explanation is real text (sr-only plus
 * a title hint), not title-only.
 */
export function LossyConfigBadge() {
  return (
    <Badge
      variant="warning"
      className="absolute right-2 top-2 gap-1"
      aria-describedby={EXPLANATION_ID}
      title="Saved configuration was invalid and reduced to defaults. Edit this widget to reconfigure."
    >
      <AlertTriangle aria-hidden="true" className="size-3" />
      Defaults used
      <span id={EXPLANATION_ID} className="sr-only">
        Saved configuration was invalid and reduced to defaults. Edit this widget to reconfigure.
      </span>
    </Badge>
  );
}

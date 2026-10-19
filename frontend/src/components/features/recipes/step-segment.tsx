import { Carrot, CookingPot, Timer, Link2 } from "lucide-react";
import type { Segment } from "@/types/models/recipe";
import { cn } from "@/lib/utils";

interface StepSegmentProps {
  segment: Segment;
  size?: "default" | "large";
}

const chip = "rounded px-1 font-medium text-foreground";

function SegmentIcon({ icon: Icon, label }: { icon: typeof Carrot; label: string }) {
  return (
    <>
      <Icon aria-hidden="true" className="mr-0.5 inline h-3 w-3 align-[-0.125em]" />
      <span className="sr-only">{label}: </span>
    </>
  );
}

export function StepSegment({ segment, size = "default" }: StepSegmentProps) {
  const large = size === "large";

  switch (segment.type) {
    case "ingredient":
      return (
        <span className={cn(chip, "bg-swatch-amber-soft")}>
          <SegmentIcon icon={Carrot} label="Ingredient" />
          {segment.name}
          {segment.quantity && (
            <span className={cn(large ? "ml-1" : "text-xs ml-0.5")}>
              ({segment.quantity}{segment.unit && ` ${segment.unit}`})
            </span>
          )}
        </span>
      );
    case "cookware":
      return (
        <span className={cn(chip, "bg-swatch-blue-soft")}>
          <SegmentIcon icon={CookingPot} label="Cookware" />
          {segment.name}
        </span>
      );
    case "timer":
      return (
        <span className={cn(chip, "bg-swatch-green-soft")}>
          <SegmentIcon icon={Timer} label="Timer" />
          {segment.quantity} {segment.unit}
        </span>
      );
    case "reference":
      return (
        <span className={cn(chip, "bg-swatch-violet-soft italic")}>
          <SegmentIcon icon={Link2} label="Reference" />
          {segment.name}
        </span>
      );
    default:
      return <span>{segment.value}</span>;
  }
}

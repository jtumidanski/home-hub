import type { Segment } from "@/types/models/recipe";
import { cn } from "@/lib/utils";

interface StepSegmentProps {
  segment: Segment;
  size?: "default" | "large";
}

export function StepSegment({ segment, size = "default" }: StepSegmentProps) {
  const large = size === "large";

  switch (segment.type) {
    case "ingredient":
      return (
        <span className={cn("font-medium text-orange-600 dark:text-orange-400")}>
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
        <span className="font-medium text-blue-600 dark:text-blue-400">
          {segment.name}
        </span>
      );
    case "timer":
      return (
        <span className="font-medium text-green-600 dark:text-green-400">
          {segment.quantity} {segment.unit}
        </span>
      );
    case "reference":
      return (
        <span className="font-medium text-purple-600 dark:text-purple-400 italic">
          {segment.name}
        </span>
      );
    default:
      return <span>{segment.value}</span>;
  }
}

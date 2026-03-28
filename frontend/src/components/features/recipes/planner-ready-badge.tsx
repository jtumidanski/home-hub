import { Badge } from "@/components/ui/badge";

interface PlannerReadyBadgeProps {
  ready: boolean;
  issues?: string[];
  className?: string;
}

export function PlannerReadyBadge({ ready, issues, className }: PlannerReadyBadgeProps) {
  if (ready) {
    return (
      <Badge variant="default" className={className}>
        Planner Ready
      </Badge>
    );
  }

  return (
    <Badge
      variant="outline"
      className={className}
      title={issues?.join(", ") ?? "Not planner ready"}
    >
      Not Planner Ready
    </Badge>
  );
}

import type { Step } from "@/types/models/recipe";
import { cn } from "@/lib/utils";

interface RecipeStepsProps {
  steps: Step[];
}

export function RecipeSteps({ steps }: RecipeStepsProps) {
  if (steps.length === 0) {
    return <p className="text-sm text-muted-foreground">No steps found.</p>;
  }

  let currentSection = "";

  return (
    <div className="space-y-4">
      {steps.map((step) => {
        const showSection = step.section && step.section !== currentSection;
        if (step.section) {
          currentSection = step.section;
        }

        return (
          <div key={step.number}>
            {showSection && (
              <h3 className="text-sm font-semibold uppercase text-muted-foreground mt-4 mb-2">
                {step.section}
              </h3>
            )}
            <div className="flex gap-3">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground text-xs font-medium">
                {step.number}
              </span>
              <p className="text-sm leading-relaxed pt-0.5">
                {step.segments.map((seg, i) => {
                  switch (seg.type) {
                    case "ingredient":
                      return (
                        <span key={i} className={cn("font-medium text-orange-600 dark:text-orange-400")}>
                          {seg.name}
                          {seg.quantity && <span className="text-xs ml-0.5">({seg.quantity}{seg.unit && ` ${seg.unit}`})</span>}
                        </span>
                      );
                    case "cookware":
                      return (
                        <span key={i} className="font-medium text-blue-600 dark:text-blue-400">
                          {seg.name}
                        </span>
                      );
                    case "timer":
                      return (
                        <span key={i} className="font-medium text-green-600 dark:text-green-400">
                          {seg.quantity} {seg.unit}
                        </span>
                      );
                    case "reference":
                      return (
                        <span key={i} className="font-medium text-purple-600 dark:text-purple-400 italic">
                          {seg.name}
                        </span>
                      );
                    default:
                      return <span key={i}>{seg.value}</span>;
                  }
                })}
              </p>
            </div>
          </div>
        );
      })}
    </div>
  );
}

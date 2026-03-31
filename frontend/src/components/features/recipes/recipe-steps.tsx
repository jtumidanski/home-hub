import type { Step } from "@/types/models/recipe";
import { StepSegment } from "./step-segment";

interface RecipeStepsProps {
  steps: Step[];
}

export function RecipeSteps({ steps }: RecipeStepsProps) {
  if (steps.length === 0) {
    return <p className="text-sm text-muted-foreground">No steps found.</p>;
  }

  return (
    <div className="space-y-4">
      {steps.map((step, idx) => {
        const prevSection = idx > 0 ? steps[idx - 1]?.section : undefined;
        const showSection = !!step.section && step.section !== prevSection;

        return (
          <div key={step.number}>
            {showSection && (
              <h4 className="text-sm font-semibold text-muted-foreground mt-4 mb-2">
                {step.section}
              </h4>
            )}
            <div className="flex gap-3">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground text-xs font-medium">
                {step.number}
              </span>
              <p className="text-sm leading-relaxed pt-0.5">
                {step.segments.map((seg, i) => (
                  <StepSegment key={i} segment={seg} />
                ))}
              </p>
            </div>
          </div>
        );
      })}
    </div>
  );
}

import { MessageSquareText } from "lucide-react";
import type { PositionalNote, Step } from "@/types/models/recipe";
import { StepSegment } from "./step-segment";

interface RecipeStepsProps {
  steps: Step[];
  notes?: PositionalNote[] | undefined;
}

function RecipeNote({ text }: { text: string }) {
  return (
    <div className="flex items-start gap-2 border-l-4 border-muted-foreground/40 pl-3 py-1 text-sm italic text-muted-foreground">
      <MessageSquareText className="h-4 w-4 mt-0.5 shrink-0" aria-hidden="true" />
      <span><span className="sr-only">Note: </span>{text}</span>
    </div>
  );
}

export function RecipeSteps({ steps, notes }: RecipeStepsProps) {
  if (steps.length === 0 && (!notes || notes.length === 0)) {
    return <p className="text-sm text-muted-foreground">No steps found.</p>;
  }

  const notesByPosition = new Map<number, PositionalNote[]>();
  for (const n of notes ?? []) {
    const list = notesByPosition.get(n.position) ?? [];
    list.push(n);
    notesByPosition.set(n.position, list);
  }

  const trailingNotes = notesByPosition.get(steps.length) ?? [];

  return (
    <div className="space-y-4">
      {steps.map((step, idx) => {
        const prevSection = idx > 0 ? steps[idx - 1]?.section : undefined;
        const showSection = !!step.section && step.section !== prevSection;
        const inlineNotes = notesByPosition.get(idx) ?? [];

        return (
          <div key={step.number} className="space-y-2">
            {inlineNotes.map((n, i) => (
              <RecipeNote key={`note-${idx}-${i}`} text={n.text} />
            ))}
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
      {trailingNotes.map((n, i) => (
        <RecipeNote key={`trailing-${i}`} text={n.text} />
      ))}
    </div>
  );
}

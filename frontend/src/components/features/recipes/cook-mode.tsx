import { useState, useEffect, useCallback, useRef } from "react";
import { X, ChevronLeft, ChevronRight, Eye } from "lucide-react";
import type { Step } from "@/types/models/recipe";
import { StepSegment } from "./step-segment";
import { Button } from "@/components/ui/button";
import { useWakeLock } from "@/lib/hooks/use-wake-lock";
import { cn } from "@/lib/utils";

interface CookModeProps {
  steps: Step[];
  title: string;
  open: boolean;
  onClose: () => void;
}

type ViewMode = "all" | "single";

export function CookMode({ steps, title, open, onClose }: CookModeProps) {
  const [viewMode, setViewMode] = useState<ViewMode>("all");
  const [currentStep, setCurrentStep] = useState(0);
  const { isActive: wakeLockActive } = useWakeLock(open);
  const touchStartX = useRef<number | null>(null);

  // Close on Escape
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [open, onClose]);

  // Arrow key navigation in single-step mode
  useEffect(() => {
    if (!open || viewMode !== "single") return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "ArrowLeft") {
        setCurrentStep((s) => Math.max(0, s - 1));
      } else if (e.key === "ArrowRight") {
        setCurrentStep((s) => Math.min(steps.length - 1, s + 1));
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [open, viewMode, steps.length]);

  // Swipe handlers
  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    const touch = e.touches[0];
    if (touch) touchStartX.current = touch.clientX;
  }, []);

  const handleTouchEnd = useCallback(
    (e: React.TouchEvent) => {
      if (touchStartX.current === null || viewMode !== "single") return;
      const touch = e.changedTouches[0];
      if (!touch) return;
      const deltaX = touch.clientX - touchStartX.current;
      touchStartX.current = null;
      if (Math.abs(deltaX) < 50) return;
      if (deltaX < 0) {
        setCurrentStep((s) => Math.min(steps.length - 1, s + 1));
      } else {
        setCurrentStep((s) => Math.max(0, s - 1));
      }
    },
    [viewMode, steps.length],
  );

  // Reset step index when switching to single mode
  const handleViewModeChange = (mode: ViewMode) => {
    setViewMode(mode);
    if (mode === "single") setCurrentStep(0);
  };

  if (!open) return null;

  // Build flat list with section tracking for all-steps view
  const stepsWithSections = steps.map((step, idx) => {
    const prevSection = idx > 0 ? steps[idx - 1]?.section : undefined;
    const showSection = !!step.section && step.section !== prevSection;
    return { step, showSection };
  });

  return (
    <div
      className="fixed inset-0 z-50 bg-background text-foreground flex flex-col"
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b shrink-0">
        <div className="flex items-center gap-3 min-w-0">
          <h2 className="text-lg font-semibold truncate">{title}</h2>
          <div className="flex gap-1 shrink-0">
            <Button
              variant={viewMode === "all" ? "default" : "outline"}
              size="sm"
              onClick={() => handleViewModeChange("all")}
            >
              All Steps
            </Button>
            <Button
              variant={viewMode === "single" ? "default" : "outline"}
              size="sm"
              onClick={() => handleViewModeChange("single")}
            >
              Single Step
            </Button>
          </div>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {wakeLockActive && (
            <Eye className="h-4 w-4 text-muted-foreground" aria-label="Screen wake lock active" />
          )}
          <Button variant="ghost" size="icon" onClick={onClose} aria-label="Close cook mode">
            <X className="h-5 w-5" />
          </Button>
        </div>
      </div>

      {/* Content */}
      {viewMode === "all" ? (
        <div className="flex-1 overflow-y-auto px-4 py-6 md:px-8">
          <div className="max-w-4xl mx-auto space-y-6">
            {stepsWithSections.map(({ step, showSection }) => (
              <div key={step.number}>
                {showSection && (
                  <h3
                    className="font-semibold text-muted-foreground mb-3 mt-4"
                    style={{ fontSize: "clamp(1rem, 2.5vw, 1.75rem)" }}
                  >
                    {step.section}
                  </h3>
                )}
                <div className="flex gap-4 items-start">
                  <span
                    className="flex shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground font-bold"
                    style={{
                      fontSize: "clamp(0.875rem, 2vw, 1.5rem)",
                      width: "clamp(2rem, 4vw, 3rem)",
                      height: "clamp(2rem, 4vw, 3rem)",
                    }}
                  >
                    {step.number}
                  </span>
                  <p
                    className="leading-relaxed pt-0.5"
                    style={{ fontSize: "clamp(1.25rem, 3.5vw, 3rem)" }}
                  >
                    {step.segments.map((seg, i) => (
                      <StepSegment key={i} segment={seg} size="large" />
                    ))}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : (
        <SingleStepView
          steps={steps}
          currentStep={currentStep}
          onPrev={() => setCurrentStep((s) => Math.max(0, s - 1))}
          onNext={() => setCurrentStep((s) => Math.min(steps.length - 1, s + 1))}
        />
      )}
    </div>
  );
}

function SingleStepView({
  steps,
  currentStep,
  onPrev,
  onNext,
}: {
  steps: Step[];
  currentStep: number;
  onPrev: () => void;
  onNext: () => void;
}) {
  const step = steps[currentStep];
  if (!step) return null;

  // Determine if we should show a section header
  const prevSection = currentStep > 0 ? steps[currentStep - 1]?.section : undefined;
  const showSection = !!step.section && step.section !== prevSection;

  return (
    <>
      <div className="flex-1 overflow-y-auto px-6 md:px-12">
        <div className="min-h-full flex flex-col items-center justify-center py-6">
          {showSection && (
            <h3
              className="font-semibold text-muted-foreground mb-4"
              style={{ fontSize: "clamp(1.25rem, 3vw, 2.5rem)" }}
            >
              {step.section}
            </h3>
          )}
          <p
            className="leading-relaxed text-center max-w-4xl"
            style={{ fontSize: "clamp(1.5rem, 5vw, 5rem)" }}
          >
            {step.segments.map((seg, i) => (
              <StepSegment key={i} segment={seg} size="large" />
            ))}
          </p>
        </div>
      </div>

      {/* Navigation bar */}
      <div className="flex items-center justify-between px-4 py-3 border-t shrink-0">
        <Button
          variant="outline"
          size="lg"
          onClick={onPrev}
          disabled={currentStep === 0}
          className={cn("min-w-[48px] min-h-[48px]")}
          aria-label="Previous step"
        >
          <ChevronLeft className="h-6 w-6" />
        </Button>
        <span
          className="font-medium text-muted-foreground"
          style={{ fontSize: "clamp(1rem, 2.5vw, 1.75rem)" }}
        >
          {currentStep + 1} / {steps.length}
        </span>
        <Button
          variant="outline"
          size="lg"
          onClick={onNext}
          disabled={currentStep === steps.length - 1}
          className={cn("min-w-[48px] min-h-[48px]")}
          aria-label="Next step"
        >
          <ChevronRight className="h-6 w-6" />
        </Button>
      </div>
    </>
  );
}

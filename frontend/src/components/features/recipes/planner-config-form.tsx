import { useState } from "react";
import { ChevronDown, ChevronRight } from "lucide-react";
import { Input } from "@/components/ui/input";
import type { PlannerConfig } from "@/types/models/recipe";

interface PlannerConfigFormProps {
  value: PlannerConfig;
  onChange: (config: PlannerConfig) => void;
}

export function PlannerConfigForm({ value, onChange }: PlannerConfigFormProps) {
  const [isOpen, setIsOpen] = useState(
    !!(value.eatWithinDays || value.minGapDays || value.maxConsecutiveDays),
  );

  const update = (key: keyof PlannerConfig, val: number | undefined) => {
    onChange({ ...value, [key]: val });
  };

  return (
    <div className="border rounded-md">
      <button
        type="button"
        className="flex items-center gap-2 w-full px-3 py-2 text-sm font-medium text-left hover:bg-muted/50 transition-colors"
        onClick={() => setIsOpen(!isOpen)}
      >
        {isOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
        Planner Scheduling
      </button>

      {isOpen && (
        <div className="px-3 pb-3 space-y-2">
          <p className="text-xs text-muted-foreground">
            Classification is derived from tags (breakfast, lunch, dinner, snack, side). Servings come from Cooklang metadata.
          </p>
          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="text-xs text-muted-foreground">Eat Within (days)</label>
              <Input
                type="number"
                min={1}
                value={value.eatWithinDays ?? ""}
                onChange={(e) => update("eatWithinDays", e.target.value ? Number(e.target.value) : undefined)}
                className="h-8 text-sm"
              />
            </div>
            <div>
              <label className="text-xs text-muted-foreground">Min Gap (days)</label>
              <Input
                type="number"
                min={0}
                value={value.minGapDays ?? ""}
                onChange={(e) => update("minGapDays", e.target.value ? Number(e.target.value) : undefined)}
                className="h-8 text-sm"
              />
            </div>
            <div>
              <label className="text-xs text-muted-foreground">Max Consecutive (days)</label>
              <Input
                type="number"
                min={1}
                value={value.maxConsecutiveDays ?? ""}
                onChange={(e) => update("maxConsecutiveDays", e.target.value ? Number(e.target.value) : undefined)}
                className="h-8 text-sm"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

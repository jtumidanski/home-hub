import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { RecipeSort } from "@/types/models/recipe";

const DEFAULT_VALUE = "default";

interface RecipeSortSelectProps {
  value: RecipeSort | undefined;
  onChange: (value: RecipeSort | undefined) => void;
  className?: string;
}

export function RecipeSortSelect({ value, onChange, className }: RecipeSortSelectProps) {
  return (
    <Select
      value={value ?? DEFAULT_VALUE}
      onValueChange={(v) => onChange(v === DEFAULT_VALUE ? undefined : (v as RecipeSort))}
    >
      <SelectTrigger className={className ?? "w-[140px]"} aria-label="Sort recipes">
        <SelectValue placeholder="Sort" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={DEFAULT_VALUE}>Default</SelectItem>
        <SelectItem value="-usageCount">Most cooked</SelectItem>
        <SelectItem value="usageCount">Least cooked</SelectItem>
      </SelectContent>
    </Select>
  );
}

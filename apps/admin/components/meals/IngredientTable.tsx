import { Ingredient } from "@/lib/api/meals";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { AlertCircle } from "lucide-react";

interface IngredientTableProps {
  ingredients: Ingredient[];
}

export function IngredientTable({ ingredients }: IngredientTableProps) {
  if (ingredients.length === 0) {
    return (
      <div className="text-center py-8 text-neutral-500 dark:text-neutral-400">
        No ingredients parsed yet
      </div>
    );
  }

  const formatValue = (value: string | number | null | undefined): string => {
    if (value === null || value === undefined || value === "") {
      return "—";
    }
    return String(value);
  };

  const formatArray = (arr: string[] | undefined | null): string => {
    if (!arr || arr.length === 0) {
      return "—";
    }
    return arr.join(", ");
  };

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[100px]">Quantity</TableHead>
            <TableHead className="w-[100px]">Unit</TableHead>
            <TableHead>Ingredient</TableHead>
            <TableHead>Preparation</TableHead>
            <TableHead>Notes</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {ingredients.map((ing) => (
            <TableRow key={ing.id}>
              <TableCell className="font-medium">
                {formatValue(ing.quantityRaw || ing.quantity)}
              </TableCell>
              <TableCell>
                {formatValue(ing.unitRaw || ing.unit)}
              </TableCell>
              <TableCell>
                <div className="flex items-center gap-2">
                  <span>{ing.ingredient}</span>
                  {ing.confidence < 0.7 && (
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <AlertCircle className="h-4 w-4 text-yellow-500 flex-shrink-0" />
                        </TooltipTrigger>
                        <TooltipContent>
                          <p className="text-sm">
                            AI parsing confidence is low. Verify this ingredient.
                          </p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  )}
                </div>
              </TableCell>
              <TableCell className="text-neutral-600 dark:text-neutral-400">
                {formatArray(ing.preparation)}
              </TableCell>
              <TableCell className="text-neutral-600 dark:text-neutral-400">
                {formatArray(ing.notes)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

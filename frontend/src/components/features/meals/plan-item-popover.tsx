import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import {
  planItemPopoverSchema,
  planItemPopoverDefaults,
  planItemAddSchema,
  planItemAddDefaults,
  servingsRefinement,
  type PlanItemPopoverFormData,
  type PlanItemAddFormData,
} from "@/lib/schemas/meals.schema";
import type { Slot, PlanItemAttributes } from "@/types/models/meal-plan";
import { SLOTS } from "@/types/models/meal-plan";

interface PlanItemPopoverProps {
  open: boolean;
  onClose: () => void;
  onSave: (data: {
    day: string;
    slot: Slot;
    serving_multiplier?: number | null;
    planned_servings?: number | null;
    notes?: string | null;
  }) => void;
  weekDays: { dateStr: string; label: string }[];
  initialDay?: string;
  initialSlot?: Slot;
  editItem?: PlanItemAttributes | null;
  recipeServings?: number | null | undefined;
}

const SLOT_LABELS: Record<Slot, string> = {
  breakfast: "Breakfast",
  lunch: "Lunch",
  dinner: "Dinner",
  snack: "Snack",
  side: "Side",
};

function EditModeForm({
  onClose,
  onSave,
  weekDays,
  editItem,
  recipeServings,
}: {
  onClose: () => void;
  onSave: PlanItemPopoverProps["onSave"];
  weekDays: PlanItemPopoverProps["weekDays"];
  editItem: PlanItemAttributes;
  recipeServings?: number | null | undefined;
}) {
  const effectiveServings = recipeServings ?? editItem.recipe_servings;
  const [servingMode, setServingMode] = useState<"multiplier" | "planned">(
    editItem.serving_multiplier ? "multiplier" : "planned"
  );

  const schema = planItemPopoverSchema.superRefine((data, ctx) =>
    servingsRefinement(data, ctx, servingMode)
  );

  const form = useForm<PlanItemPopoverFormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      ...planItemPopoverDefaults,
      day: editItem.day,
      slot: editItem.slot,
      serving_multiplier: editItem.serving_multiplier ?? null,
      planned_servings: editItem.planned_servings ?? null,
      notes: editItem.notes ?? null,
    },
  });

  const onSubmit = (values: PlanItemPopoverFormData) => {
    const data: Parameters<PlanItemPopoverProps["onSave"]>[0] = {
      day: values.day,
      slot: values.slot,
    };
    if (servingMode === "planned" && values.planned_servings) {
      data.planned_servings = values.planned_servings;
      data.serving_multiplier = null;
    } else if (servingMode === "multiplier" && values.serving_multiplier) {
      data.serving_multiplier = values.serving_multiplier;
      data.planned_servings = null;
    }
    if (values.notes?.trim()) {
      data.notes = values.notes.trim();
    }
    onSave(data);
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 py-2">
        <div className="grid grid-cols-2 gap-3">
          <FormField
            control={form.control}
            name="day"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Day</FormLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <FormControl>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {weekDays.map((d) => (
                      <SelectItem key={d.dateStr} value={d.dateStr}>{d.label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="slot"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Slot</FormLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <FormControl>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {SLOTS.map((s) => (
                      <SelectItem key={s} value={s}>{SLOT_LABELS[s]}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <ServingsFields
          form={form}
          servingMode={servingMode}
          setServingMode={setServingMode}
          recipeServings={effectiveServings}
        />
        <NotesField form={form} />

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
          <Button type="submit">Save</Button>
        </DialogFooter>
      </form>
    </Form>
  );
}

function AddModeForm({
  onClose,
  onSave,
  weekDays,
  initialDay,
  initialSlot,
  recipeServings,
}: {
  onClose: () => void;
  onSave: PlanItemPopoverProps["onSave"];
  weekDays: PlanItemPopoverProps["weekDays"];
  initialDay?: string | undefined;
  initialSlot?: Slot | undefined;
  recipeServings?: number | null | undefined;
}) {
  const [servingMode, setServingMode] = useState<"multiplier" | "planned">("planned");

  const schema = planItemAddSchema.superRefine((data, ctx) =>
    servingsRefinement(data, ctx, servingMode)
  );

  const form = useForm<PlanItemAddFormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      ...planItemAddDefaults,
      days: initialDay ? [initialDay] : [],
      slot: initialSlot ?? "dinner",
    },
  });

  const onSubmit = (values: PlanItemAddFormData) => {
    const base: Omit<Parameters<PlanItemPopoverProps["onSave"]>[0], "day"> = {
      slot: values.slot,
    };
    if (servingMode === "planned" && values.planned_servings) {
      base.planned_servings = values.planned_servings;
      base.serving_multiplier = null;
    } else if (servingMode === "multiplier" && values.serving_multiplier) {
      base.serving_multiplier = values.serving_multiplier;
      base.planned_servings = null;
    }
    if (values.notes?.trim()) {
      base.notes = values.notes.trim();
    }

    for (const day of values.days) {
      onSave({ ...base, day, slot: values.slot });
    }
  };

  const selectedDays = form.watch("days");

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 py-2">
        <FormField
          control={form.control}
          name="days"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Days</FormLabel>
              <div className="flex flex-wrap gap-1.5">
                {weekDays.map((d) => {
                  const checked = field.value.includes(d.dateStr);
                  return (
                    <button
                      key={d.dateStr}
                      type="button"
                      className={`px-2.5 py-1 rounded-md text-xs font-medium border transition-colors ${
                        checked
                          ? "bg-primary text-primary-foreground border-primary"
                          : "bg-background text-muted-foreground border-input hover:bg-accent hover:text-accent-foreground"
                      }`}
                      onClick={() => {
                        const next = checked
                          ? field.value.filter((v: string) => v !== d.dateStr)
                          : [...field.value, d.dateStr];
                        field.onChange(next);
                      }}
                    >
                      {d.label}
                    </button>
                  );
                })}
              </div>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="slot"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Slot</FormLabel>
              <Select value={field.value} onValueChange={field.onChange}>
                <FormControl>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                </FormControl>
                <SelectContent>
                  {SLOTS.map((s) => (
                    <SelectItem key={s} value={s}>{SLOT_LABELS[s]}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        <ServingsFields
          form={form}
          servingMode={servingMode}
          setServingMode={setServingMode}
          recipeServings={recipeServings}
        />
        <NotesField form={form} />

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
          <Button type="submit">
            {selectedDays.length > 1 ? `Add to ${selectedDays.length} days` : "Save"}
          </Button>
        </DialogFooter>
      </form>
    </Form>
  );
}

function ServingsFields({
  form,
  servingMode,
  setServingMode,
  recipeServings,
}: {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: any;
  servingMode: "multiplier" | "planned";
  setServingMode: (mode: "multiplier" | "planned") => void;
  recipeServings?: number | null | undefined;
}) {
  return (
    <div>
      <FormLabel>Servings</FormLabel>
      {recipeServings != null && (
        <p className="text-xs text-muted-foreground mt-0.5">
          Recipe serves {recipeServings}
        </p>
      )}
      <div className="flex gap-2 mt-1">
        <Select value={servingMode} onValueChange={(v) => v && setServingMode(v as "multiplier" | "planned")}>
          <SelectTrigger className="w-[140px]"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="planned">Planned</SelectItem>
            <SelectItem value="multiplier">Multiplier</SelectItem>
          </SelectContent>
        </Select>
        {servingMode === "multiplier" ? (
          <FormField
            control={form.control}
            name="serving_multiplier"
            render={({ field }: { field: { value: number | null; onChange: (v: number | null) => void } }) => (
              <FormItem className="flex-1">
                <FormControl>
                  <Input
                    type="number"
                    step="0.25"
                    min="0.25"
                    placeholder="1.0"
                    value={field.value ?? ""}
                    onChange={(e) => field.onChange(e.target.value ? parseFloat(e.target.value) : null)}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        ) : (
          <FormField
            control={form.control}
            name="planned_servings"
            render={({ field }: { field: { value: number | null; onChange: (v: number | null) => void } }) => (
              <FormItem className="flex-1">
                <FormControl>
                  <Input
                    type="number"
                    step="1"
                    min="1"
                    placeholder={recipeServings ? String(recipeServings) : "4"}
                    value={field.value ?? ""}
                    onChange={(e) => field.onChange(e.target.value ? parseInt(e.target.value, 10) : null)}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        )}
      </div>
    </div>
  );
}

function NotesField({
  form,
}: {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: any;
}) {
  return (
    <FormField
      control={form.control}
      name="notes"
      render={({ field }: { field: { value: string | null; onChange: (v: string | null) => void } }) => (
        <FormItem>
          <FormLabel>Notes</FormLabel>
          <FormControl>
            <Textarea
              placeholder="Optional notes..."
              value={field.value ?? ""}
              onChange={(e) => field.onChange(e.target.value || null)}
              rows={2}
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}

export function PlanItemPopover({
  open,
  onClose,
  onSave,
  weekDays,
  initialDay,
  initialSlot,
  editItem,
  recipeServings,
}: PlanItemPopoverProps) {
  const isEdit = !!editItem;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Item" : "Add Item"}</DialogTitle>
        </DialogHeader>
        {isEdit ? (
          <EditModeForm
            onClose={onClose}
            onSave={onSave}
            weekDays={weekDays}
            editItem={editItem!}
            recipeServings={recipeServings}
          />
        ) : (
          <AddModeForm
            onClose={onClose}
            onSave={onSave}
            weekDays={weekDays}
            initialDay={initialDay}
            initialSlot={initialSlot}
            recipeServings={recipeServings}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

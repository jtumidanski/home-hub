import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { planItemPopoverSchema, planItemPopoverDefaults, type PlanItemPopoverFormData } from "@/lib/schemas/meals.schema";
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
}

const SLOT_LABELS: Record<Slot, string> = {
  breakfast: "Breakfast",
  lunch: "Lunch",
  dinner: "Dinner",
  snack: "Snack",
  side: "Side",
};

export function PlanItemPopover({
  open,
  onClose,
  onSave,
  weekDays,
  initialDay,
  initialSlot,
  editItem,
}: PlanItemPopoverProps) {
  const [servingMode, setServingMode] = useState<"multiplier" | "planned">(
    editItem?.planned_servings ? "planned" : "multiplier"
  );

  const form = useForm<PlanItemPopoverFormData>({
    resolver: zodResolver(planItemPopoverSchema),
    defaultValues: {
      ...planItemPopoverDefaults,
      day: editItem?.day ?? initialDay ?? weekDays[0]?.dateStr ?? "",
      slot: editItem?.slot ?? initialSlot ?? "dinner",
      serving_multiplier: editItem?.serving_multiplier ?? null,
      planned_servings: editItem?.planned_servings ?? null,
      notes: editItem?.notes ?? null,
    },
  });

  useEffect(() => {
    if (open) {
      setServingMode(editItem?.planned_servings ? "planned" : "multiplier");
      form.reset({
        ...planItemPopoverDefaults,
        day: editItem?.day ?? initialDay ?? weekDays[0]?.dateStr ?? "",
        slot: editItem?.slot ?? initialSlot ?? "dinner",
        serving_multiplier: editItem?.serving_multiplier ?? null,
        planned_servings: editItem?.planned_servings ?? null,
        notes: editItem?.notes ?? null,
      });
    }
  }, [open, editItem, initialDay, initialSlot]); // eslint-disable-line react-hooks/exhaustive-deps

  const onSubmit = (values: PlanItemPopoverFormData) => {
    const data: {
      day: string;
      slot: Slot;
      serving_multiplier?: number | null;
      planned_servings?: number | null;
      notes?: string | null;
    } = { day: values.day, slot: values.slot };

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
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>{editItem ? "Edit Item" : "Add Item"}</DialogTitle>
        </DialogHeader>
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

            <div>
              <FormLabel>Servings</FormLabel>
              <div className="flex gap-2 mt-1">
                <Select value={servingMode} onValueChange={(v) => v && setServingMode(v as "multiplier" | "planned")}>
                  <SelectTrigger className="w-[140px]"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="multiplier">Multiplier</SelectItem>
                    <SelectItem value="planned">Planned</SelectItem>
                  </SelectContent>
                </Select>
                {servingMode === "multiplier" ? (
                  <FormField
                    control={form.control}
                    name="serving_multiplier"
                    render={({ field }) => (
                      <FormItem className="flex-1">
                        <FormControl>
                          <Input
                            type="number"
                            step="0.5"
                            min="0.5"
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
                    render={({ field }) => (
                      <FormItem className="flex-1">
                        <FormControl>
                          <Input
                            type="number"
                            step="1"
                            min="1"
                            placeholder="4"
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

            <FormField
              control={form.control}
              name="notes"
              render={({ field }) => (
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

            <DialogFooter>
              <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
              <Button type="submit">Save</Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

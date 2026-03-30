import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
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
  const [day, setDay] = useState(editItem?.day ?? initialDay ?? weekDays[0]?.dateStr ?? "");
  const [slot, setSlot] = useState<Slot>(editItem?.slot ?? initialSlot ?? "dinner");
  const [servingMode, setServingMode] = useState<"multiplier" | "planned">(
    editItem?.planned_servings ? "planned" : "multiplier"
  );
  const [servingMultiplier, setServingMultiplier] = useState(
    editItem?.serving_multiplier?.toString() ?? ""
  );
  const [plannedServings, setPlannedServings] = useState(
    editItem?.planned_servings?.toString() ?? ""
  );
  const [notes, setNotes] = useState(editItem?.notes ?? "");

  const handleSave = () => {
    const data: {
      day: string;
      slot: Slot;
      serving_multiplier?: number | null;
      planned_servings?: number | null;
      notes?: string | null;
    } = { day, slot };

    if (servingMode === "planned" && plannedServings) {
      data.planned_servings = parseInt(plannedServings, 10);
      data.serving_multiplier = null;
    } else if (servingMode === "multiplier" && servingMultiplier) {
      data.serving_multiplier = parseFloat(servingMultiplier);
      data.planned_servings = null;
    }

    if (notes.trim()) {
      data.notes = notes.trim();
    }

    onSave(data);
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>{editItem ? "Edit Item" : "Add Item"}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label>Day</Label>
              <Select value={day} onValueChange={(v) => setDay(v ?? "")}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {weekDays.map((d) => (
                    <SelectItem key={d.dateStr} value={d.dateStr}>{d.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Slot</Label>
              <Select value={slot} onValueChange={(v) => v && setSlot(v as Slot)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {SLOTS.map((s) => (
                    <SelectItem key={s} value={s}>{SLOT_LABELS[s]}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div>
            <Label>Servings</Label>
            <div className="flex gap-2 mt-1">
              <Select value={servingMode} onValueChange={(v) => v && setServingMode(v as "multiplier" | "planned")}>
                <SelectTrigger className="w-[140px]"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="multiplier">Multiplier</SelectItem>
                  <SelectItem value="planned">Planned</SelectItem>
                </SelectContent>
              </Select>
              {servingMode === "multiplier" ? (
                <Input
                  type="number"
                  step="0.5"
                  min="0.5"
                  placeholder="1.0"
                  value={servingMultiplier}
                  onChange={(e) => setServingMultiplier(e.target.value)}
                />
              ) : (
                <Input
                  type="number"
                  step="1"
                  min="1"
                  placeholder="4"
                  value={plannedServings}
                  onChange={(e) => setPlannedServings(e.target.value)}
                />
              )}
            </div>
          </div>

          <div>
            <Label>Notes</Label>
            <Textarea
              placeholder="Optional notes..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={2}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button onClick={handleSave}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

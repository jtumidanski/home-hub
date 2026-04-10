import { useState } from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  useWorkoutExercises,
  useWorkoutThemes,
  useWorkoutRegions,
  useCreateWorkoutExercise,
  useDeleteWorkoutExercise,
  type CreateExerciseAttrs,
} from "@/lib/hooks/api/use-workouts";
import type { WorkoutKind, WeightType } from "@/types/models/workout";
import { toast } from "sonner";

// Catalog screen. List on the left, create dialog opens on the right. Edit
// flows are intentionally simple — name, regions, defaults — and kind +
// weightType are read-only after creation, matching the immutability rule.
export function WorkoutExercisesPage() {
  const exercises = useWorkoutExercises();
  const themes = useWorkoutThemes();
  const regions = useWorkoutRegions();
  const create = useCreateWorkoutExercise();
  const remove = useDeleteWorkoutExercise();

  const [open, setOpen] = useState(false);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Exercises</h2>
        <Button onClick={() => setOpen(true)}>
          <Plus className="h-4 w-4 mr-1" /> New
        </Button>
      </div>

      {exercises.isLoading ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : (
        <Card>
          <CardContent className="p-3">
            {exercises.data?.data.length === 0 ? (
              <p className="text-muted-foreground text-sm">
                No exercises yet. Click "New" to create one.
              </p>
            ) : (
              <ul className="divide-y">
                {exercises.data?.data.map((e) => {
                  const theme = themes.data?.data.find((t) => t.id === e.attributes.themeId);
                  const region = regions.data?.data.find((r) => r.id === e.attributes.regionId);
                  return (
                    <li key={e.id} className="flex items-center justify-between py-2">
                      <div>
                        <p className="font-medium">{e.attributes.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {e.attributes.kind} · {theme?.attributes.name ?? "—"} ·{" "}
                          {region?.attributes.name ?? "—"}
                        </p>
                      </div>
                      <Button
                        size="icon"
                        variant="ghost"
                        onClick={() =>
                          remove.mutate(e.id, { onError: () => toast.error("Delete failed") })
                        }
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </li>
                  );
                })}
              </ul>
            )}
          </CardContent>
        </Card>
      )}

      <ExerciseCreateDialog
        open={open}
        onOpenChange={setOpen}
        themes={themes.data?.data ?? []}
        regions={regions.data?.data ?? []}
        onCreate={(attrs) =>
          create.mutate(attrs, {
            onSuccess: () => {
              toast.success("Exercise created");
              setOpen(false);
            },
            onError: (e) => toast.error((e as Error).message ?? "Create failed"),
          })
        }
      />
    </div>
  );
}

function ExerciseCreateDialog({
  open,
  onOpenChange,
  themes,
  regions,
  onCreate,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  themes: Array<{ id: string; attributes: { name: string } }>;
  regions: Array<{ id: string; attributes: { name: string } }>;
  onCreate: (attrs: CreateExerciseAttrs) => void;
}) {
  const [name, setName] = useState("");
  const [kind, setKind] = useState<WorkoutKind>("strength");
  const [weightType, setWeightType] = useState<WeightType>("free");
  const [themeId, setThemeId] = useState<string>("");
  const [regionId, setRegionId] = useState<string>("");

  const submit = () => {
    if (!name || !themeId || !regionId) {
      toast.error("Name, theme, and region are required");
      return;
    }
    onCreate({
      name,
      kind,
      weightType,
      themeId,
      regionId,
      secondaryRegionIds: [],
      defaults: {},
    });
    setName("");
    setThemeId("");
    setRegionId("");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New exercise</DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          <div>
            <Label htmlFor="ex-name">Name</Label>
            <Input id="ex-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div>
              <Label>Kind</Label>
              <Select value={kind} onValueChange={(v) => setKind(v as WorkoutKind)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="strength">strength</SelectItem>
                  <SelectItem value="isometric">isometric</SelectItem>
                  <SelectItem value="cardio">cardio</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Weight type</Label>
              <Select value={weightType} onValueChange={(v) => setWeightType(v as WeightType)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="free">free</SelectItem>
                  <SelectItem value="bodyweight">bodyweight</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div>
            <Label>Theme</Label>
            <Select value={themeId} onValueChange={(v) => setThemeId(v ?? "")}>
              <SelectTrigger>
                <SelectValue placeholder="Select theme">
                  {themes.find((t) => t.id === themeId)?.attributes.name}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                {themes.map((t) => (
                  <SelectItem key={t.id} value={t.id}>
                    {t.attributes.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label>Primary region</Label>
            <Select value={regionId} onValueChange={(v) => setRegionId(v ?? "")}>
              <SelectTrigger>
                <SelectValue placeholder="Select region">
                  {regions.find((r) => r.id === regionId)?.attributes.name}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                {regions.map((r) => (
                  <SelectItem key={r.id} value={r.id}>
                    {r.attributes.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button onClick={submit}>Create</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

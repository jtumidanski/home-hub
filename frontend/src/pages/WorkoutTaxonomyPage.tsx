import { useState } from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  useWorkoutThemes,
  useCreateWorkoutTheme,
  useDeleteWorkoutTheme,
  useWorkoutRegions,
  useCreateWorkoutRegion,
  useDeleteWorkoutRegion,
} from "@/lib/hooks/api/use-workouts";
import { toast } from "sonner";

// Taxonomy management — themes and regions. Default lists are seeded by the
// backend on first request, so the screen will already show seeded rows when
// the user first visits it.
export function WorkoutTaxonomyPage() {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      <ThemesPanel />
      <RegionsPanel />
    </div>
  );
}

function ThemesPanel() {
  const themes = useWorkoutThemes();
  const create = useCreateWorkoutTheme();
  const remove = useDeleteWorkoutTheme();
  const [name, setName] = useState("");

  return (
    <Card>
      <CardHeader>
        <CardTitle>Themes</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex gap-2">
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="New theme" />
          <Button
            size="icon"
            onClick={() => {
              if (!name.trim()) return;
              create.mutate(
                { name: name.trim() },
                {
                  onSuccess: () => setName(""),
                  onError: (e) => toast.error((e as Error).message ?? "Create failed"),
                },
              );
            }}
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>
        <ul className="divide-y">
          {themes.data?.data.map((t) => (
            <li key={t.id} className="flex items-center justify-between py-2 text-sm">
              <span>{t.attributes.name}</span>
              <Button
                size="icon"
                variant="ghost"
                onClick={() =>
                  remove.mutate(t.id, { onError: () => toast.error("Delete failed") })
                }
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  );
}

function RegionsPanel() {
  const regions = useWorkoutRegions();
  const create = useCreateWorkoutRegion();
  const remove = useDeleteWorkoutRegion();
  const [name, setName] = useState("");

  return (
    <Card>
      <CardHeader>
        <CardTitle>Regions</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex gap-2">
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="New region" />
          <Button
            size="icon"
            onClick={() => {
              if (!name.trim()) return;
              create.mutate(
                { name: name.trim() },
                {
                  onSuccess: () => setName(""),
                  onError: (e) => toast.error((e as Error).message ?? "Create failed"),
                },
              );
            }}
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>
        <ul className="divide-y">
          {regions.data?.data.map((r) => (
            <li key={r.id} className="flex items-center justify-between py-2 text-sm">
              <span>{r.attributes.name}</span>
              <Button
                size="icon"
                variant="ghost"
                onClick={() =>
                  remove.mutate(r.id, { onError: () => toast.error("Delete failed") })
                }
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  );
}

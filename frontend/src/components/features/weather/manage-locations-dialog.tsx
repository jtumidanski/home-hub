import { useState } from "react";
import { toast } from "sonner";
import { Pencil, X, Check, Plus } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { LocationSearch } from "@/components/features/weather/location-search";
import {
  useCreateLocationOfInterest,
  useDeleteLocationOfInterest,
  useLocationsOfInterest,
  useUpdateLocationOfInterest,
} from "@/lib/hooks/api/use-locations-of-interest";
import { createErrorFromUnknown } from "@/lib/api/errors";

const MAX_LOCATIONS = 10;
const MAX_LABEL_LENGTH = 64;

interface ManageLocationsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

interface PendingPlace {
  name: string;
  latitude: number;
  longitude: number;
}

export function ManageLocationsDialog({ open, onOpenChange }: ManageLocationsDialogProps) {
  const { data, isLoading } = useLocationsOfInterest();
  const createLocation = useCreateLocationOfInterest();
  const updateLocation = useUpdateLocationOfInterest();
  const deleteLocation = useDeleteLocationOfInterest();

  const [editingId, setEditingId] = useState<string | null>(null);
  const [editLabel, setEditLabel] = useState("");
  const [pendingPlace, setPendingPlace] = useState<PendingPlace | null>(null);
  const [pendingLabel, setPendingLabel] = useState("");

  const locations = data?.data ?? [];
  const atCap = locations.length >= MAX_LOCATIONS;

  const handleStartRename = (id: string, currentLabel: string | null) => {
    setEditingId(id);
    setEditLabel(currentLabel ?? "");
  };

  const handleConfirmRename = async (id: string) => {
    const trimmed = editLabel.trim();
    if (trimmed.length > MAX_LABEL_LENGTH) {
      toast.error(`Label must be ${MAX_LABEL_LENGTH} characters or fewer`);
      return;
    }
    try {
      await updateLocation.mutateAsync({
        id,
        attrs: { label: trimmed === "" ? null : trimmed },
      });
      setEditingId(null);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to rename location").message);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteLocation.mutateAsync(id);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to remove location").message);
    }
  };

  const handlePlaceSelected = (place: PendingPlace) => {
    setPendingPlace(place);
    setPendingLabel("");
  };

  const handleCancelPending = () => {
    setPendingPlace(null);
    setPendingLabel("");
  };

  const handleConfirmAdd = async () => {
    if (!pendingPlace) return;
    const trimmed = pendingLabel.trim();
    if (trimmed.length > MAX_LABEL_LENGTH) {
      toast.error(`Label must be ${MAX_LABEL_LENGTH} characters or fewer`);
      return;
    }
    try {
      await createLocation.mutateAsync({
        placeName: pendingPlace.name,
        latitude: pendingPlace.latitude,
        longitude: pendingPlace.longitude,
        label: trimmed === "" ? null : trimmed,
      });
      setPendingPlace(null);
      setPendingLabel("");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to save location").message);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Manage saved locations</DialogTitle>
          <DialogDescription>
            Save up to {MAX_LOCATIONS} extra places to view their forecasts. The
            household primary location is always available separately.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3">
          {isLoading ? (
            <p className="text-sm text-muted-foreground">Loading…</p>
          ) : locations.length === 0 ? (
            <p className="text-sm text-muted-foreground">No saved locations yet.</p>
          ) : (
            <div className="space-y-1">
              {locations.map((loc) => (
                <div
                  key={loc.id}
                  className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm"
                >
                  {editingId === loc.id ? (
                    <>
                      <Input
                        value={editLabel}
                        maxLength={MAX_LABEL_LENGTH}
                        onChange={(e) => setEditLabel(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && handleConfirmRename(loc.id)}
                        placeholder={loc.attributes.placeName}
                        className="h-7 text-sm flex-1"
                        autoFocus
                      />
                      <Button
                        size="sm"
                        variant="ghost"
                        className="h-7 w-7 p-0"
                        onClick={() => handleConfirmRename(loc.id)}
                      >
                        <Check className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="h-7 w-7 p-0"
                        onClick={() => setEditingId(null)}
                      >
                        <X className="h-3.5 w-3.5" />
                      </Button>
                    </>
                  ) : (
                    <>
                      <div className="flex-1 min-w-0">
                        <p className="font-medium truncate">
                          {loc.attributes.label ?? loc.attributes.placeName}
                        </p>
                        {loc.attributes.label && (
                          <p className="text-xs text-muted-foreground truncate">
                            {loc.attributes.placeName}
                          </p>
                        )}
                      </div>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="h-7 w-7 p-0"
                        onClick={() =>
                          handleStartRename(loc.id, loc.attributes.label)
                        }
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="h-7 w-7 p-0 text-destructive hover:text-destructive"
                        onClick={() => void handleDelete(loc.id)}
                      >
                        <X className="h-3.5 w-3.5" />
                      </Button>
                    </>
                  )}
                </div>
              ))}
            </div>
          )}

          <div className="border-t pt-3 space-y-2">
            {pendingPlace ? (
              <div className="space-y-2">
                <div className="rounded-md border px-3 py-2 text-sm">
                  <p className="font-medium truncate">{pendingPlace.name}</p>
                </div>
                <div className="space-y-1">
                  <Label htmlFor="pending-label" className="text-xs">
                    Friendly name (optional)
                  </Label>
                  <Input
                    id="pending-label"
                    value={pendingLabel}
                    maxLength={MAX_LABEL_LENGTH}
                    onChange={(e) => setPendingLabel(e.target.value)}
                    placeholder="Mom's house"
                    className="h-8 text-sm"
                  />
                </div>
                <div className="flex gap-2 justify-end">
                  <Button size="sm" variant="outline" onClick={handleCancelPending}>
                    Cancel
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => void handleConfirmAdd()}
                    disabled={createLocation.isPending}
                  >
                    Save
                  </Button>
                </div>
              </div>
            ) : (
              <>
                <Label className="text-xs">Add a location</Label>
                {atCap ? (
                  <p className="flex items-center gap-1 text-xs text-muted-foreground">
                    <Plus className="h-3 w-3" />
                    You've reached the {MAX_LOCATIONS}-location limit. Remove
                    one to add another.
                  </p>
                ) : (
                  <LocationSearch
                    value={null}
                    onSelect={handlePlaceSelected}
                    onClear={() => {}}
                  />
                )}
              </>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

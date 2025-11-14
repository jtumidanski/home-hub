"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createHousehold, updateHousehold, Household } from "@/lib/api/households";
import { toast } from "sonner";
import { GeocodingSearchInput } from "./GeocodingSearchInput";
import { getTimezoneFromCoords, type GeocodingResult } from "@/lib/api/geocoding";

interface HouseholdFormModalProps {
  open: boolean;
  mode: "create" | "edit";
  household?: Household;
  onClose: () => void;
  onSave: () => void;
}

export function HouseholdFormModal({
  open,
  mode,
  household,
  onClose,
  onSave,
}: HouseholdFormModalProps) {
  const [name, setName] = useState("");
  const [latitude, setLatitude] = useState<number | undefined>(undefined);
  const [longitude, setLongitude] = useState<number | undefined>(undefined);
  const [timezone, setTimezone] = useState<string | undefined>(undefined);
  const [saving, setSaving] = useState(false);
  const [locating, setLocating] = useState(false);

  // Reset or populate form when modal opens
  useEffect(() => {
    if (open) {
      if (mode === "edit" && household) {
        setName(household.name);
        setLatitude(household.latitude);
        setLongitude(household.longitude);
        setTimezone(household.timezone);
      } else if (mode === "create") {
        setName("");
        setLatitude(undefined);
        setLongitude(undefined);
        // Auto-detect timezone on create
        setTimezone(Intl.DateTimeFormat().resolvedOptions().timeZone);
      }
    }
  }, [open, mode, household]);

  // Handle geocoding search selection (city search)
  const handleGeocodingSelect = async (result: GeocodingResult) => {
    setLatitude(result.lat);
    setLongitude(result.lon);

    // Get timezone from coordinates
    const tz = await getTimezoneFromCoords(result.lat, result.lon);
    setTimezone(tz);

    toast.success(`Location set to ${result.displayName}`);
  };

  // Get current location using browser Geolocation API
  const handleGetLocation = () => {
    if (!navigator.geolocation) {
      toast.error("Geolocation is not supported by your browser");
      return;
    }

    // Check if running on HTTPS or localhost
    const isSecureContext = window.isSecureContext;
    if (!isSecureContext) {
      toast.error("Geolocation requires HTTPS. Please enter coordinates manually below.");
      return;
    }

    setLocating(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setLatitude(position.coords.latitude);
        setLongitude(position.coords.longitude);
        // Auto-detect timezone
        setTimezone(Intl.DateTimeFormat().resolvedOptions().timeZone);
        setLocating(false);
        toast.success("Location detected successfully");
      },
      (error) => {
        console.error("Error getting location:", error);
        setLocating(false);

        let errorMessage = "Failed to get location. ";
        switch (error.code) {
          case error.PERMISSION_DENIED:
            errorMessage += "Location permission denied. Please enable location access in your browser settings or enter coordinates manually.";
            break;
          case error.POSITION_UNAVAILABLE:
            errorMessage += "Location information unavailable. Please enter coordinates manually.";
            break;
          case error.TIMEOUT:
            errorMessage += "Location request timed out. Please try again or enter coordinates manually.";
            break;
          default:
            errorMessage += "Please enter coordinates manually.";
        }
        toast.error(errorMessage);
      },
      {
        enableHighAccuracy: true,
        timeout: 10000,
        maximumAge: 0,
      }
    );
  };

  // Validation
  const isValid = name.trim().length > 0;
  const hasChanges =
    mode === "create" ||
    (household &&
      (name !== household.name ||
        latitude !== household.latitude ||
        longitude !== household.longitude ||
        timezone !== household.timezone));

  const handleSave = async () => {
    if (!isValid) return;

    try {
      setSaving(true);

      if (mode === "create") {
        await createHousehold({
          name,
          latitude,
          longitude,
          timezone,
        });
        toast.success("Household created successfully");
      } else if (household) {
        await updateHousehold(household.id, {
          name,
          latitude,
          longitude,
          timezone,
        });
        toast.success("Household updated successfully");
      }

      onSave();
    } catch (error) {
      console.error("Failed to save household:", error);
      toast.error(`Failed to ${mode} household`);
    } finally {
      setSaving(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && isValid && hasChanges && !saving) {
      handleSave();
    }
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>
            {mode === "create" ? "Create Household" : "Edit Household"}
          </DialogTitle>
          <DialogDescription>
            {mode === "create"
              ? "Create a new household"
              : "Update household information"}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Name field */}
          <div className="space-y-2">
            <Label htmlFor="name">
              Name <span className="text-red-600">*</span>
            </Label>
            <Input
              id="name"
              placeholder="Enter household name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              onKeyDown={handleKeyDown}
              disabled={saving}
              autoFocus
            />
            {name.trim().length === 0 && name.length > 0 && (
              <p className="text-sm text-red-600">Name cannot be empty</p>
            )}
          </div>

          {/* Location section */}
          <div className="space-y-4">
            <Label>Location (for weather)</Label>

            {/* Method 1: City Search (recommended for development) */}
            <div className="space-y-2">
              <GeocodingSearchInput
                onLocationSelect={handleGeocodingSelect}
                disabled={saving || locating}
              />
            </div>

            {/* OR divider */}
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-neutral-200 dark:border-neutral-800" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-white dark:bg-neutral-950 px-2 text-neutral-500">
                  Or
                </span>
              </div>
            </div>

            {/* Method 2: Browser Geolocation */}
            <div className="space-y-2">
              <Button
                type="button"
                variant="outline"
                onClick={handleGetLocation}
                disabled={saving || locating}
                className="w-full"
              >
                {locating ? "Getting location..." : "Use my current location"}
              </Button>
              <p className="text-xs text-muted-foreground">
                Requires HTTPS and location permission (may not work in development)
              </p>
            </div>
          </div>

          {/* Manual coordinate entry */}
          <div className="space-y-3">
            <div className="space-y-2">
              <Label htmlFor="latitude">Latitude (optional)</Label>
              <Input
                id="latitude"
                type="number"
                step="0.000001"
                placeholder="e.g., 37.7749"
                value={latitude ?? ""}
                onChange={(e) => setLatitude(e.target.value ? parseFloat(e.target.value) : undefined)}
                disabled={saving}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="longitude">Longitude (optional)</Label>
              <Input
                id="longitude"
                type="number"
                step="0.000001"
                placeholder="e.g., -122.4194"
                value={longitude ?? ""}
                onChange={(e) => setLongitude(e.target.value ? parseFloat(e.target.value) : undefined)}
                disabled={saving}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="timezone">Timezone (optional)</Label>
              <Input
                id="timezone"
                placeholder="e.g., America/Los_Angeles"
                value={timezone ?? ""}
                onChange={(e) => setTimezone(e.target.value || undefined)}
                disabled={saving}
              />
              <p className="text-xs text-muted-foreground">
                IANA timezone identifier
              </p>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={saving}>
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={!isValid || !hasChanges || saving}
          >
            {saving
              ? mode === "create"
                ? "Creating..."
                : "Saving..."
              : mode === "create"
                ? "Create"
                : "Save Changes"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

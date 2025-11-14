"use client";

import { useState, useEffect } from "react";
import {
  createDevice,
  getDevicePreferences,
  updateDevicePreferences,
  DeviceType,
  Theme,
  TemperatureUnit,
} from "@/lib/api/devices";
import { listHouseholds, Household } from "@/lib/api/households";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle } from "lucide-react";

interface RegisterDeviceDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function RegisterDeviceDialog({
  open,
  onOpenChange,
  onSuccess,
}: RegisterDeviceDialogProps) {
  const [households, setHouseholds] = useState<Household[]>([]);
  const [loadingHouseholds, setLoadingHouseholds] = useState(true);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [name, setName] = useState("");
  const [type, setType] = useState<DeviceType>("kiosk");
  const [householdId, setHouseholdId] = useState("");
  const [theme, setTheme] = useState<Theme>("dark");
  const [temperatureUnit, setTemperatureUnit] = useState<TemperatureUnit>("household");

  // Fetch households when dialog opens
  useEffect(() => {
    if (open) {
      fetchHouseholds();
    }
  }, [open]);

  // Reset form when dialog closes
  useEffect(() => {
    if (!open) {
      setName("");
      setType("kiosk");
      setHouseholdId("");
      setTheme("dark");
      setTemperatureUnit("household");
      setError(null);
    }
  }, [open]);

  const fetchHouseholds = async () => {
    try {
      setLoadingHouseholds(true);
      const data = await listHouseholds();
      setHouseholds(data);

      // Auto-select first household if only one exists
      if (data.length === 1) {
        setHouseholdId(data[0].id);
      }
    } catch (err) {
      console.error("Failed to fetch households:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch households");
    } finally {
      setLoadingHouseholds(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      setError("Device name is required");
      return;
    }

    if (!householdId) {
      setError("Please select a household");
      return;
    }

    try {
      setCreating(true);
      setError(null);

      // Step 1: Create device
      // Note: Backend uses household from auth context, not from request body
      const device = await createDevice({
        name: name.trim(),
        type,
      });

      // Step 2: Fetch auto-created preferences
      const preferences = await getDevicePreferences(device.id);

      // Step 3: Update preferences if user selected non-default values
      const needsUpdate =
        theme !== preferences.theme ||
        temperatureUnit !== preferences.temperatureUnit;

      if (needsUpdate) {
        await updateDevicePreferences(device.id, {
          theme,
          temperatureUnit,
        });
      }

      // Success - notify parent and close
      onSuccess();
      onOpenChange(false);
    } catch (err) {
      console.error("Failed to create device:", err);
      setError(err instanceof Error ? err.message : "Failed to create device");
    } finally {
      setCreating(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Register New Device</DialogTitle>
          <DialogDescription>
            Add a new kiosk or device to your household
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit}>
          {/* Error Alert */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="grid gap-6 md:grid-cols-2 py-4">
            {/* Device Information Card */}
            <Card>
              <CardHeader>
                <CardTitle>Device Information</CardTitle>
                <CardDescription>Basic device details</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Device Name *</Label>
                  <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g., Kitchen Kiosk, Living Room Display"
                    required
                    autoFocus
                  />
                  <p className="text-sm text-neutral-500 dark:text-neutral-400">
                    Choose a descriptive name to identify this device
                  </p>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="type">Device Type *</Label>
                  <Select value={type} onValueChange={(value) => setType(value as DeviceType)}>
                    <SelectTrigger id="type">
                      <SelectValue placeholder="Select device type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="kiosk">Kiosk</SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-sm text-neutral-500 dark:text-neutral-400">
                    Currently only kiosk devices are supported
                  </p>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="household">Household *</Label>
                  <Select
                    value={householdId}
                    onValueChange={setHouseholdId}
                    disabled={loadingHouseholds}
                  >
                    <SelectTrigger id="household">
                      <SelectValue placeholder="Select household" />
                    </SelectTrigger>
                    <SelectContent>
                      {households.map((household) => (
                        <SelectItem key={household.id} value={household.id}>
                          {household.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <p className="text-sm text-neutral-500 dark:text-neutral-400">
                    Select which household this device belongs to
                  </p>
                </div>
              </CardContent>
            </Card>

            {/* Device Preferences Card */}
            <Card>
              <CardHeader>
                <CardTitle>Initial Preferences</CardTitle>
                <CardDescription>
                  Set the default appearance and behavior
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="theme">Theme</Label>
                  <Select value={theme} onValueChange={(value) => setTheme(value as Theme)}>
                    <SelectTrigger id="theme">
                      <SelectValue placeholder="Select theme" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="light">Light</SelectItem>
                      <SelectItem value="dark">Dark</SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-sm text-neutral-500 dark:text-neutral-400">
                    Choose the color scheme for this device
                  </p>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="temperatureUnit">Temperature Unit</Label>
                  <Select
                    value={temperatureUnit}
                    onValueChange={(value) => setTemperatureUnit(value as TemperatureUnit)}
                  >
                    <SelectTrigger id="temperatureUnit">
                      <SelectValue placeholder="Select unit" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="household">Household Default</SelectItem>
                      <SelectItem value="fahrenheit">Fahrenheit (°F)</SelectItem>
                      <SelectItem value="celsius">Celsius (°C)</SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-sm text-neutral-500 dark:text-neutral-400">
                    Set to "Household Default" to use the household's preference
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={creating}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={creating || !name.trim() || !householdId}
            >
              {creating ? "Creating..." : "Create Device"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

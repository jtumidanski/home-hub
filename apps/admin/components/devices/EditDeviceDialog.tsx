"use client";

import { useEffect, useState } from "react";
import {
  getDevice,
  getDevicePreferences,
  updateDevice,
  updateDevicePreferences,
  Device,
  DevicePreferences,
  Theme,
  TemperatureUnit,
  formatDeviceType,
} from "@/lib/api/devices";
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
import { AlertCircle, Save } from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";

interface EditDeviceDialogProps {
  deviceId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function EditDeviceDialog({
  deviceId,
  open,
  onOpenChange,
  onSuccess,
}: EditDeviceDialogProps) {
  const [device, setDevice] = useState<Device | null>(null);
  const [preferences, setPreferences] = useState<DevicePreferences | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [name, setName] = useState("");
  const [theme, setTheme] = useState<Theme>("dark");
  const [temperatureUnit, setTemperatureUnit] = useState<TemperatureUnit>("household");

  // Fetch device and preferences when deviceId changes
  useEffect(() => {
    if (open && deviceId) {
      fetchData();
    }
  }, [open, deviceId]);

  // Reset state when dialog closes
  useEffect(() => {
    if (!open) {
      setDevice(null);
      setPreferences(null);
      setName("");
      setTheme("dark");
      setTemperatureUnit("household");
      setError(null);
    }
  }, [open]);

  const fetchData = async () => {
    if (!deviceId) return;

    try {
      setLoading(true);
      setError(null);

      const [deviceData, preferencesData] = await Promise.all([
        getDevice(deviceId),
        getDevicePreferences(deviceId),
      ]);

      setDevice(deviceData);
      setPreferences(preferencesData);

      // Set form state
      setName(deviceData.name);
      setTheme(preferencesData.theme);
      setTemperatureUnit(preferencesData.temperatureUnit);
    } catch (err) {
      console.error("Failed to fetch device data:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch device data");
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!device || !preferences || !deviceId) return;

    if (!name.trim()) {
      setError("Device name is required");
      return;
    }

    try {
      setSaving(true);
      setError(null);

      // Determine what changed
      const nameChanged = name !== device.name;
      const preferencesChanged =
        theme !== preferences.theme ||
        temperatureUnit !== preferences.temperatureUnit;

      // Update device name if changed
      if (nameChanged) {
        await updateDevice(deviceId, { name: name.trim() });
      }

      // Update preferences if changed
      if (preferencesChanged) {
        await updateDevicePreferences(deviceId, {
          theme,
          temperatureUnit,
        });
      }

      toast.success("Device updated successfully");
      onSuccess();
      onOpenChange(false);
    } catch (err) {
      console.error("Failed to save device:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to save device";
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setSaving(false);
    }
  };

  const formatDate = (dateString: string): string => {
    if (!dateString) return "—";
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return "Invalid Date";
    return date.toLocaleString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {loading ? "Loading..." : `Edit ${device?.name || "Device"}`}
          </DialogTitle>
          <DialogDescription>
            {device && `${formatDeviceType(device.type)} • ID: ${device.id}`}
          </DialogDescription>
        </DialogHeader>

        {/* Loading State */}
        {loading && (
          <div className="text-center py-12">
            <p className="text-neutral-600 dark:text-neutral-400">Loading device...</p>
          </div>
        )}

        {/* Error State */}
        {error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* Content */}
        {!loading && device && preferences && (
          <form onSubmit={handleSave}>
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
                      placeholder="e.g., Kitchen Kiosk"
                      required
                      autoFocus
                    />
                  </div>

                  <div className="space-y-2">
                    <Label>Device Type</Label>
                    <Input value={formatDeviceType(device.type)} disabled />
                  </div>

                  <div className="space-y-2">
                    <Label>Household</Label>
                    <div className="flex items-center">
                      <Link
                        href={`/households?householdId=${device.householdId}`}
                        className="text-blue-600 hover:underline dark:text-blue-400"
                        onClick={(e) => e.stopPropagation()}
                      >
                        View Household
                      </Link>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label>Created</Label>
                    <Input value={formatDate(device.created_at)} disabled />
                  </div>

                  <div className="space-y-2">
                    <Label>Last Updated</Label>
                    <Input value={formatDate(device.updated_at)} disabled />
                  </div>
                </CardContent>
              </Card>

              {/* Device Preferences Card */}
              <Card>
                <CardHeader>
                  <CardTitle>Device Preferences</CardTitle>
                  <CardDescription>
                    Customize the appearance and behavior
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
                        <SelectItem value="F">Fahrenheit (°F)</SelectItem>
                        <SelectItem value="C">Celsius (°C)</SelectItem>
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
                disabled={saving}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={saving || !name.trim()}>
                <Save className="h-4 w-4 mr-2" />
                {saving ? "Saving..." : "Save Changes"}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}

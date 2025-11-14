"use client";

import { useEffect, useState } from "react";
import {
  getDevice,
  getDevicePreferences,
  Device,
  DevicePreferences,
  formatDeviceType,
  formatTheme,
  formatTemperatureUnit,
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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, Edit } from "lucide-react";
import Link from "next/link";

interface ViewDeviceDialogProps {
  deviceId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onEdit: (deviceId: string) => void;
}

export function ViewDeviceDialog({
  deviceId,
  open,
  onOpenChange,
  onEdit,
}: ViewDeviceDialogProps) {
  const [device, setDevice] = useState<Device | null>(null);
  const [preferences, setPreferences] = useState<DevicePreferences | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
    } catch (err) {
      console.error("Failed to fetch device data:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch device data");
    } finally {
      setLoading(false);
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

  const handleEdit = () => {
    if (deviceId) {
      onEdit(deviceId);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {loading ? "Loading..." : device?.name || "Device Details"}
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
        {!loading && !error && device && preferences && (
          <div className="grid gap-6 md:grid-cols-2 py-4">
            {/* Device Information Card */}
            <Card>
              <CardHeader>
                <CardTitle>Device Information</CardTitle>
                <CardDescription>Basic device details</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Device Name</Label>
                  <Input value={device.name} disabled />
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
                  Current appearance and behavior settings
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Theme</Label>
                  <Input value={formatTheme(preferences.theme)} disabled />
                  <p className="text-sm text-neutral-500 dark:text-neutral-400">
                    The color scheme for this device
                  </p>
                </div>

                <div className="space-y-2">
                  <Label>Temperature Unit</Label>
                  <Input value={formatTemperatureUnit(preferences.temperatureUnit)} disabled />
                  <p className="text-sm text-neutral-500 dark:text-neutral-400">
                    Temperature display preference
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
          {device && (
            <Button onClick={handleEdit}>
              <Edit className="h-4 w-4 mr-2" />
              Edit
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

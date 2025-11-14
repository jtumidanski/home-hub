"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getDevices, createDevice, Device } from "@/lib/api/devices";
import { Monitor, Plus, AlertCircle } from "lucide-react";
import { Card } from "@/app/components/ui/Card";

export default function SelectDevicePage() {
  const router = useRouter();
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [newDeviceName, setNewDeviceName] = useState("");
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    fetchDevices();
  }, []);

  const fetchDevices = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getDevices();
      // Filter only kiosk devices
      const kioskDevices = data.filter((d) => d.type === "kiosk");
      setDevices(kioskDevices);
    } catch (err) {
      console.error("Failed to fetch devices:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch devices");
    } finally {
      setLoading(false);
    }
  };

  const handleDeviceClick = (deviceId: string) => {
    // Navigate to kiosk with device ID
    router.push(`/?kioskId=${deviceId}`);
  };

  const handleCreateDevice = async () => {
    if (!newDeviceName.trim()) {
      setError("Please enter a device name");
      return;
    }

    try {
      setCreating(true);
      setError(null);

      // Household ID is automatically extracted from auth context by the backend
      const device = await createDevice({
        name: newDeviceName.trim(),
        type: "kiosk",
      });

      setCreateDialogOpen(false);
      setNewDeviceName("");

      // Navigate to kiosk with new device ID
      router.push(`/?kioskId=${device.id}`);
    } catch (err) {
      console.error("Failed to create device:", err);
      setError(err instanceof Error ? err.message : "Failed to create device");
    } finally {
      setCreating(false);
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="text-center">
          <Monitor className="mx-auto h-12 w-12 text-muted-foreground animate-pulse" />
          <p className="mt-4 text-muted-foreground">
            Loading devices...
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-6">
      <div className="w-full max-w-4xl space-y-6">
        {/* Header */}
        <div className="text-center">
          <h1 className="text-4xl font-bold tracking-tight">
            Select Your Kiosk
          </h1>
          <p className="mt-2 text-muted-foreground">
            Choose a kiosk to view the dashboard or register a new one
          </p>
        </div>

        {/* Error Alert */}
        {error && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <div className="flex items-center gap-3">
              <AlertCircle className="h-5 w-5 text-red-600 dark:text-red-400" />
              <div>
                <h3 className="font-semibold text-red-900 dark:text-red-100">Error</h3>
                <p className="text-sm text-red-700 dark:text-red-300">{error}</p>
              </div>
            </div>
          </div>
        )}

        {/* Devices Grid */}
        {devices.length > 0 ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {devices.map((device) => (
              <div
                key={device.id}
                className="bg-card text-card-foreground rounded-lg shadow-lg border border-border cursor-pointer transition-all hover:shadow-xl hover:scale-105 p-6"
                onClick={() => handleDeviceClick(device.id)}
              >
                <div className="flex items-center gap-3">
                  <Monitor className="h-8 w-8 text-primary" />
                  <div>
                    <h3 className="text-xl font-semibold">
                      {device.name}
                    </h3>
                    <p className="text-sm text-muted-foreground">
                      Kiosk Device
                    </p>
                  </div>
                </div>
              </div>
            ))}

            {/* Register New Device Card */}
            <div
              className="bg-card text-card-foreground rounded-lg shadow-lg border-2 border-dashed border-border cursor-pointer transition-all hover:shadow-xl hover:scale-105 p-6"
              onClick={() => setCreateDialogOpen(true)}
            >
              <div className="flex items-center gap-3">
                <Plus className="h-8 w-8 text-muted-foreground" />
                <div>
                  <h3 className="text-xl font-semibold">
                    Register New Kiosk
                  </h3>
                  <p className="text-sm text-muted-foreground">
                    Add a new device
                  </p>
                </div>
              </div>
            </div>
          </div>
        ) : (
          // Empty state - no devices
          <Card>
            <div className="flex flex-col items-center justify-center py-12">
              <Monitor className="h-16 w-16 text-muted-foreground" />
              <h3 className="mt-4 text-xl font-semibold">
                No Kiosks Found
              </h3>
              <p className="mt-2 text-center text-muted-foreground">
                Register your first kiosk to get started
              </p>
              <button
                className="mt-6 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors flex items-center gap-2"
                onClick={() => setCreateDialogOpen(true)}
              >
                <Plus className="h-4 w-4" />
                Register New Kiosk
              </button>
            </div>
          </Card>
        )}

        {/* Create Device Dialog */}
        {createDialogOpen && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
            <div className="bg-card text-card-foreground rounded-lg shadow-xl max-w-md w-full p-6 border border-border">
              <h2 className="text-2xl font-bold mb-2">
                Register New Kiosk
              </h2>
              <p className="text-sm text-muted-foreground mb-6">
                Enter a name for this kiosk device
              </p>

              <div className="space-y-4">
                <div>
                  <label
                    htmlFor="deviceName"
                    className="block text-sm font-medium mb-2"
                  >
                    Kiosk Name
                  </label>
                  <input
                    id="deviceName"
                    type="text"
                    value={newDeviceName}
                    onChange={(e) => setNewDeviceName(e.target.value)}
                    placeholder="e.g., Kitchen Kiosk, Living Room Display"
                    className="w-full px-3 py-2 border border-input rounded-lg bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                    onKeyDown={(e) => {
                      if (e.key === "Enter") {
                        handleCreateDevice();
                      }
                    }}
                  />
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={() => {
                    setCreateDialogOpen(false);
                    setNewDeviceName("");
                  }}
                  disabled={creating}
                  className="flex-1 px-4 py-2 border border-border bg-secondary text-secondary-foreground rounded-lg hover:bg-secondary/80 transition-colors disabled:opacity-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreateDevice}
                  disabled={creating || !newDeviceName.trim()}
                  className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {creating ? "Creating..." : "Register"}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

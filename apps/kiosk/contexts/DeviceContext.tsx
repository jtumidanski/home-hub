"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import {
  getDevice,
  getDevicePreferences,
  Device,
  DevicePreferences,
  TemperatureUnit,
  Theme,
} from "@/lib/api/devices";

interface DeviceContextType {
  deviceId: string | null;
  device: Device | null;
  preferences: DevicePreferences | null;
  loading: boolean;
  error: string | null;
  theme: Theme;
  temperatureUnit: TemperatureUnit;
  resolvedTemperatureUnit: "fahrenheit" | "celsius";
  refresh: () => Promise<void>;
}

const DeviceContext = createContext<DeviceContextType | undefined>(undefined);

interface DeviceProviderProps {
  children: React.ReactNode;
  deviceId: string | null; // Passed from parent component
  householdTemperatureUnit?: "fahrenheit" | "celsius"; // Passed from household data
}

export function DeviceProvider({ children, deviceId, householdTemperatureUnit = "fahrenheit" }: DeviceProviderProps) {

  const [device, setDevice] = useState<Device | null>(null);
  const [preferences, setPreferences] = useState<DevicePreferences | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchDeviceData = async () => {
    if (!deviceId) {
      setDevice(null);
      setPreferences(null);
      return;
    }

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

  useEffect(() => {
    fetchDeviceData();
  }, [deviceId]);

  // Temperature unit resolution: device preference → household → system default ("fahrenheit")
  // If device preference is "household", use household's temperature unit
  // If device preference is explicit ("fahrenheit" or "celsius"), use that
  // If no device preferences loaded yet, use household default
  const resolvedTemperatureUnit: "fahrenheit" | "celsius" = preferences
    ? preferences.temperatureUnit === "household"
      ? householdTemperatureUnit
      : preferences.temperatureUnit
    : householdTemperatureUnit;

  // Theme resolution: device preference → system default ("dark")
  // Note: Theme is device-level only. Households do not have theme preferences.
  const resolvedTheme: Theme = preferences?.theme || "dark";

  // Apply theme whenever it changes
  useEffect(() => {
    applyTheme(resolvedTheme);
  }, [resolvedTheme]);

  // Apply theme by adding/removing 'dark' class on document root
  // Also cache theme in localStorage to prevent FOUC on next page load
  const applyTheme = (theme: Theme) => {
    if (typeof window !== "undefined") {
      const htmlElement = document.documentElement;

      // Apply theme class
      if (theme === "dark") {
        htmlElement.classList.add("dark");
      } else {
        htmlElement.classList.remove("dark");
      }

      // Cache theme in localStorage for instant application on next load
      try {
        localStorage.setItem("kiosk-theme", theme);
      } catch (e) {
        // localStorage might not be available, silently fail
        console.warn("Failed to cache theme in localStorage:", e);
      }
    }
  };

  const value: DeviceContextType = {
    deviceId,
    device,
    preferences,
    loading,
    error,
    theme: resolvedTheme,
    temperatureUnit: preferences?.temperatureUnit || "household",
    resolvedTemperatureUnit,
    refresh: fetchDeviceData,
  };

  return <DeviceContext.Provider value={value}>{children}</DeviceContext.Provider>;
}

export function useDevice() {
  const context = useContext(DeviceContext);
  if (context === undefined) {
    throw new Error("useDevice must be used within a DeviceProvider");
  }
  return context;
}

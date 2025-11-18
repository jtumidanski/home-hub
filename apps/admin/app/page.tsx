"use client";

import { useAuth } from "@/lib/auth";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Building2, Users, Monitor, Activity } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { getHouseholdCount, getUserCount } from "@/lib/api/users";
import { getDevices } from "@/lib/api/devices";

export default function Home() {
  const { user, loading } = useAuth();
  const [householdCount, setHouseholdCount] = useState<number | null>(null);
  const [userCount, setUserCount] = useState<number | null>(null);
  const [deviceCount, setDeviceCount] = useState<number | null>(null);
  const [countsLoading, setCountsLoading] = useState(true);
  const [countsError, setCountsError] = useState<string | null>(null);

  // Fetch counts when component mounts and user is authenticated
  useEffect(() => {
    if (!user) {
      setCountsLoading(false);
      return;
    }

    const fetchCounts = async () => {
      try {
        setCountsLoading(true);
        setCountsError(null);

        const [households, users, devices] = await Promise.all([
          getHouseholdCount(),
          getUserCount(),
          getDevices(),
        ]);

        setHouseholdCount(households);
        setUserCount(users);
        setDeviceCount(devices.length);
      } catch (error) {
        console.error("Failed to fetch counts:", error);
        setCountsError(
          error instanceof Error ? error.message : "Failed to fetch data",
        );
      } finally {
        setCountsLoading(false);
      }
    };

    fetchCounts();
  }, [user]);

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-10 w-64 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse" />
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="h-32 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse"
            />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight">Admin Portal</h1>
        <p className="text-neutral-600 dark:text-neutral-400">
          {user
            ? `Hello, ${user.displayName}!`
            : "Please sign in to manage your households, users, and services."}
        </p>
      </div>

      {/* Not authenticated message */}
      {!user && (
        <Card className="border-orange-200 bg-orange-50 dark:border-orange-800 dark:bg-orange-950">
          <CardHeader>
            <CardTitle className="text-orange-900 dark:text-orange-100">
              Authentication Required
            </CardTitle>
            <CardDescription className="text-orange-700 dark:text-orange-300">
              You need to sign in to access the admin portal
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-orange-800 dark:text-orange-200">
              Click the "Sign In" button in the top-right corner to authenticate
              with Google.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Stats Cards */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Households
            </CardTitle>
            <Link href="/households">
              <Building2 className="h-4 w-4 text-muted-foreground hover:text-primary transition-colors cursor-pointer" />
            </Link>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {countsLoading ? (
                <span className="text-neutral-400">...</span>
              ) : countsError ? (
                <span className="text-red-600 text-sm">Error</span>
              ) : householdCount !== null ? (
                householdCount
              ) : (
                "—"
              )}
            </div>
            <p className="text-xs text-muted-foreground">
              {countsError ? countsError : "Registered households"}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Users</CardTitle>
            <Link href="/users">
              <Users className="h-4 w-4 text-muted-foreground hover:text-primary transition-colors cursor-pointer" />
            </Link>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {countsLoading ? (
                <span className="text-neutral-400">...</span>
              ) : countsError ? (
                <span className="text-red-600 text-sm">Error</span>
              ) : userCount !== null ? (
                userCount
              ) : (
                "—"
              )}
            </div>
            <p className="text-xs text-muted-foreground">
              {countsError ? "Unable to fetch data" : "Total registered users"}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Registered Devices
            </CardTitle>
            <Link href="/devices">
              <Monitor className="h-4 w-4 text-muted-foreground hover:text-primary transition-colors cursor-pointer" />
            </Link>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {countsLoading ? (
                <span className="text-neutral-400">...</span>
              ) : countsError ? (
                <span className="text-red-600 text-sm">Error</span>
              ) : deviceCount !== null ? (
                deviceCount
              ) : (
                "—"
              )}
            </div>
            <p className="text-xs text-muted-foreground">
              {countsError ? "Unable to fetch data" : "Registered devices"}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">System Status</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">Online</div>
            <p className="text-xs text-muted-foreground">
              All services operational
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
          <CardDescription>Common administrative tasks</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          <Button variant="outline" asChild>
            <Link href="/households">View Households</Link>
          </Button>
          <Button variant="outline" asChild>
            <Link href="/users">Manage Users</Link>
          </Button>
          <Button variant="outline" asChild>
            <Link href="/devices">View Devices</Link>
          </Button>
          <Button variant="outline" asChild>
            <Link href="/system/logs">View Logs</Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

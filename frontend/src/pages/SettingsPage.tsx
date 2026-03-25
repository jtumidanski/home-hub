import { useAuth } from "@/components/providers/auth-provider";
import { useThemeToggle } from "@/lib/hooks/use-theme-toggle";
import { Button } from "@/components/ui/button";
import { ErrorCard } from "@/components/common/error-card";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Moon, Sun } from "lucide-react";

function SettingsPageSkeleton() {
  return (
    <div className="p-6 space-y-6" role="status" aria-label="Loading">
      <Skeleton className="h-8 w-32" />
      <Skeleton className="h-40 w-full" />
      <Skeleton className="h-24 w-full" />
    </div>
  );
}

export function SettingsPage() {
  const { user, appContext, isLoading } = useAuth();
  const { theme, toggleTheme } = useThemeToggle();

  if (isLoading) {
    return <SettingsPageSkeleton />;
  }

  if (!user || !appContext) {
    return (
      <div className="p-6">
        <ErrorCard message="Failed to load settings. Try refreshing the page." />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-semibold">Settings</h1>

      <Card>
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Your account information</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2">
          <div><span className="text-sm font-medium">Name:</span> <span className="text-sm">{user.attributes.displayName}</span></div>
          <div><span className="text-sm font-medium">Email:</span> <span className="text-sm">{user.attributes.email}</span></div>
          <div><span className="text-sm font-medium">Role:</span> <span className="text-sm">{appContext.attributes.resolvedRole}</span></div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
          <CardDescription>Customize how Home Hub looks</CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" onClick={toggleTheme}>
            {theme === "light" ? <Moon className="mr-2 h-4 w-4" /> : <Sun className="mr-2 h-4 w-4" />}
            {theme === "light" ? "Switch to Dark Mode" : "Switch to Light Mode"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

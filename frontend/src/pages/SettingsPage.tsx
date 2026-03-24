import { useAuth } from "@/components/providers/auth-provider";
import { useTheme } from "@/components/providers/theme-provider";
import { accountService } from "@/services/api/account";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Moon, Sun } from "lucide-react";

export function SettingsPage() {
  const { user, appContext } = useAuth();
  const { theme, setTheme } = useTheme();

  const handleThemeToggle = async () => {
    const newTheme = theme === "light" ? "dark" : "light";
    setTheme(newTheme);
    if (appContext?.relationships?.preference?.data?.id) {
      await accountService.updatePreferenceTheme(
        appContext.relationships.preference.data.id,
        newTheme
      );
    }
  };

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-semibold">Settings</h1>

      <Card>
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Your account information</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2">
          <div><span className="text-sm font-medium">Name:</span> <span className="text-sm">{user?.attributes.displayName}</span></div>
          <div><span className="text-sm font-medium">Email:</span> <span className="text-sm">{user?.attributes.email}</span></div>
          <div><span className="text-sm font-medium">Role:</span> <span className="text-sm">{appContext?.attributes.resolvedRole}</span></div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
          <CardDescription>Customize how Home Hub looks</CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" onClick={handleThemeToggle}>
            {theme === "light" ? <Moon className="mr-2 h-4 w-4" /> : <Sun className="mr-2 h-4 w-4" />}
            {theme === "light" ? "Switch to Dark Mode" : "Switch to Light Mode"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

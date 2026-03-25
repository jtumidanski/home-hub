import { useTheme } from "@/components/providers/theme-provider";
import { useUpdatePreferenceTheme } from "@/lib/hooks/api/use-context";
import { useAuth } from "@/components/providers/auth-provider";
import { useTenant } from "@/context/tenant-context";

export function useThemeToggle() {
  const { theme, setTheme } = useTheme();
  const { appContext } = useAuth();
  const { tenant } = useTenant();
  const updatePreference = useUpdatePreferenceTheme();

  const toggleTheme = () => {
    const newTheme = theme === "light" ? "dark" : "light";
    setTheme(newTheme);

    const preferenceId = appContext?.relationships?.preference?.data?.id;
    if (tenant && preferenceId) {
      updatePreference.mutate({ tenant, preferenceId, theme: newTheme });
    }
  };

  return { theme, toggleTheme };
}

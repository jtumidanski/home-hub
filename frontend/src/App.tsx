import { toasterStyle, toastOptions } from "@/lib/toast-options";
import { BrowserRouter } from "react-router-dom";
import { Toaster } from "sonner";
import { QueryProvider } from "@/components/providers/query-provider";
import { ThemeProvider, useTheme } from "@/components/providers/theme-provider";
import { AuthProvider } from "@/components/providers/auth-provider";
import { TenantProvider } from "@/context/tenant-context";
import { TooltipProvider } from "@/components/ui/tooltip";
import { AppRoutes } from "@/app-routes";

function ThemedToaster() {
  const { theme } = useTheme();
  return (
    <Toaster
      theme={theme}
      closeButton
      style={toasterStyle}
      toastOptions={toastOptions}
    />
  );
}

export function App() {
  return (
    <BrowserRouter>
      <QueryProvider>
        <ThemeProvider>
          <AuthProvider>
            <TenantProvider>
              <TooltipProvider delay={200}>
                <ThemedToaster />
                <AppRoutes />
              </TooltipProvider>
            </TenantProvider>
          </AuthProvider>
        </ThemeProvider>
      </QueryProvider>
    </BrowserRouter>
  );
}

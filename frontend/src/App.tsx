import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Toaster } from "sonner";
import { QueryProvider } from "@/components/providers/query-provider";
import { ThemeProvider } from "@/components/providers/theme-provider";
import { AuthProvider } from "@/components/providers/auth-provider";
import { TenantProvider } from "@/context/tenant-context";
import { ProtectedRoute } from "@/components/features/navigation/protected-route";
import { AppShell } from "@/components/features/navigation/app-shell";
import { LoginPage } from "@/pages/LoginPage";
import { OnboardingPage } from "@/pages/OnboardingPage";
import { DashboardPage } from "@/pages/DashboardPage";
import { TasksPage } from "@/pages/TasksPage";
import { RemindersPage } from "@/pages/RemindersPage";
import { SettingsPage } from "@/pages/SettingsPage";
import { HouseholdsPage } from "@/pages/HouseholdsPage";
import { WeatherPage } from "@/pages/WeatherPage";
import { Error404Page } from "@/components/common/error-page";

export function App() {
  return (
    <BrowserRouter>
      <QueryProvider>
        <ThemeProvider>
          <AuthProvider>
            <TenantProvider>
              <Toaster richColors closeButton />
              <Routes>
                <Route path="/" element={<Navigate to="/app" replace />} />
                <Route path="/login" element={<LoginPage />} />
                <Route path="/onboarding" element={<OnboardingPage />} />
                <Route
                  path="/app"
                  element={
                    <ProtectedRoute>
                      <AppShell />
                    </ProtectedRoute>
                  }
                >
                  <Route index element={<DashboardPage />} />
                  <Route path="tasks" element={<TasksPage />} />
                  <Route path="reminders" element={<RemindersPage />} />
                  <Route path="settings" element={<SettingsPage />} />
                  <Route path="weather" element={<WeatherPage />} />
                  <Route path="households" element={<HouseholdsPage />} />
                </Route>
                <Route path="*" element={<Error404Page />} />
              </Routes>
            </TenantProvider>
          </AuthProvider>
        </ThemeProvider>
      </QueryProvider>
    </BrowserRouter>
  );
}

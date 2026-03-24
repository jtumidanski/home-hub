import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { QueryProvider } from "@/components/providers/query-provider";
import { ThemeProvider } from "@/components/providers/theme-provider";
import { AuthProvider } from "@/components/providers/auth-provider";
import { ProtectedRoute } from "@/components/features/protected-route";
import { AppShell } from "@/components/features/app-shell";
import { LoginPage } from "@/pages/LoginPage";
import { OnboardingPage } from "@/pages/OnboardingPage";
import { DashboardPage } from "@/pages/DashboardPage";

export function App() {
  return (
    <BrowserRouter>
      <QueryProvider>
        <ThemeProvider>
          <AuthProvider>
            <Routes>
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
                <Route path="tasks" element={<div className="p-6"><h1 className="text-2xl font-semibold">Tasks</h1></div>} />
                <Route path="reminders" element={<div className="p-6"><h1 className="text-2xl font-semibold">Reminders</h1></div>} />
                <Route path="settings" element={<div className="p-6"><h1 className="text-2xl font-semibold">Settings</h1></div>} />
              </Route>
              <Route path="*" element={<Navigate to="/app" replace />} />
            </Routes>
          </AuthProvider>
        </ThemeProvider>
      </QueryProvider>
    </BrowserRouter>
  );
}

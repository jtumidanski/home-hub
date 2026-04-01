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
import { HouseholdMembersPage } from "@/pages/HouseholdMembersPage";
import { WeatherPage } from "@/pages/WeatherPage";
import { RecipesPage } from "@/pages/RecipesPage";
import { RecipeDetailPage } from "@/pages/RecipeDetailPage";
import { RecipeFormPage } from "@/pages/RecipeFormPage";
import { IngredientsPage } from "@/pages/IngredientsPage";
import { IngredientDetailPage } from "@/pages/IngredientDetailPage";
import { MealsPage } from "@/pages/MealsPage";
import { ShoppingListsPage } from "@/pages/ShoppingListsPage";
import { ShoppingListDetailPage } from "@/pages/ShoppingListDetailPage";
import { CalendarPage } from "@/pages/CalendarPage";
import { PackagesPage } from "@/pages/PackagesPage";
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
                  <Route path="calendar" element={<CalendarPage />} />
                  <Route path="packages" element={<PackagesPage />} />
                  <Route path="settings" element={<SettingsPage />} />
                  <Route path="weather" element={<WeatherPage />} />
                  <Route path="recipes" element={<RecipesPage />} />
                  <Route path="recipes/new" element={<RecipeFormPage />} />
                  <Route path="recipes/:id" element={<RecipeDetailPage />} />
                  <Route path="recipes/:id/edit" element={<RecipeFormPage />} />
                  <Route path="meals" element={<MealsPage />} />
                  <Route path="shopping" element={<ShoppingListsPage />} />
                  <Route path="shopping/:id" element={<ShoppingListDetailPage />} />
                  <Route path="ingredients" element={<IngredientsPage />} />
                  <Route path="ingredients/:id" element={<IngredientDetailPage />} />
                  <Route path="households" element={<HouseholdsPage />} />
                  <Route path="households/:id/members" element={<HouseholdMembersPage />} />
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

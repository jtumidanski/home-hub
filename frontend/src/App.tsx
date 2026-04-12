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
import { DataRetentionPage } from "@/pages/DataRetentionPage";
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
import { WishListPage } from "@/pages/WishListPage";
import { CalendarPage } from "@/pages/CalendarPage";
import { PackagesPage } from "@/pages/PackagesPage";
import { TrackerPage } from "@/pages/TrackerPage";
import { WorkoutShell } from "@/components/features/workout/workout-shell";
import { WorkoutTodayPage } from "@/pages/WorkoutTodayPage";
import { WorkoutWeekPage } from "@/pages/WorkoutWeekPage";
import { WorkoutExercisesPage } from "@/pages/WorkoutExercisesPage";
import { WorkoutTaxonomyPage } from "@/pages/WorkoutTaxonomyPage";
import { WorkoutSummaryPage } from "@/pages/WorkoutSummaryPage";
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
                  <Route path="habits" element={<TrackerPage />} />
                  <Route path="tracker" element={<Navigate to="/app/habits" replace />} />
                  <Route path="workouts" element={<WorkoutShell />}>
                    <Route index element={<Navigate to="today" replace />} />
                    <Route path="today" element={<WorkoutTodayPage />} />
                    <Route path="week" element={<WorkoutWeekPage />} />
                    <Route path="week/:weekStart" element={<WorkoutWeekPage />} />
                    <Route path="exercises" element={<WorkoutExercisesPage />} />
                    <Route path="taxonomy" element={<WorkoutTaxonomyPage />} />
                    <Route path="summary" element={<WorkoutSummaryPage />} />
                    <Route path="summary/:weekStart" element={<WorkoutSummaryPage />} />
                  </Route>
                  <Route path="settings" element={<SettingsPage />} />
                  <Route path="settings/data-retention" element={<DataRetentionPage />} />
                  <Route path="weather" element={<WeatherPage />} />
                  <Route path="recipes" element={<RecipesPage />} />
                  <Route path="recipes/new" element={<RecipeFormPage />} />
                  <Route path="recipes/:id" element={<RecipeDetailPage />} />
                  <Route path="recipes/:id/edit" element={<RecipeFormPage />} />
                  <Route path="meals" element={<MealsPage />} />
                  <Route path="shopping" element={<Navigate to="/app/shopping/grocery" replace />} />
                  <Route path="shopping/grocery" element={<ShoppingListsPage />} />
                  <Route path="shopping/grocery/:id" element={<ShoppingListDetailPage />} />
                  <Route path="shopping/wish-list" element={<WishListPage />} />
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

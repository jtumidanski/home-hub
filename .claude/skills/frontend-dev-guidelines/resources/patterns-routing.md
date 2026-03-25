# Routing & Pages Patterns

## Overview

Home Hub UI uses React Router for client-side routing. All pages are client-rendered — this is a SPA that relies on React hooks and browser APIs.

## Route Structure

```
pages/
├── LoginPage.tsx            # Login page
├── OnboardingPage.tsx       # Onboarding flow
├── DashboardPage.tsx        # Dashboard home
├── TasksPage.tsx            # Task list
├── RemindersPage.tsx        # Reminder list
├── SettingsPage.tsx         # User settings
├── HouseholdsPage.tsx       # Household management
└── NotFoundPage.tsx         # 404 handler
```

## Route Configuration

```tsx
// App.tsx or routes.tsx
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/onboarding" element={<OnboardingPage />} />
        <Route element={<AppLayout />}>
          <Route path="/" element={<Navigate to="/app" replace />} />
          <Route path="/app" element={<DashboardPage />} />
          <Route path="/app/tasks" element={<TasksPage />} />
          <Route path="/app/reminders" element={<RemindersPage />} />
          <Route path="/app/settings" element={<SettingsPage />} />
          <Route path="/app/households" element={<HouseholdsPage />} />
        </Route>
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  );
}
```

## List Page Pattern

```tsx
import { useState, useEffect, useCallback } from "react";
import { useTenant } from "@/context/tenant-context";
import { DataTable } from "@/components/data-table";

export function TasksPage() {
  const { activeTenant } = useTenant();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchTasks = useCallback(async () => {
    if (!activeTenant) return;
    setLoading(true);
    try {
      const data = await tasksService.getAllTasks(activeTenant);
      setTasks(data);
    } catch (err) {
      const errorInfo = createErrorFromUnknown(err, "Failed to fetch tasks");
      toast.error(errorInfo.message);
    } finally {
      setLoading(false);
    }
  }, [activeTenant]);

  useEffect(() => { fetchTasks(); }, [fetchTasks]);

  if (loading) return <TasksPageSkeleton />;

  return (
    <div className="flex flex-col gap-4 p-4">
      <DataTable columns={columns} data={tasks} onRefresh={fetchTasks} />
    </div>
  );
}
```

## Detail Page Pattern

```tsx
import { useParams, useNavigate } from "react-router-dom";

export function TaskDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { activeTenant } = useTenant();

  // ... fetch and display logic
}
```

## Root Layout

The root layout wraps the app with providers:

```tsx
// App.tsx
export function App() {
  return (
    <TenantProvider>
      <QueryProvider>
        <ThemeProvider defaultTheme="dark">
          <BrowserRouter>
            <Routes>
              {/* routes */}
            </Routes>
          </BrowserRouter>
          <Toaster />
        </ThemeProvider>
      </QueryProvider>
    </TenantProvider>
  );
}
```

## Navigation Patterns

- **Sidebar:** Static navigation groups defined in `app-sidebar.tsx`
- **Navigation:** `useNavigate()` hook from React Router
- **Back navigation:** `navigate(-1)` or `navigate("/parent-route")`
- **Post-action redirect:** `navigate("/resource/" + id)` after success

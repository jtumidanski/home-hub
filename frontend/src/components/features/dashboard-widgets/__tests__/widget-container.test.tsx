import { describe, it, expect, vi } from "vitest";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { widgetRegistry } from "@/lib/dashboard/widget-registry";

const loaded = { data: { data: [] }, isLoading: false, isError: false };

vi.mock("@/context/tenant-context", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useTenant: () => ({ household: { attributes: { timezone: "UTC", units: "imperial", latitude: 1, longitude: 2, locationName: "Home" } } }),
}));
vi.mock("@/lib/hooks/api/use-tasks", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useTasks: () => loaded,
  useTaskSummary: () => ({
    data: { data: { attributes: { pendingCount: 3, overdueCount: 1, completedTodayCount: 2 } } },
    isLoading: false,
  }),
}));
vi.mock("@/lib/hooks/api/use-reminders", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useReminders: () => loaded,
  useReminderSummary: () => ({
    data: { data: { attributes: { dueNowCount: 1, snoozedCount: 0, upcomingCount: 2 } } },
    isLoading: false,
  }),
}));
vi.mock("@/lib/hooks/api/use-calendar", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useCalendarEvents: () => loaded,
}));
vi.mock("@/lib/hooks/api/use-weather", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useWeatherForecast: () => loaded,
  useCurrentWeather: () => ({
    data: {
      data: {
        attributes: {
          temperature: 70, temperatureUnit: "F", highTemperature: 75, lowTemperature: 60,
          summary: "Sunny", icon: "sunny", fetchedAt: new Date().toISOString(),
        },
      },
    },
    isLoading: false,
    isError: false,
  }),
}));
vi.mock("@/lib/hooks/api/use-trackers", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useTrackerToday: () => loaded,
}));
vi.mock("@/lib/hooks/api/use-workouts", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useWorkoutToday: () => ({ data: { data: { attributes: { isRestDay: false, items: [] } } }, isLoading: false, isError: false }),
}));
vi.mock("@/lib/hooks/api/use-packages", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  usePackageSummary: () => ({
    data: { data: { attributes: { arrivingTodayCount: 1, inTransitCount: 2, exceptionCount: 0 } } },
    isLoading: false,
    isError: false,
  }),
}));
vi.mock("@/lib/hooks/api/use-meals", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  usePlans: () => loaded,
  usePlan: () => ({ data: undefined, isLoading: false, isError: false }),
}));

function renderWidget(type: string) {
  const def = widgetRegistry.find((w) => w.type === type)!;
  const Component = def.component;
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <Component config={def.defaultConfig} />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("dashboard widget container queries", () => {
  it.each(widgetRegistry.map((w) => [w.type]))("%s root is an @container", (type) => {
    const { container } = renderWidget(type as string);
    expect(container.firstElementChild).toHaveClass("@container");
  });

  it.each(widgetRegistry.map((w) => [w.type]))("%s title is font-heading and container-scaled", (type) => {
    const { container } = renderWidget(type as string);
    const titles = container.querySelectorAll('[data-slot="card-title"]');
    expect(titles.length).toBeGreaterThan(0);
    for (const title of titles) {
      expect(title).toHaveClass("font-heading");
      // CardTitle supplies font-heading by default; the size must scale with the container.
      expect(title).toHaveClass("text-lg", "@xl:text-2xl");
      expect(title).not.toHaveClass("text-sm");
    }
  });
});

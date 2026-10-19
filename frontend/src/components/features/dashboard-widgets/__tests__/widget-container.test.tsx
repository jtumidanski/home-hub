import { describe, it, expect, vi } from "vitest";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { widgetRegistry } from "@/lib/dashboard/widget-registry";

const loaded = { data: { data: [] }, isLoading: false, isError: false };

vi.mock("@/context/tenant-context", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useTenant: () => ({ household: { attributes: { timezone: "UTC", units: "imperial" } } }),
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

  it.each(widgetRegistry.map((w) => [w.type]))("%s title uses font-heading", (type) => {
    const { container } = renderWidget(type as string);
    for (const title of container.querySelectorAll('[data-slot="card-title"]')) {
      expect(title).toHaveClass("font-heading");
    }
  });
});

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

const mockUseTaskSummary = vi.fn();
const mockUseReminderSummary = vi.fn();
const mockUseAuth = vi.fn();

vi.mock("@/lib/hooks/api/use-tasks", () => ({
  useTaskSummary: () => mockUseTaskSummary(),
}));

vi.mock("@/lib/hooks/api/use-reminders", () => ({
  useReminderSummary: () => mockUseReminderSummary(),
}));

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => mockUseAuth(),
}));

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ tenant: { id: "t1" }, household: { id: "h1", attributes: { timezone: "America/New_York" } } }),
}));

vi.mock("@/components/features/weather/weather-widget", () => ({
  WeatherWidget: () => <div data-testid="weather-widget">Weather Widget</div>,
}));

vi.mock("@/components/features/packages/package-summary-widget", () => ({
  PackageSummaryWidget: () => null,
}));

vi.mock("@/components/features/meals/meal-plan-widget", () => ({
  MealPlanWidget: () => null,
}));

vi.mock("@/components/features/calendar/calendar-widget", () => ({
  CalendarWidget: () => null,
}));

vi.mock("@/components/features/trackers/habits-widget", () => ({
  HabitsWidget: () => null,
}));

vi.mock("@/components/features/workouts/workout-widget", () => ({
  WorkoutWidget: () => null,
}));

import { DashboardPage } from "../DashboardPage";

function renderWithRouter(ui: React.ReactElement) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>
  );
}

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuth.mockReturnValue({
      appContext: { attributes: { resolvedRole: "owner" }, relationships: {} },
    });
    mockUseTaskSummary.mockReturnValue({ data: null, isLoading: false, isError: false });
    mockUseReminderSummary.mockReturnValue({ data: null, isLoading: false, isError: false });
  });

  it("renders loading skeleton when either summary is loading", () => {
    mockUseTaskSummary.mockReturnValue({ data: null, isLoading: true, isError: false });
    mockUseReminderSummary.mockReturnValue({ data: null, isLoading: false, isError: false });
    renderWithRouter(<DashboardPage />);
    expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
  });

  it("renders error card when task summary fails", () => {
    mockUseTaskSummary.mockReturnValue({ data: null, isLoading: false, isError: true });
    renderWithRouter(<DashboardPage />);
    expect(screen.getByText(/failed to load some dashboard data/i)).toBeInTheDocument();
  });

  it("renders dashboard with summary data", () => {
    mockUseTaskSummary.mockReturnValue({
      data: { data: { attributes: { pendingCount: 5, completedTodayCount: 2, overdueCount: 1 } } },
      isLoading: false,
      isError: false,
    });
    mockUseReminderSummary.mockReturnValue({
      data: { data: { attributes: { dueNowCount: 3, upcomingCount: 7, snoozedCount: 1 } } },
      isLoading: false,
      isError: false,
    });
    renderWithRouter(<DashboardPage />);

    expect(screen.getByText("Dashboard")).toBeInTheDocument();
    expect(screen.getByText("5")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
    expect(screen.getByText("2 completed today")).toBeInTheDocument();
    expect(screen.getByText("7 upcoming")).toBeInTheDocument();
  });

  it("renders dash placeholders when summary data is null", () => {
    renderWithRouter(<DashboardPage />);
    expect(screen.getByText("Dashboard")).toBeInTheDocument();
    const dashes = screen.getAllByText("-");
    expect(dashes.length).toBe(3);
  });
});

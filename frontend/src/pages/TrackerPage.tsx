import { useLocation, useNavigate } from "react-router-dom";
import { Settings, CalendarDays, ListChecks } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TrackerSetup } from "@/components/features/tracker/tracker-setup";
import { TodayView } from "@/components/features/tracker/today-view";
import { CalendarGrid } from "@/components/features/tracker/calendar-grid";
import { MonthReport } from "@/components/features/tracker/month-report";
import { PageHeader } from "@/components/common/page-header";
import { getLocalMonth } from "@/lib/date-utils";
import { useQueryParamState, type QueryParamCodec } from "@/lib/hooks/use-query-param-state";

export type HabitsView = "today" | "calendar" | "setup" | "report";

const MONTH_RE = /^\d{4}-(0[1-9]|1[0-2])$/;

const monthCodec: QueryParamCodec<string> = {
  key: "month",
  parse: (raw) => (MONTH_RE.test(raw) ? raw : null),
  serialize: (value) => value,
  current: getLocalMonth,
  equals: (a, b) => a === b,
};

const VIEWS = [
  { id: "today", label: "Today", icon: ListChecks },
  { id: "calendar", label: "Calendar", icon: CalendarDays },
  { id: "setup", label: "Setup", icon: Settings },
] as const;

export function TrackerPage({ view }: { view: HabitsView }) {
  const [month, setMonth] = useQueryParamState(monthCodec);
  const navigate = useNavigate();
  const location = useLocation();
  const goTo = (target: HabitsView) =>
    navigate({ pathname: `/app/habits/${target}`, search: location.search });

  const widthClass =
    view === "calendar" ? "max-w-none" :
    view === "report" ? "max-w-6xl" :
    "max-w-3xl";

  return (
    <div className="p-4 md:p-6 space-y-4">
      <PageHeader
        title="Habits"
        actions={
          <div role="group" aria-label="Habits view" className="flex gap-1">
            {VIEWS.map(({ id, label, icon: Icon }) => (
              <Button
                key={id}
                variant={view === id ? "default" : "ghost"}
                size="sm"
                aria-pressed={view === id}
                onClick={() => goTo(id)}
              >
                <Icon className="h-4 w-4 mr-1" aria-hidden="true" /> {label}
              </Button>
            ))}
          </div>
        }
      />

      <div className={`mx-auto ${widthClass}`}>
        {view === "today" && (
          <TodayView />
        )}
        {view === "calendar" && (
          <CalendarGrid month={month} onMonthChange={setMonth} onViewReport={() => goTo("report")} onOpenSetup={() => goTo("setup")} />
        )}
        {view === "setup" && <TrackerSetup />}
        {view === "report" && (
          <MonthReport month={month} />
        )}
      </div>
    </div>
  );
}

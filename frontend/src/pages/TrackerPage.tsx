import { useState } from "react";
import { Settings, CalendarDays, ListChecks } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TrackerSetup } from "@/components/features/tracker/tracker-setup";
import { TodayView } from "@/components/features/tracker/today-view";
import { CalendarGrid } from "@/components/features/tracker/calendar-grid";
import { MonthReport } from "@/components/features/tracker/month-report";
import { PageHeader } from "@/components/common/page-header";
import { getLocalMonth } from "@/lib/date-utils";

type View = "today" | "calendar" | "setup" | "report";

const VIEWS = [
  { id: "today", label: "Today", icon: ListChecks },
  { id: "calendar", label: "Calendar", icon: CalendarDays },
  { id: "setup", label: "Setup", icon: Settings },
] as const;

export function TrackerPage() {
  const [view, setView] = useState<View>("today");
  const [month, setMonth] = useState(getLocalMonth);

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
                onClick={() => setView(id)}
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
          <CalendarGrid month={month} onMonthChange={setMonth} onViewReport={() => setView("report")} onOpenSetup={() => setView("setup")} />
        )}
        {view === "setup" && <TrackerSetup />}
        {view === "report" && (
          <MonthReport month={month} />
        )}
      </div>
    </div>
  );
}

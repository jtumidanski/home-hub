import { useState } from "react";
import { Settings, CalendarDays, ListChecks } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TrackerSetup } from "@/components/features/tracker/tracker-setup";
import { TodayView } from "@/components/features/tracker/today-view";
import { CalendarGrid } from "@/components/features/tracker/calendar-grid";
import { MonthReport } from "@/components/features/tracker/month-report";

type View = "today" | "calendar" | "setup" | "report";

function getCurrentMonth() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
}

export function TrackerPage() {
  const [view, setView] = useState<View>("today");
  const [month, setMonth] = useState(getCurrentMonth);

  const widthClass =
    view === "calendar" ? "max-w-none" :
    view === "report" ? "max-w-6xl" :
    "max-w-3xl";

  return (
    <div className="py-4 px-4 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Habits</h1>
        <div className="flex gap-1">
          <Button variant={view === "today" ? "default" : "ghost"} size="sm" onClick={() => setView("today")}>
            <ListChecks className="h-4 w-4 mr-1" /> Today
          </Button>
          <Button variant={view === "calendar" ? "default" : "ghost"} size="sm" onClick={() => setView("calendar")}>
            <CalendarDays className="h-4 w-4 mr-1" /> Calendar
          </Button>
          <Button variant={view === "setup" ? "default" : "ghost"} size="sm" onClick={() => setView("setup")}>
            <Settings className="h-4 w-4 mr-1" /> Setup
          </Button>
        </div>
      </div>

      <div className={`mx-auto ${widthClass}`}>
        {view === "today" && (
          <TodayView />
        )}
        {view === "calendar" && (
          <CalendarGrid month={month} onMonthChange={setMonth} onViewReport={() => setView("report")} />
        )}
        {view === "setup" && <TrackerSetup />}
        {view === "report" && (
          <MonthReport month={month} onBackToCalendar={() => setView("calendar")} />
        )}
      </div>
    </div>
  );
}

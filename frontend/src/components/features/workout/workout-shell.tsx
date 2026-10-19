import { NavLink, Outlet } from "react-router-dom";
import { CalendarDays, Dumbbell, ListChecks, Tag, BarChart3 } from "lucide-react";
import { cn } from "@/lib/utils";
import { PageHeader } from "@/components/common/page-header";

// Shared chrome for the Workout section. Today / Week / Exercises / Taxonomy /
// Review tabs map to nested routes; the rendered page sits inside <Outlet />.
const tabs = [
  { to: "/app/workouts/today", icon: ListChecks, label: "Today" },
  { to: "/app/workouts/week", icon: CalendarDays, label: "Week" },
  { to: "/app/workouts/exercises", icon: Dumbbell, label: "Exercises" },
  { to: "/app/workouts/taxonomy", icon: Tag, label: "Taxonomy" },
  { to: "/app/workouts/review", icon: BarChart3, label: "Review" },
];

export function WorkoutShell() {
  return (
    <div className="p-4 md:p-6 space-y-4">
      <PageHeader title="Workouts" />
      <nav aria-label="Workout sections" className="-mx-1 overflow-x-auto px-1">
        <ul className="flex min-w-max gap-1 border-b">
          {tabs.map((t) => (
            <li key={t.to}>
              <NavLink
                to={t.to}
                className={({ isActive }) =>
                  cn(
                    "-mb-px flex min-h-9 items-center gap-1.5 border-b-2 px-3 text-sm transition-colors coarse:min-h-11 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
                    isActive
                      ? "border-primary font-semibold text-foreground"
                      : "border-transparent font-medium text-muted-foreground hover:bg-muted hover:text-foreground",
                  )
                }
              >
                <t.icon className="h-4 w-4" aria-hidden="true" />
                {t.label}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
      <Outlet />
    </div>
  );
}

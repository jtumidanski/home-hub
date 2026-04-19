import { NavLink, Outlet } from "react-router-dom";
import { CalendarDays, Dumbbell, ListChecks, Tag, BarChart3 } from "lucide-react";
import { cn } from "@/lib/utils";

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
    <div className="py-4 px-4 space-y-4">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h1 className="text-2xl font-bold">Workouts</h1>
        <nav className="flex flex-wrap gap-1">
          {tabs.map((t) => (
            <NavLink
              key={t.to}
              to={t.to}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                  isActive ? "bg-primary text-primary-foreground" : "hover:bg-muted",
                )
              }
            >
              <t.icon className="h-4 w-4" />
              {t.label}
            </NavLink>
          ))}
        </nav>
      </div>
      <Outlet />
    </div>
  );
}

import { Home, CheckSquare, Bell, Calendar, Package, CloudSun, UtensilsCrossed, Carrot, CalendarDays, ShoppingCart, Settings, type LucideIcon } from "lucide-react";

export interface NavItem {
  to: string;
  icon: LucideIcon;
  label: string;
  end?: boolean;
  badgeKey?: string;
}

export interface NavGroup {
  key: string;
  label: string;
  items: NavItem[];
}

export const navGroups: NavGroup[] = [
  {
    key: "home",
    label: "Home",
    items: [
      { to: "/app", icon: Home, label: "Dashboard", end: true },
    ],
  },
  {
    key: "productivity",
    label: "Productivity",
    items: [
      { to: "/app/tasks", icon: CheckSquare, label: "Tasks" },
      { to: "/app/reminders", icon: Bell, label: "Reminders" },
      { to: "/app/calendar", icon: Calendar, label: "Calendar" },
      { to: "/app/packages", icon: Package, label: "Packages", badgeKey: "inTransitCount" },
    ],
  },
  {
    key: "lifestyle",
    label: "Lifestyle",
    items: [
      { to: "/app/recipes", icon: UtensilsCrossed, label: "Recipes" },
      { to: "/app/meals", icon: CalendarDays, label: "Meal Planner" },
      { to: "/app/shopping", icon: ShoppingCart, label: "Shopping" },
      { to: "/app/ingredients", icon: Carrot, label: "Ingredients" },
      { to: "/app/weather", icon: CloudSun, label: "Weather" },
    ],
  },
  {
    key: "management",
    label: "Management",
    items: [
      { to: "/app/households", icon: Home, label: "Households", badgeKey: "pendingInvitationCount" },
    ],
  },
];

export const settingsNavItem: NavItem = {
  to: "/app/settings",
  icon: Settings,
  label: "Settings",
};

'use client';

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Home,
  Building2,
  Users,
  Monitor,
  Calendar,
  Cloud,
  CheckSquare,
  UtensilsCrossed,
  Bell,
  Network,
  Cog,
  FileText,
  type LucideIcon,
} from "lucide-react";

export interface SidebarProps {
  mobileMenuOpen: boolean;
  onClose: () => void;
}

interface MenuItem {
  label: string;
  href: string;
  icon: LucideIcon;
}

interface MenuSection {
  title: string;
  items: MenuItem[];
}

const menuConfig: MenuSection[] = [
  {
    title: "Management",
    items: [
      { label: "Tenants", href: "/tenants", icon: Building2 },
      { label: "Households", href: "/households", icon: Home },
      { label: "Users", href: "/users", icon: Users },
      { label: "Devices", href: "/devices", icon: Monitor },
    ],
  },
  {
    title: "Services",
    items: [
      { label: "Calendar", href: "/calendar", icon: Calendar },
      { label: "Weather", href: "/weather", icon: Cloud },
      { label: "Tasks", href: "/tasks", icon: CheckSquare },
      { label: "Meals", href: "/meals", icon: UtensilsCrossed },
      { label: "Reminders", href: "/reminders", icon: Bell },
    ],
  },
  {
    title: "System",
    items: [
      { label: "Gateway", href: "/system/gateway", icon: Network },
      { label: "Workers", href: "/system/workers", icon: Cog },
      { label: "Logs", href: "/system/logs", icon: FileText },
    ],
  },
];

export function Sidebar({ mobileMenuOpen, onClose }: SidebarProps) {
  const pathname = usePathname();

  const isActive = (href: string) => {
    if (href === "/") {
      return pathname === "/";
    }
    return pathname.startsWith(href);
  };

  const sidebarContent = (
    <div className="flex h-full flex-col gap-2 p-4">
      {/* Navigation */}
      <nav className="flex flex-1 flex-col gap-1">
        {/* Dashboard link */}
        <Button
          variant={isActive("/") ? "secondary" : "ghost"}
          className="justify-start"
          asChild
          onClick={onClose}
        >
          <Link href="/">
            <Home className="mr-2 h-4 w-4" />
            Dashboard
          </Link>
        </Button>

        <Separator className="my-2" />

        {/* Menu sections */}
        {menuConfig.map((section) => (
          <div key={section.title} className="space-y-1">
            <h3 className="px-3 py-2 text-xs font-semibold text-neutral-500 dark:text-neutral-400">
              {section.title}
            </h3>
            {section.items.map((item) => {
              const Icon = item.icon;
              return (
                <Button
                  key={item.href}
                  variant={isActive(item.href) ? "secondary" : "ghost"}
                  className="justify-start w-full"
                  asChild
                  onClick={onClose}
                >
                  <Link href={item.href}>
                    <Icon className="mr-2 h-4 w-4" />
                    {item.label}
                  </Link>
                </Button>
              );
            })}
            <Separator className="my-2" />
          </div>
        ))}
      </nav>

      {/* Footer */}
      <div className="border-t border-neutral-200 pt-4 dark:border-neutral-800">
        <p className="text-xs text-neutral-500 dark:text-neutral-400">
          Home Hub v1.0
        </p>
      </div>
    </div>
  );

  return (
    <>
      {/* Desktop sidebar - always visible on lg+ */}
      <aside className="hidden w-64 flex-col border-r border-neutral-200 bg-neutral-50 dark:border-neutral-800 dark:bg-neutral-950 lg:flex">
        {sidebarContent}
      </aside>

      {/* Mobile sidebar - drawer overlay */}
      {mobileMenuOpen && (
        <div className="fixed inset-0 z-40 lg:hidden">
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black/50 backdrop-blur-sm"
            onClick={onClose}
            aria-hidden="true"
          />

          {/* Sidebar drawer */}
          <aside className="fixed inset-y-0 left-0 z-50 w-64 flex flex-col border-r border-neutral-200 bg-neutral-50 dark:border-neutral-800 dark:bg-neutral-950 shadow-lg">
            {sidebarContent}
          </aside>
        </div>
      )}
    </>
  );
}

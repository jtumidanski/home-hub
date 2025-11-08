import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";

export function Sidebar() {
  return (
    <aside className="hidden w-64 flex-col border-r border-neutral-200 bg-neutral-50 dark:border-neutral-800 dark:bg-neutral-950 lg:flex">
      <div className="flex h-full flex-col gap-2 p-4">
        {/* Navigation */}
        <nav className="flex flex-1 flex-col gap-1">
          <Button
            variant="ghost"
            className="justify-start"
            asChild
          >
            <Link href="/">Dashboard</Link>
          </Button>

          <Separator className="my-2" />

          <div className="space-y-1">
            <h3 className="px-3 py-2 text-xs font-semibold text-neutral-500 dark:text-neutral-400">
              Management
            </h3>
            <Button
              variant="ghost"
              className="justify-start"
              asChild
            >
              <Link href="/households">Households</Link>
            </Button>
            <Button
              variant="ghost"
              className="justify-start"
              asChild
            >
              <Link href="/users">Users</Link>
            </Button>
            <Button
              variant="ghost"
              className="justify-start"
              asChild
            >
              <Link href="/devices">Devices</Link>
            </Button>
          </div>

          <Separator className="my-2" />

          <div className="space-y-1">
            <h3 className="px-3 py-2 text-xs font-semibold text-neutral-500 dark:text-neutral-400">
              Services
            </h3>
            <Button
              variant="ghost"
              className="justify-start"
              asChild
            >
              <Link href="/tasks">Tasks</Link>
            </Button>
            <Button
              variant="ghost"
              className="justify-start"
              asChild
            >
              <Link href="/meals">Meals</Link>
            </Button>
            <Button
              variant="ghost"
              className="justify-start"
              asChild
            >
              <Link href="/calendar">Calendar</Link>
            </Button>
            <Button
              variant="ghost"
              className="justify-start"
              asChild
            >
              <Link href="/reminders">Reminders</Link>
            </Button>
          </div>
        </nav>

        {/* Footer */}
        <div className="border-t border-neutral-200 pt-4 dark:border-neutral-800">
          <p className="text-xs text-neutral-500 dark:text-neutral-400">
            Next.js 16 • React 19
          </p>
        </div>
      </div>
    </aside>
  );
}

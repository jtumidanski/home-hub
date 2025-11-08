import { Separator } from "@/components/ui/separator";

export function Footer() {
  return (
    <footer className="border-t border-neutral-200 bg-neutral-50 dark:border-neutral-800 dark:bg-neutral-950">
      <div className="container py-6 px-4">
        <div className="flex flex-col items-center justify-between gap-4 md:flex-row">
          <div className="flex flex-col items-center gap-2 md:flex-row md:gap-4">
            <span className="text-sm font-semibold">Home Hub</span>
            <Separator
              orientation="vertical"
              className="hidden h-4 md:block"
            />
            <p className="text-sm text-neutral-600 dark:text-neutral-400">
              Multi-tenant household management platform
            </p>
          </div>
          <div className="flex gap-4">
            <a
              href="/docs"
              className="text-sm text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100"
            >
              Documentation
            </a>
            <a
              href="/support"
              className="text-sm text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100"
            >
              Support
            </a>
          </div>
        </div>
        <Separator className="my-4" />
        <div className="flex flex-col items-center justify-between gap-2 text-xs text-neutral-500 dark:text-neutral-400 md:flex-row">
          <p>© 2025 Home Hub. All rights reserved.</p>
          <p>Built with Next.js, React, and shadcn/ui</p>
        </div>
      </div>
    </footer>
  );
}

'use client';

import { UserProfile } from "@/components/auth/UserProfile";

export interface MobileHeaderProps {
  // Future: could add props for title customization, etc.
}

export function MobileHeader({}: MobileHeaderProps) {
  return (
    <header className="sticky top-0 z-[60] w-full border-b border-neutral-200 bg-white/95 backdrop-blur supports-[backdrop-filter]:bg-white/60 dark:border-neutral-800 dark:bg-neutral-950/95">
      <div className="flex h-16 items-center justify-between px-4">
        {/* Logo and Title */}
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-neutral-900 dark:bg-neutral-100">
            <span className="text-sm font-bold text-white dark:text-neutral-900">
              HH
            </span>
          </div>
          <div className="flex flex-col">
            <span className="text-sm font-semibold leading-none">
              Home Hub
            </span>
            <span className="text-xs text-neutral-500 dark:text-neutral-400">
              Mobile
            </span>
          </div>
        </div>

        {/* User Profile */}
        <UserProfile />
      </div>
    </header>
  );
}

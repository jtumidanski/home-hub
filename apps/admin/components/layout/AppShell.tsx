'use client';

import React from 'react';
import { Header } from './Header';
import { Sidebar } from './Sidebar';

export interface AppShellProps {
  children: React.ReactNode;
}

/**
 * AppShell is the main layout wrapper for the admin portal.
 * It provides a consistent header, sidebar, and content area structure.
 *
 * Layout structure:
 * - Header: Sticky at top with logo and user profile
 * - Sidebar: Fixed on left (desktop), drawer on mobile
 * - Content: Main scrollable area for page content
 *
 * Usage:
 * ```tsx
 * <AppShell>
 *   <YourPageContent />
 * </AppShell>
 * ```
 */
export function AppShell({ children }: AppShellProps) {
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);

  return (
    <div className="relative min-h-screen flex flex-col">
      {/* Header - sticky at top */}
      <Header
        onMenuToggle={() => setMobileMenuOpen(!mobileMenuOpen)}
        mobileMenuOpen={mobileMenuOpen}
      />

      {/* Main layout container */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar - fixed on desktop, drawer on mobile */}
        <Sidebar
          mobileMenuOpen={mobileMenuOpen}
          onClose={() => setMobileMenuOpen(false)}
        />

        {/* Main content area */}
        <main className="flex-1 overflow-y-auto bg-neutral-50 dark:bg-neutral-900">
          <div className="container mx-auto p-6 lg:p-8">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}

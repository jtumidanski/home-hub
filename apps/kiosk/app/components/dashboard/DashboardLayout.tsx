'use client';

import React from 'react';

interface DashboardLayoutProps {
  children: React.ReactNode;
  className?: string;
}

export function DashboardLayout({ children, className = '' }: DashboardLayoutProps) {
  return (
    <div className={`min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800 ${className}`}>
      <div className="h-screen overflow-hidden p-4 md:p-6">
        {/* 4-column responsive grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4 h-full auto-rows-fr">
          {children}
        </div>
      </div>
    </div>
  );
}

interface DashboardColumnProps {
  children: React.ReactNode;
  className?: string;
}

export function DashboardColumn({ children, className = '' }: DashboardColumnProps) {
  return (
    <div className={`flex flex-col gap-4 overflow-y-auto ${className}`}>
      {children}
    </div>
  );
}

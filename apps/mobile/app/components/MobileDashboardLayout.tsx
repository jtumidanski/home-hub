import { ReactNode } from 'react';

interface MobileDashboardLayoutProps {
  children: ReactNode;
}

export function MobileDashboardLayout({ children }: MobileDashboardLayoutProps) {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 p-4 md:p-6 max-w-7xl mx-auto">
        {children}
      </div>
    </div>
  );
}

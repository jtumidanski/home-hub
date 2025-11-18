import { ReactNode } from 'react';

interface MobileDashboardLayoutProps {
  children: ReactNode;
  header?: ReactNode;
}

export function MobileDashboardLayout({ children, header }: MobileDashboardLayoutProps) {
  return (
    <div className="min-h-screen bg-background">
      {header}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 p-4 md:p-6 max-w-7xl mx-auto">
        {children}
      </div>
    </div>
  );
}

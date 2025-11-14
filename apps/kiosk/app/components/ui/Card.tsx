import React from 'react';

interface CardProps {
  children?: React.ReactNode;
  title?: string;
  className?: string;
  loading?: boolean;
}

export function Card({ children, title, className = '', loading = false }: CardProps) {
  return (
    <div className={`bg-card text-card-foreground rounded-lg shadow-lg border border-border overflow-hidden ${className}`}>
      {title && (
        <div className="px-6 py-4 border-b border-border">
          <h2 className="text-lg font-semibold">
            {title}
          </h2>
        </div>
      )}
      <div className="p-6">
        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        ) : (
          children
        )}
      </div>
    </div>
  );
}

interface CardSectionProps {
  title?: string;
  children: React.ReactNode;
  className?: string;
}

export function CardSection({ title, children, className = '' }: CardSectionProps) {
  return (
    <div className={`space-y-3 ${className}`}>
      {title && (
        <h3 className="text-sm font-medium text-muted-foreground">
          {title}
        </h3>
      )}
      {children}
    </div>
  );
}

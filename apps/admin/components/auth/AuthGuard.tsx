'use client';

import React from 'react';
import { useAuth } from '@/lib/auth/AuthContext';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { AlertCircle } from 'lucide-react';

export interface AuthGuardProps {
  children: React.ReactNode;
  /** If provided, requires the user to have this specific role */
  requireRole?: string;
  /** If provided, requires the user to have any of these roles */
  requireAnyRole?: string[];
  /** Custom fallback to show when not authenticated */
  fallback?: React.ReactNode;
  /** Custom fallback to show when missing required role */
  unauthorizedFallback?: React.ReactNode;
}

/**
 * AuthGuard protects child components by only rendering them if the user is authenticated
 * and optionally has the required role(s).
 *
 * Usage:
 * ```tsx
 * <AuthGuard>
 *   <ProtectedContent />
 * </AuthGuard>
 *
 * // With role requirement
 * <AuthGuard requireRole="admin">
 *   <AdminContent />
 * </AuthGuard>
 * ```
 */
export function AuthGuard({
  children,
  requireRole,
  requireAnyRole,
  fallback,
  unauthorizedFallback,
}: AuthGuardProps) {
  const { user, loading, error, hasRole, hasAnyRole } = useAuth();

  // Show loading state
  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="flex flex-col items-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
          <p className="text-sm text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  // Show error state
  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Alert variant="destructive" className="max-w-md">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Authentication Error</AlertTitle>
          <AlertDescription>
            {error.message || 'Failed to load authentication status'}
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  // Show fallback if not authenticated
  if (!user) {
    if (fallback) {
      return <>{fallback}</>;
    }

    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Alert className="max-w-md">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Authentication Required</AlertTitle>
          <AlertDescription>
            You must be logged in to access this page.
            Redirecting to login...
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  // Check role requirements
  if (requireRole && !hasRole(requireRole)) {
    if (unauthorizedFallback) {
      return <>{unauthorizedFallback}</>;
    }

    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Alert variant="destructive" className="max-w-md">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Unauthorized</AlertTitle>
          <AlertDescription>
            You do not have the required role ({requireRole}) to access this
            page.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  if (requireAnyRole && !hasAnyRole(requireAnyRole)) {
    if (unauthorizedFallback) {
      return <>{unauthorizedFallback}</>;
    }

    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Alert variant="destructive" className="max-w-md">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Unauthorized</AlertTitle>
          <AlertDescription>
            You do not have any of the required roles to access this page.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  // User is authenticated and authorized
  return <>{children}</>;
}

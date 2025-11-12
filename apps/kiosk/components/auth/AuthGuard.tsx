'use client';

import React from 'react';
import { useAuth } from '@/lib/auth';

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
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
        <div className="flex flex-col items-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          <p className="text-sm text-gray-600 dark:text-gray-400">Loading...</p>
        </div>
      </div>
    );
  }

  // Show error state
  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4 bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-6 max-w-md">
          <h3 className="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">
            Authentication Error
          </h3>
          <p className="text-sm text-red-700 dark:text-red-300">
            {error.message || 'Failed to load authentication status'}
          </p>
        </div>
      </div>
    );
  }

  // Show fallback if not authenticated
  if (!user) {
    if (fallback) {
      return <>{fallback}</>;
    }

    return (
      <div className="flex items-center justify-center min-h-screen p-4 bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-6 max-w-md">
          <h3 className="text-lg font-semibold text-blue-900 dark:text-blue-100 mb-2">
            Authentication Required
          </h3>
          <p className="text-sm text-blue-700 dark:text-blue-300">
            You must be logged in to access this page.
          </p>
        </div>
      </div>
    );
  }

  // Check role requirements
  if (requireRole && !hasRole(requireRole)) {
    if (unauthorizedFallback) {
      return <>{unauthorizedFallback}</>;
    }

    return (
      <div className="flex items-center justify-center min-h-screen p-4 bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-6 max-w-md">
          <h3 className="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">
            Unauthorized
          </h3>
          <p className="text-sm text-red-700 dark:text-red-300">
            You do not have the required role ({requireRole}) to access this page.
          </p>
        </div>
      </div>
    );
  }

  if (requireAnyRole && !hasAnyRole(requireAnyRole)) {
    if (unauthorizedFallback) {
      return <>{unauthorizedFallback}</>;
    }

    return (
      <div className="flex items-center justify-center min-h-screen p-4 bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-6 max-w-md">
          <h3 className="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">
            Unauthorized
          </h3>
          <p className="text-sm text-red-700 dark:text-red-300">
            You do not have any of the required roles to access this page.
          </p>
        </div>
      </div>
    );
  }

  // User is authenticated and authorized
  return <>{children}</>;
}

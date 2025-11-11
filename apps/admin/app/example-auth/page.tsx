'use client';

import { useAuth } from '@/lib/auth';
import { UserProfile, UserBadge, AuthGuard } from '@/components/auth';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Shield, User, Calendar, RefreshCw } from 'lucide-react';

/**
 * Example page demonstrating auth system usage
 * Visit /example-auth to see this page
 */
export default function ExampleAuthPage() {
  return (
    <AuthGuard>
      <ExampleAuthContent />
    </AuthGuard>
  );
}

function ExampleAuthContent() {
  const { user, roles, loading, hasRole, isAdmin, refetch } = useAuth();

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="text-muted-foreground">Loading authentication...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return null; // AuthGuard handles this
  }

  return (
    <div className="container mx-auto p-8 space-y-8">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Authentication Example</h1>
          <p className="text-muted-foreground">
            Demonstrating the Home Hub auth system
          </p>
        </div>
        <UserProfile />
      </div>

      {/* User Info Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <User className="h-5 w-5" />
            Current User
          </CardTitle>
          <CardDescription>
            Information about the currently authenticated user
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">
                Display Name
              </p>
              <p className="text-lg">{user.displayName}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Email</p>
              <p className="text-lg">{user.email}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">
                Provider
              </p>
              <Badge variant="outline">{user.provider}</Badge>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">
                User ID
              </p>
              <p className="text-xs font-mono">{user.id}</p>
            </div>
          </div>

          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">
              Timestamps
            </p>
            <div className="flex gap-4 text-sm">
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <span>Created: {new Date(user.createdAt).toLocaleDateString()}</span>
              </div>
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <span>Updated: {new Date(user.updatedAt).toLocaleDateString()}</span>
              </div>
            </div>
          </div>

          <Button onClick={refetch} variant="outline" size="sm">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh User Data
          </Button>
        </CardContent>
      </Card>

      {/* Roles Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Roles & Permissions
          </CardTitle>
          <CardDescription>
            User roles determine access to features
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">
              Assigned Roles
            </p>
            <div className="flex flex-wrap gap-2">
              {roles.map((role) => (
                <Badge key={role} variant="secondary">
                  {role}
                </Badge>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <p className="text-sm font-medium text-muted-foreground">
              Role Checks
            </p>
            <div className="space-y-1 text-sm">
              <div className="flex items-center gap-2">
                <Badge variant={isAdmin() ? 'default' : 'outline'}>
                  {isAdmin() ? '✓' : '✗'}
                </Badge>
                <span>Is Admin</span>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant={hasRole('user') ? 'default' : 'outline'}>
                  {hasRole('user') ? '✓' : '✗'}
                </Badge>
                <span>Has User Role</span>
              </div>
              <div className="flex items-center gap-2">
                <Badge
                  variant={hasRole('household_admin') ? 'default' : 'outline'}
                >
                  {hasRole('household_admin') ? '✓' : '✗'}
                </Badge>
                <span>Has Household Admin Role</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Component Examples */}
      <Card>
        <CardHeader>
          <CardTitle>Component Examples</CardTitle>
          <CardDescription>
            Pre-built components for common auth use cases
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">
              UserBadge Component
            </p>
            <UserBadge />
          </div>

          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">
              UserProfile Component (Check header →)
            </p>
            <p className="text-sm text-muted-foreground">
              The UserProfile dropdown is shown in the page header above.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Code Examples */}
      <Card>
        <CardHeader>
          <CardTitle>Usage Examples</CardTitle>
          <CardDescription>
            How to use auth in your components
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <p className="text-sm font-medium mb-2">Basic Hook Usage</p>
              <pre className="bg-muted p-4 rounded-lg text-xs overflow-x-auto">
                {`import { useAuth } from '@/lib/auth';

function MyComponent() {
  const { user, roles, loading } = useAuth();

  if (loading) return <div>Loading...</div>;
  if (!user) return <div>Not authenticated</div>;

  return <div>Hello, {user.displayName}!</div>;
}`}
              </pre>
            </div>

            <div>
              <p className="text-sm font-medium mb-2">Using AuthGuard</p>
              <pre className="bg-muted p-4 rounded-lg text-xs overflow-x-auto">
                {`import { AuthGuard } from '@/components/auth';

// Basic protection
<AuthGuard>
  <ProtectedContent />
</AuthGuard>

// Require admin role
<AuthGuard requireRole="admin">
  <AdminOnlyContent />
</AuthGuard>`}
              </pre>
            </div>

            <div>
              <p className="text-sm font-medium mb-2">Role Checking</p>
              <pre className="bg-muted p-4 rounded-lg text-xs overflow-x-auto">
                {`const { hasRole, hasAnyRole, isAdmin } = useAuth();

if (isAdmin()) {
  // Show admin features
}

if (hasAnyRole(['admin', 'household_admin'])) {
  // Show management features
}`}
              </pre>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

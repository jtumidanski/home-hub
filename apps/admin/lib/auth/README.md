# Authentication Library

This directory contains the authentication system for the Home Hub admin application.

## Overview

The authentication system uses oauth2-proxy for OAuth flows with Google and GitHub, and provides a React context-based API for accessing user information and managing authentication state.

## Components

### AuthContext.tsx

Provides React context for authentication state.

**Exports:**
- `AuthProvider` - Context provider component
- `useAuth()` - Hook to access auth state
- `useRequireAuth()` - Hook that throws if not authenticated

**Usage:**
```tsx
import { useAuth } from '@/lib/auth';

function MyComponent() {
  const { user, roles, loading, hasRole, isAdmin } = useAuth();

  if (loading) return <div>Loading...</div>;
  if (!user) return <div>Not logged in</div>;

  return (
    <div>
      <p>Hello, {user.displayName}!</p>
      {isAdmin() && <p>You are an admin!</p>}
    </div>
  );
}
```

### api.ts

Client-side API functions for authentication.

**Exports:**
- `fetchMe()` - Fetch current user info from /api/me
- `logout(provider)` - Log out via oauth2-proxy

**Types:**
- `User` - User model
- `MeResponse` - Response from /api/me

## User Interface Components

Located in `/components/auth/`:

### UserProfile.tsx

Dropdown menu component showing user info and logout option.

```tsx
import { UserProfile } from '@/components/auth';

// In your header/navbar
<UserProfile />
```

### AuthGuard.tsx

Protects routes by requiring authentication and optional roles.

```tsx
import { AuthGuard } from '@/components/auth';

// Basic protection
<AuthGuard>
  <ProtectedContent />
</AuthGuard>

// Require admin role
<AuthGuard requireRole="admin">
  <AdminOnlyContent />
</AuthGuard>

// Require any of multiple roles
<AuthGuard requireAnyRole={['admin', 'household_admin']}>
  <ManagerContent />
</AuthGuard>
```

## Setup

The `AuthProvider` is already configured in the root layout (`app/layout.tsx`). All pages automatically have access to authentication state.

## Flow

1. User visits protected page
2. NGINX checks authentication via oauth2-proxy
3. If not authenticated, redirect to OAuth provider
4. After successful auth, user is returned to app with session cookie
5. Frontend calls `/api/me` on mount
6. Backend validates headers, gets/creates user, loads roles
7. Frontend receives user info and roles
8. Components use `useAuth()` to access auth state

## Available Roles

- `user` - Standard authenticated user (default)
- `admin` - Full system access
- `household_admin` - Manage household settings
- `device_manager` - Manage devices

## Examples

### Display user badge
```tsx
import { UserBadge } from '@/components/auth';

<UserBadge />
```

### Check if user has role
```tsx
const { hasRole, hasAnyRole } = useAuth();

if (hasRole('admin')) {
  // Show admin features
}

if (hasAnyRole(['admin', 'household_admin'])) {
  // Show management features
}
```

### Logout button
```tsx
import { logout } from '@/lib/auth';

<button onClick={() => logout('google')}>
  Log out
</button>
```

### Refetch user data
```tsx
const { refetch } = useAuth();

// After updating user profile
await refetch();
```

## Error Handling

The `useAuth()` hook provides an `error` field:

```tsx
const { error } = useAuth();

if (error) {
  return <div>Error: {error.message}</div>;
}
```

## TypeScript

All components and hooks are fully typed. Import types as needed:

```tsx
import type { User, MeResponse } from '@/lib/auth';
```

## Testing

When testing components that use auth:

```tsx
import { AuthProvider } from '@/lib/auth';

// Wrap test component
<AuthProvider>
  <YourComponent />
</AuthProvider>
```

## See Also

- `/docs/auth.md` - Complete authentication architecture documentation
- `/docs/auth-setup-google.md` - Google OAuth setup guide
- `/docs/auth-setup-github.md` - GitHub OAuth setup guide

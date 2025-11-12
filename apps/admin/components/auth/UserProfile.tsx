'use client';

import React, { useState } from 'react';
import { useAuth } from '@/lib/auth/AuthContext';
import { logout } from '@/lib/auth/api';
import { Button } from '@/components/ui/button';
import { ProfileModal } from './ProfileModal';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { User, LogOut, Shield } from 'lucide-react';

/**
 * UserProfile component displays the authenticated user's information
 * with a dropdown menu for logout and profile management.
 *
 * Usage:
 * ```tsx
 * <UserProfile />
 * ```
 */
export function UserProfile() {
  const { user, roles, loading } = useAuth();
  const [profileOpen, setProfileOpen] = useState(false);

  if (loading) {
    return (
      <div className="flex items-center space-x-2">
        <div className="h-8 w-8 rounded-full bg-gray-200 animate-pulse" />
        <div className="h-4 w-20 bg-gray-200 rounded animate-pulse" />
      </div>
    );
  }

  if (!user) {
    return (
      <Button
        variant="default"
        size="sm"
        onClick={() => {
          // Redirect to the admin page which will trigger OAuth
          window.location.href = '/admin';
        }}
      >
        Sign In
      </Button>
    );
  }

  const initials = user.displayName
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .substring(0, 2);

  const handleLogout = () => {
    logout(user.provider || 'google'); // Default to google if provider not specified
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="relative h-10 w-10 rounded-full">
          <Avatar className="h-10 w-10">
            <AvatarFallback className="bg-primary text-primary-foreground">
              {initials}
            </AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-64" align="end" forceMount>
        <DropdownMenuLabel className="font-normal">
          <div className="flex flex-col space-y-2">
            <p className="text-sm font-medium leading-none">
              {user.displayName}
            </p>
            <p className="text-xs leading-none text-muted-foreground">
              {user.email}
            </p>
            <div className="flex flex-wrap gap-1 mt-2">
              {roles.map((role) => (
                <Badge key={role} variant="secondary" className="text-xs">
                  {role === 'admin' && <Shield className="w-3 h-3 mr-1" />}
                  {role}
                </Badge>
              ))}
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              via {user.provider}
            </div>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => setProfileOpen(true)}>
          <User className="mr-2 h-4 w-4" />
          <span>Profile</span>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleLogout} className="text-red-600">
          <LogOut className="mr-2 h-4 w-4" />
          <span>Log out</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
      <ProfileModal open={profileOpen} onOpenChange={setProfileOpen} />
    </DropdownMenu>
  );
}

/**
 * Simple user badge component for displaying user info inline
 */
export function UserBadge() {
  const { user, loading } = useAuth();

  if (loading || !user) {
    return null;
  }

  return (
    <div className="flex items-center space-x-2 px-3 py-2 bg-secondary rounded-lg">
      <User className="h-4 w-4 text-muted-foreground" />
      <span className="text-sm font-medium">{user.displayName}</span>
    </div>
  );
}

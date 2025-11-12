'use client';

import React from 'react';
import { useAuth } from '@/lib/auth/AuthContext';
import { useTheme } from '@/lib/theme/ThemeContext';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Monitor, Moon, Sun } from 'lucide-react';

export interface ProfileModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * ProfileModal component displays user profile settings
 * including theme preferences
 */
export function ProfileModal({ open, onOpenChange }: ProfileModalProps) {
  const { user, roles } = useAuth();
  const { mode, resolvedTheme, setTheme } = useTheme();

  if (!user) {
    return null;
  }

  const themeOptions = [
    {
      value: 'system' as const,
      label: 'System',
      description: 'Follow system preference',
      icon: Monitor,
    },
    {
      value: 'light' as const,
      label: 'Light',
      description: 'Light theme',
      icon: Sun,
    },
    {
      value: 'dark' as const,
      label: 'Dark',
      description: 'Dark theme',
      icon: Moon,
    },
  ];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Profile Settings</DialogTitle>
          <DialogDescription>
            Manage your account settings and preferences
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* User Information Section */}
          <div className="space-y-3">
            <h3 className="text-sm font-medium">Account Information</h3>
            <div className="space-y-2 text-sm">
              <div>
                <span className="text-muted-foreground">Name:</span>{' '}
                <span className="font-medium">{user.displayName}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Email:</span>{' '}
                <span className="font-medium">{user.email}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Provider:</span>{' '}
                <span className="font-medium capitalize">{user.provider}</span>
              </div>
              {roles.length > 0 && (
                <div>
                  <span className="text-muted-foreground">Roles:</span>{' '}
                  <div className="inline-flex gap-1 ml-1">
                    {roles.map((role) => (
                      <Badge key={role} variant="secondary" className="text-xs">
                        {role}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>

          <Separator />

          {/* Theme Preference Section */}
          <div className="space-y-3">
            <div>
              <h3 className="text-sm font-medium">Appearance</h3>
              <p className="text-xs text-muted-foreground">
                Choose your interface theme{' '}
                {mode === 'system' && (
                  <span>(currently {resolvedTheme})</span>
                )}
              </p>
            </div>

            <div className="grid grid-cols-3 gap-3">
              {themeOptions.map((option) => {
                const Icon = option.icon;
                const isSelected = mode === option.value;

                return (
                  <button
                    key={option.value}
                    onClick={() => setTheme(option.value)}
                    className={`
                      relative flex flex-col items-center gap-2 rounded-lg border-2 p-4
                      transition-all hover:bg-accent
                      ${
                        isSelected
                          ? 'border-primary bg-accent'
                          : 'border-border'
                      }
                    `}
                    aria-label={`Set theme to ${option.label}`}
                    aria-pressed={isSelected}
                  >
                    <Icon
                      className={`h-5 w-5 ${
                        isSelected ? 'text-primary' : 'text-muted-foreground'
                      }`}
                    />
                    <div className="text-center">
                      <div
                        className={`text-sm font-medium ${
                          isSelected ? 'text-primary' : ''
                        }`}
                      >
                        {option.label}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {option.description}
                      </div>
                    </div>
                    {isSelected && (
                      <div className="absolute top-2 right-2 h-2 w-2 rounded-full bg-primary" />
                    )}
                  </button>
                );
              })}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

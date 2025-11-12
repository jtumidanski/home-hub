'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/lib/auth/AuthContext';
import {
  ThemeMode,
  DEFAULT_THEME_MODE,
  THEME_STORAGE_KEY,
  THEME_PREFERENCE_KEY,
  getSystemTheme,
  resolveTheme,
  applyTheme
} from './constants';
import { setThemePreference } from './api';

interface ThemeContextValue {
  /** Current theme mode (system, light, or dark) */
  mode: ThemeMode;
  /** Resolved theme (light or dark) */
  resolvedTheme: 'light' | 'dark';
  /** Set the theme mode */
  setTheme: (mode: ThemeMode) => void;
  /** Whether the theme is loading */
  loading: boolean;
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

export interface ThemeProviderProps {
  children: React.ReactNode;
}

/**
 * ThemeProvider manages the application theme
 * - Loads theme from user preferences (if authenticated)
 * - Syncs to localStorage for instant loading
 * - Syncs to backend API (debounced)
 * - Listens to system theme changes
 */
export function ThemeProvider({ children }: ThemeProviderProps) {
  const { user, loading: authLoading } = useAuth();
  const [mode, setMode] = useState<ThemeMode>(DEFAULT_THEME_MODE);
  const [resolvedTheme, setResolvedTheme] = useState<'light' | 'dark'>('light');
  const [loading, setLoading] = useState(true);

  // Initialize theme from localStorage or user preferences
  useEffect(() => {
    if (authLoading) {
      return;
    }

    // Priority 1: User preferences from backend (if logged in)
    if (user?.preferences?.[THEME_PREFERENCE_KEY]) {
      const userTheme = user.preferences[THEME_PREFERENCE_KEY] as ThemeMode;
      if (userTheme === 'system' || userTheme === 'light' || userTheme === 'dark') {
        setMode(userTheme);
        setLoading(false);
        return;
      }
    }

    // Priority 2: localStorage
    try {
      const stored = localStorage.getItem(THEME_STORAGE_KEY);
      if (stored && (stored === 'system' || stored === 'light' || stored === 'dark')) {
        setMode(stored as ThemeMode);
      }
    } catch (error) {
      console.error('Failed to load theme from localStorage:', error);
    }

    setLoading(false);
  }, [user, authLoading]);

  // Resolve and apply theme whenever mode changes or system theme changes
  useEffect(() => {
    const resolved = resolveTheme(mode);
    setResolvedTheme(resolved);
    applyTheme(resolved);
  }, [mode]);

  // Listen to system theme changes (only if mode is 'system')
  useEffect(() => {
    if (mode !== 'system') {
      return;
    }

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

    const handler = (e: MediaQueryListEvent) => {
      const systemTheme = e.matches ? 'dark' : 'light';
      setResolvedTheme(systemTheme);
      applyTheme(systemTheme);
    };

    // Modern browsers
    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handler);
      return () => mediaQuery.removeEventListener('change', handler);
    }

    // Fallback for older browsers
    if (mediaQuery.addListener) {
      mediaQuery.addListener(handler);
      return () => mediaQuery.removeListener(handler);
    }
  }, [mode]);

  // Sync to backend (debounced)
  useEffect(() => {
    if (!user || authLoading) {
      return;
    }

    // Debounce backend sync to avoid spamming the API
    const timer = setTimeout(async () => {
      try {
        await setThemePreference(mode);
      } catch (error) {
        console.error('Failed to save theme preference to backend:', error);
        // Don't show error to user - theme is still applied locally
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [mode, user, authLoading]);

  // Set theme handler
  const handleSetTheme = useCallback((newMode: ThemeMode) => {
    setMode(newMode);

    // Immediately save to localStorage
    try {
      localStorage.setItem(THEME_STORAGE_KEY, newMode);
    } catch (error) {
      console.error('Failed to save theme to localStorage:', error);
    }
  }, []);

  const value: ThemeContextValue = {
    mode,
    resolvedTheme,
    setTheme: handleSetTheme,
    loading,
  };

  return (
    <ThemeContext.Provider value={value}>
      {children}
    </ThemeContext.Provider>
  );
}

/**
 * Hook to access the theme context
 */
export function useTheme(): ThemeContextValue {
  const context = useContext(ThemeContext);

  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }

  return context;
}

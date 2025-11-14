import { useState, useEffect } from 'react';
import { searchLocations, GeocodingResult } from '@/lib/api/geocoding';

/**
 * Custom hook for debounced geocoding search
 *
 * Provides debounced search functionality to prevent excessive API calls
 * while the user is typing. Automatically handles loading and error states.
 *
 * @param query - Search query string
 * @param debounceMs - Debounce delay in milliseconds (default: 500ms)
 * @returns Object containing results, loading state, and error state
 *
 * @example
 * ```typescript
 * function MyComponent() {
 *   const [query, setQuery] = useState('');
 *   const { results, loading, error } = useGeocodingSearch(query);
 *
 *   return (
 *     <input
 *       value={query}
 *       onChange={(e) => setQuery(e.target.value)}
 *     />
 *   );
 * }
 * ```
 */
export function useGeocodingSearch(query: string, debounceMs: number = 500) {
  const [results, setResults] = useState<GeocodingResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Clear results if query is too short
    if (!query || query.trim().length < 3) {
      setResults([]);
      setLoading(false);
      setError(null);
      return;
    }

    // Set loading state immediately
    setLoading(true);
    setError(null);

    // Debounce the search
    const timeoutId = setTimeout(async () => {
      try {
        const data = await searchLocations(query);
        setResults(data);
        setError(null);
      } catch (err) {
        console.error('Search error:', err);
        setError('Failed to search locations. Please try again.');
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, debounceMs);

    // Cleanup: cancel the timeout if query changes before delay completes
    return () => {
      clearTimeout(timeoutId);
    };
  }, [query, debounceMs]);

  return { results, loading, error };
}

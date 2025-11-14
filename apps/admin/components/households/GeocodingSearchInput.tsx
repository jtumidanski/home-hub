import { useState, useRef, useEffect } from 'react';
import { Search, Loader2, MapPin } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { useGeocodingSearch } from '@/hooks/useGeocodingSearch';
import type { GeocodingResult } from '@/lib/api/geocoding';

interface GeocodingSearchInputProps {
  /** Callback when a location is selected from the results */
  onLocationSelect: (result: GeocodingResult) => void;
  /** Whether the input should be disabled */
  disabled?: boolean;
  /** Placeholder text for the input */
  placeholder?: string;
}

/**
 * Autocomplete search input for city/country geocoding
 *
 * Provides a search interface with debounced autocomplete dropdown
 * for selecting locations. Uses Nominatim (OpenStreetMap) API.
 *
 * @example
 * ```tsx
 * <GeocodingSearchInput
 *   onLocationSelect={(result) => {
 *     setLatitude(result.lat);
 *     setLongitude(result.lon);
 *   }}
 * />
 * ```
 */
export function GeocodingSearchInput({
  onLocationSelect,
  disabled = false,
  placeholder = "Search for a city (e.g., San Francisco, US)",
}: GeocodingSearchInputProps) {
  const [query, setQuery] = useState('');
  const [showResults, setShowResults] = useState(false);
  const { results, loading, error } = useGeocodingSearch(query);
  const inputRef = useRef<HTMLInputElement>(null);
  const resultsRef = useRef<HTMLDivElement>(null);

  // Close results dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      const target = event.target as Node;

      // Check if click is outside both input and results dropdown
      if (
        resultsRef.current &&
        !resultsRef.current.contains(target) &&
        inputRef.current &&
        !inputRef.current.contains(target)
      ) {
        setShowResults(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSelect = (result: GeocodingResult) => {
    onLocationSelect(result);
    setQuery(''); // Clear input after selection
    setShowResults(false);
  };

  return (
    <div className="relative">
      {/* Search Input */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-neutral-400" />
        <Input
          ref={inputRef}
          type="text"
          placeholder={placeholder}
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setShowResults(true);
          }}
          onFocus={() => {
            if (query.length >= 3) {
              setShowResults(true);
            }
          }}
          disabled={disabled}
          className="pl-10 pr-10"
        />
        {loading && (
          <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-neutral-400" />
        )}
      </div>

      {/* Results Dropdown */}
      {showResults && query.length >= 3 && (
        <div
          ref={resultsRef}
          className="absolute z-50 w-full mt-2 bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-md shadow-lg max-h-64 overflow-y-auto"
        >
          {/* Loading State */}
          {loading && (
            <div className="p-4 text-center text-sm text-neutral-500">
              Searching...
            </div>
          )}

          {/* Error State */}
          {error && !loading && (
            <div className="p-4 text-center text-sm text-red-500">{error}</div>
          )}

          {/* No Results State */}
          {!loading && !error && results.length === 0 && (
            <div className="p-4 text-center text-sm text-neutral-500">
              No locations found. Try a different search.
            </div>
          )}

          {/* Results List */}
          {!loading && !error && results.length > 0 && (
            <div className="py-2">
              {results.map((result, index) => (
                <button
                  key={index}
                  onClick={() => handleSelect(result)}
                  className="w-full px-4 py-2 text-left hover:bg-neutral-100 dark:hover:bg-neutral-800 flex items-start gap-2 transition-colors"
                >
                  <MapPin className="h-4 w-4 mt-0.5 text-neutral-400 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium truncate">
                      {result.city || 'Unknown City'}
                    </div>
                    <div className="text-xs text-neutral-500 truncate">
                      {result.displayName}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Helper Text */}
      <p className="text-xs text-neutral-500 mt-2">
        Type at least 3 characters to search for a location
      </p>
    </div>
  );
}

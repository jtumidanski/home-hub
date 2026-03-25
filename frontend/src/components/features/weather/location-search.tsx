import { useState, useRef, useEffect, useCallback } from "react";
import { useGeocodingSearch } from "@/lib/hooks/api/use-weather";
import { Input } from "@/components/ui/input";
import { X, MapPin, Loader2 } from "lucide-react";
import type { GeocodingResult } from "@/types/models/weather";

interface LocationSearchProps {
  value: string | null | undefined;
  onSelect: (place: { name: string; latitude: number; longitude: number }) => void;
  onClear: () => void;
  isPending?: boolean;
}

export function LocationSearch({ value, onSelect, onClear, isPending }: LocationSearchProps) {
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [open, setOpen] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [optimisticValue, setOptimisticValue] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const { data } = useGeocodingSearch(debouncedQuery);
  const results = data?.data ?? [];

  // Clear optimistic value when server value arrives
  const displayValue = value ?? optimisticValue;
  useEffect(() => {
    if (value) {
      setOptimisticValue(null);
    }
  }, [value]);

  // Debounce query
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query), 300);
    return () => clearTimeout(timer);
  }, [query]);

  // Open dropdown when results arrive
  useEffect(() => {
    if (results.length > 0 && query.length >= 2) {
      setOpen(true);
      setSelectedIndex(-1);
    }
  }, [results, query]);

  // Close dropdown on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const handleSelect = useCallback(
    (result: GeocodingResult) => {
      const displayName = [result.attributes.name, result.attributes.admin1, result.attributes.country]
        .filter(Boolean)
        .join(", ");
      setOptimisticValue(displayName);
      onSelect({
        name: displayName,
        latitude: result.attributes.latitude,
        longitude: result.attributes.longitude,
      });
      setQuery("");
      setOpen(false);
    },
    [onSelect],
  );

  const handleClear = useCallback(() => {
    setOptimisticValue(null);
    onClear();
  }, [onClear]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!open) return;
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((i) => Math.min(i + 1, results.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter" && selectedIndex >= 0 && results[selectedIndex]) {
      e.preventDefault();
      handleSelect(results[selectedIndex]);
    } else if (e.key === "Escape") {
      setOpen(false);
    }
  };

  if (displayValue) {
    return (
      <div className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
        <MapPin className="h-4 w-4 text-muted-foreground shrink-0" />
        <span className="flex-1 truncate">{displayValue}</span>
        {isPending ? (
          <Loader2 className="h-4 w-4 animate-spin text-muted-foreground shrink-0" />
        ) : (
          <button
            type="button"
            onClick={handleClear}
            className="shrink-0 text-muted-foreground hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
    );
  }

  return (
    <div ref={containerRef} className="relative">
      <Input
        ref={inputRef}
        placeholder="Search for a city..."
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onKeyDown={handleKeyDown}
        onFocus={() => results.length > 0 && query.length >= 2 && setOpen(true)}
      />
      {open && results.length > 0 && (
        <div className="absolute z-50 mt-1 w-full rounded-md border bg-popover shadow-md max-h-60 overflow-auto">
          {results.map((result, index) => (
            <button
              key={result.id}
              type="button"
              className={`w-full px-3 py-2 text-left text-sm hover:bg-accent ${
                index === selectedIndex ? "bg-accent" : ""
              }`}
              onClick={() => handleSelect(result)}
            >
              <span className="font-medium">{result.attributes.name}</span>
              {result.attributes.admin1 && (
                <span className="text-muted-foreground">, {result.attributes.admin1}</span>
              )}
              {result.attributes.country && (
                <span className="text-muted-foreground">, {result.attributes.country}</span>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

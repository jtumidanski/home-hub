import { useId, useState, useRef, useEffect, useCallback, useMemo } from "react";
import { useGeocodingSearch } from "@/lib/hooks/api/use-weather";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { X, MapPin, Loader2 } from "lucide-react";
import type { GeocodingResult } from "@/types/models/weather";
import { cn } from "@/lib/utils";

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
  const listboxId = useId();
  const optionId = (index: number) => `${listboxId}-option-${index}`;

  const { data } = useGeocodingSearch(debouncedQuery);
  const results = useMemo(() => data?.data ?? [], [data]);

  // Clear optimistic value when server value arrives
  const displayValue = value ?? optimisticValue;
  if (value && optimisticValue !== null) {
    setOptimisticValue(null);
  }

  // Debounce query
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query), 300);
    return () => clearTimeout(timer);
  }, [query]);

  // Open dropdown when results arrive
  const shouldOpen = results.length > 0 && query.length >= 2;
  if (shouldOpen && !open) {
    setOpen(true);
    setSelectedIndex(-1);
  }

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
          <Button
            type="button"
            variant="ghost"
            size="icon-xs"
            onClick={handleClear}
            aria-label="Clear location"
            className="shrink-0 text-muted-foreground hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </Button>
        )}
      </div>
    );
  }

  return (
    <div ref={containerRef} className="relative">
      <Input
        ref={inputRef}
        placeholder="Search for a city..."
        aria-label="Search for a city"
        role="combobox"
        aria-autocomplete="list"
        aria-expanded={open && results.length > 0}
        aria-controls={open && results.length > 0 ? listboxId : undefined}
        aria-activedescendant={open && selectedIndex >= 0 ? optionId(selectedIndex) : undefined}
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onKeyDown={handleKeyDown}
        onFocus={() => results.length > 0 && query.length >= 2 && setOpen(true)}
      />
      <p role="status" aria-live="polite" className="sr-only">
        {open && results.length > 0
          ? `${results.length} result${results.length === 1 ? "" : "s"} available`
          : ""}
      </p>
      {open && results.length > 0 && (
        <div
          id={listboxId}
          role="listbox"
          aria-label="Location results"
          className="absolute z-50 mt-1 w-full rounded-md border bg-popover shadow-overlay max-h-60 overflow-auto"
        >
          {results.map((result, index) => (
            <div
              key={result.id}
              id={optionId(index)}
              role="option"
              aria-selected={index === selectedIndex}
              className={cn(
                "w-full cursor-pointer px-3 py-2 text-left text-sm hover:bg-accent",
                index === selectedIndex && "bg-accent",
              )}
              onClick={() => handleSelect(result)}
            >
              <span className="font-medium">{result.attributes.name}</span>
              {result.attributes.admin1 && (
                <span className="text-muted-foreground">, {result.attributes.admin1}</span>
              )}
              {result.attributes.country && (
                <span className="text-muted-foreground">, {result.attributes.country}</span>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

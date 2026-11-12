import { useEffect } from "react";
import { useSearchParams } from "react-router-dom";

export interface QueryParamCodec<T> {
  key: string;
  /** null = invalid; may normalize. */
  parse: (raw: string) => T | null;
  /** Canonical string form. */
  serialize: (value: T) => string;
  /** "Current period" value, evaluated per call. */
  current: () => T;
  equals: (a: T, b: T) => boolean;
}

function applyParam<T>(params: URLSearchParams, codec: QueryParamCodec<T>, value: T): URLSearchParams {
  const next = new URLSearchParams(params);
  if (codec.equals(value, codec.current())) {
    next.delete(codec.key);
  } else {
    next.set(codec.key, codec.serialize(value));
  }
  return next;
}

export function useQueryParamState<T>(codec: QueryParamCodec<T>): [T, (next: T) => void] {
  const [searchParams, setSearchParams] = useSearchParams();
  const raw = searchParams.get(codec.key);
  const parsed = raw != null ? codec.parse(raw) : null;
  const value = parsed ?? codec.current();

  // Not memoized: the codec may change every render and must not be captured stale.
  const setValue = (next: T) => {
    setSearchParams((prev) => applyParam(prev, codec, next), { replace: true });
  };

  useEffect(() => {
    if (raw == null) return;
    if (parsed == null || codec.serialize(value) !== raw || codec.equals(value, codec.current())) {
      setSearchParams((prev) => applyParam(prev, codec, value), { replace: true });
    }
    // Keyed on raw only: codec identity changes every render by design.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [raw]);

  return [value, setValue];
}

export function withParam<T>(search: URLSearchParams, codec: QueryParamCodec<T>, value: T): string {
  const s = applyParam(search, codec, value).toString();
  return s ? `?${s}` : "";
}

import { transformError, type AppError } from "./errors";

const BASE_URL = "/api/v1";

export interface ProgressInfo {
  loaded: number;
  total: number;
  percentage: number;
}

export interface RequestOptions {
  signal?: AbortSignal;
  maxRetries?: number;
  retryDelay?: number;
  skipDeduplication?: boolean;
  timeout?: number;
  skipTenantHeaders?: boolean;
  headers?: Record<string, string>;
  onProgress?: (progress: ProgressInfo) => void;
  cache?: { ttl: number; staleWhileRevalidate?: boolean };
  staleWhileRevalidate?: boolean;
}

interface CacheEntry<T = unknown> {
  value: T;
  expiresAt: number;
  staleWhileRevalidate: boolean;
}

const DEFAULT_MAX_RETRIES = 3;
const DEFAULT_RETRY_DELAY = 1000;
const MAX_RETRY_DELAY = 10000;

const SHORT_TTL = 30_000;
const LONG_TTL = 300_000;

function isRetryableStatus(status: number): boolean {
  return status >= 500 || status === 429;
}

/** Returns cache options with a short TTL (default 30s). */
export function shortLived(ms: number = SHORT_TTL): RequestOptions["cache"] {
  return { ttl: ms };
}

/** Returns cache options with a long TTL (default 5min). */
export function longLived(ms: number = LONG_TTL): RequestOptions["cache"] {
  return { ttl: ms };
}

/** Returns a function that prepends a prefix to cache keys via path. */
export function withPrefix(prefix: string) {
  return (path: string) => `${prefix}${path}`;
}

class ApiClient {
  private baseUrl: string;
  private tenantId: string | null = null;
  private householdId: string | null = null;
  private pendingRequests = new Map<string, Promise<unknown>>();
  private cache = new Map<string, CacheEntry>();
  private refreshPromise: Promise<boolean> | null = null;
  private isRedirecting = false;

  onAuthFailure: (() => void) | null = null;

  constructor(baseUrl: string = BASE_URL) {
    this.baseUrl = baseUrl;
  }

  resetAuthState() {
    this.refreshPromise = null;
    this.isRedirecting = false;
  }

  private isRefreshPath(path: string): boolean {
    return path === "/auth/token/refresh";
  }

  private async attemptRefresh(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/auth/token/refresh`, {
        method: "POST",
        credentials: "include",
        headers: { Accept: "application/vnd.api+json" },
      });
      return response.ok;
    } catch {
      console.warn("Token refresh failed due to network error");
      return false;
    }
  }

  private async handleUnauthorized<T>(retryFn: () => Promise<T>): Promise<T> {
    if (this.isRedirecting) {
      throw transformError(new ApiRequestError("Session expired", 401));
    }

    if (!this.refreshPromise) {
      this.refreshPromise = this.attemptRefresh().finally(() => {
        this.refreshPromise = null;
      });
    }

    const refreshed = await this.refreshPromise;

    if (refreshed) {
      return retryFn();
    }

    this.isRedirecting = true;
    this.onAuthFailure?.();
    if (window.location.pathname !== "/login") {
      window.location.href = "/login";
    }
    throw transformError(new ApiRequestError("Session expired", 401));
  }

  setTenant(tenant: { id: string }) {
    this.tenantId = tenant.id;
  }

  setHousehold(household: { id: string }) {
    this.householdId = household.id;
  }

  clearTenant() {
    this.tenantId = null;
    this.householdId = null;
  }

  private buildHeaders(contentType?: string, options?: RequestOptions): Record<string, string> {
    const headers: Record<string, string> = { Accept: "application/vnd.api+json" };
    if (contentType) headers["Content-Type"] = contentType;
    if (this.tenantId && !options?.skipTenantHeaders) {
      headers["X-Tenant-ID"] = this.tenantId;
    }
    if (this.householdId && !options?.skipTenantHeaders) {
      headers["X-Household-ID"] = this.householdId;
    }
    if (options?.headers) {
      Object.assign(headers, options.headers);
    }
    return headers;
  }

  private async fetchWithRetry(
    url: string,
    init: RequestInit,
    options: RequestOptions = {},
  ): Promise<Response> {
    const maxRetries = options.maxRetries ?? DEFAULT_MAX_RETRIES;
    const baseDelay = options.retryDelay ?? DEFAULT_RETRY_DELAY;

    let lastError: Error | undefined;

    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        let signal = options.signal;

        if (options.timeout && !signal) {
          const controller = new AbortController();
          signal = controller.signal;
          setTimeout(() => controller.abort(), options.timeout);
        }

        const response = await fetch(url, {
          ...init,
          ...(signal != null ? { signal } : {}),
        });

        if (!response.ok && isRetryableStatus(response.status) && attempt < maxRetries) {
          const delay = Math.min(baseDelay * Math.pow(2, attempt), MAX_RETRY_DELAY);
          await new Promise((resolve) => setTimeout(resolve, delay));
          continue;
        }

        return response;
      } catch (error) {
        if (options.signal?.aborted) {
          throw error;
        }
        lastError = error instanceof Error ? error : new Error(String(error));
        if (attempt < maxRetries) {
          const delay = Math.min(baseDelay * Math.pow(2, attempt), MAX_RETRY_DELAY);
          await new Promise((resolve) => setTimeout(resolve, delay));
        }
      }
    }

    throw lastError ?? new Error("Request failed after retries");
  }

  private deduplicatedRequest<T>(key: string, fetcher: () => Promise<T>): Promise<T> {
    const existing = this.pendingRequests.get(key);
    if (existing) return existing as Promise<T>;

    const promise = fetcher().finally(() => {
      this.pendingRequests.delete(key);
    });

    this.pendingRequests.set(key, promise);
    return promise;
  }

  private getCached<T>(key: string, options?: RequestOptions): { value: T; stale: boolean } | null {
    const entry = this.cache.get(key);
    if (!entry) return null;

    const now = Date.now();
    if (now < entry.expiresAt) {
      return { value: entry.value as T, stale: false };
    }

    if (entry.staleWhileRevalidate || options?.staleWhileRevalidate) {
      return { value: entry.value as T, stale: true };
    }

    this.cache.delete(key);
    return null;
  }

  private setCache<T>(key: string, value: T, options?: RequestOptions): void {
    if (!options?.cache) return;
    this.cache.set(key, {
      value,
      expiresAt: Date.now() + options.cache.ttl,
      staleWhileRevalidate: options.cache.staleWhileRevalidate ?? options.staleWhileRevalidate ?? false,
    });
  }

  async get<T>(path: string, options?: RequestOptions): Promise<T> {
    const url = `${this.baseUrl}${path}`;
    const dedupeKey = `GET:${url}:${this.tenantId ?? ""}`;

    if (options?.cache) {
      const cached = this.getCached<T>(dedupeKey, options);
      if (cached && !cached.stale) return cached.value;

      if (cached?.stale) {
        // Return stale value and revalidate in background
        this.revalidate<T>(url, dedupeKey, options);
        return cached.value;
      }
    }

    const fetcher = async (): Promise<T> => {
      const response = await this.fetchWithRetry(
        url,
        { credentials: "include", headers: this.buildHeaders(undefined, options) },
        options,
      );
      if (response.status === 401 && !this.isRefreshPath(path)) {
        return this.handleUnauthorized<T>(() => fetcher());
      }
      if (!response.ok) throw await this.handleError(response);
      const result = await response.json() as T;
      this.setCache(dedupeKey, result, options);
      return result;
    };

    if (options?.skipDeduplication) return fetcher();
    return this.deduplicatedRequest(dedupeKey, fetcher);
  }

  async getText(path: string, options?: RequestOptions): Promise<string> {
    const url = `${this.baseUrl}${path}`;
    const fetcher = async (): Promise<string> => {
      const response = await this.fetchWithRetry(
        url,
        { credentials: "include", headers: this.buildHeaders(undefined, options) },
        options,
      );
      if (response.status === 401 && !this.isRefreshPath(path)) {
        return this.handleUnauthorized<string>(() => fetcher());
      }
      if (!response.ok) throw await this.handleError(response);
      return response.text();
    };
    if (options?.skipDeduplication) return fetcher();
    return this.deduplicatedRequest(`GET:${url}:${this.tenantId ?? ""}`, fetcher);
  }

  async getList<T>(path: string, options?: RequestOptions): Promise<T[]> {
    return this.get<T[]>(path, options);
  }

  async getOne<T>(path: string, options?: RequestOptions): Promise<T> {
    return this.get<T>(path, options);
  }

  async post<T = void>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
    const url = `${this.baseUrl}${path}`;
    const dedupeKey = `POST:${url}:${this.tenantId ?? ""}:${JSON.stringify(body ?? null)}`;

    const fetcher = async (): Promise<T> => {
      const hasBody = body !== undefined && body !== null;
      const response = await this.fetchWithRetry(
        url,
        {
          method: "POST",
          credentials: "include",
          headers: this.buildHeaders(
            hasBody ? "application/vnd.api+json" : undefined,
            options,
          ),
          ...(hasBody ? { body: JSON.stringify(body) } : {}),
        },
        { ...options, maxRetries: options?.maxRetries ?? 0 },
      );
      if (response.status === 401 && !this.isRefreshPath(path)) {
        return this.handleUnauthorized<T>(() => fetcher());
      }
      if (!response.ok) throw await this.handleError(response);

      const contentLength = response.headers?.get?.("content-length");
      if (response.status === 204 || contentLength === "0") {
        return undefined as unknown as T;
      }

      const text = await response.text();
      if (!text) return undefined as unknown as T;
      return JSON.parse(text) as T;
    };

    if (options?.skipDeduplication) return fetcher();
    return this.deduplicatedRequest(dedupeKey, fetcher);
  }

  async put<T>(path: string, body: unknown, options?: RequestOptions): Promise<T> {
    const doRequest = async (): Promise<T> => {
      const response = await this.fetchWithRetry(
        `${this.baseUrl}${path}`,
        {
          method: "PUT",
          credentials: "include",
          headers: this.buildHeaders("application/vnd.api+json", options),
          body: JSON.stringify(body),
        },
        { ...options, maxRetries: options?.maxRetries ?? 0 },
      );
      if (response.status === 401) {
        return this.handleUnauthorized<T>(() => doRequest());
      }
      if (!response.ok) throw await this.handleError(response);
      return response.json() as Promise<T>;
    };
    return doRequest();
  }

  async patch<T>(path: string, body: unknown, options?: RequestOptions): Promise<T> {
    const doRequest = async (): Promise<T> => {
      const response = await this.fetchWithRetry(
        `${this.baseUrl}${path}`,
        {
          method: "PATCH",
          credentials: "include",
          headers: this.buildHeaders("application/vnd.api+json", options),
          body: JSON.stringify(body),
        },
        { ...options, maxRetries: options?.maxRetries ?? 0 },
      );
      if (response.status === 401) {
        return this.handleUnauthorized<T>(() => doRequest());
      }
      if (!response.ok) throw await this.handleError(response);
      return response.json() as Promise<T>;
    };
    return doRequest();
  }

  async delete(path: string, options?: RequestOptions): Promise<void> {
    const doRequest = async (): Promise<void> => {
      const response = await this.fetchWithRetry(
        `${this.baseUrl}${path}`,
        {
          method: "DELETE",
          credentials: "include",
          headers: this.buildHeaders(undefined, options),
        },
        { ...options, maxRetries: options?.maxRetries ?? 0 },
      );
      if (response.status === 401) {
        return this.handleUnauthorized<void>(() => doRequest());
      }
      if (!response.ok) throw await this.handleError(response);
    };
    return doRequest();
  }

  async upload<T>(
    path: string,
    formData: FormData,
    options?: RequestOptions,
  ): Promise<T> {
    const url = `${this.baseUrl}${path}`;

    if (options?.onProgress) {
      return this.uploadWithProgress<T>(path, url, formData, options);
    }

    const doRequest = async (): Promise<T> => {
      const headers = this.buildHeaders(undefined, options);
      // Do not set Content-Type for FormData; the browser sets it with the boundary
      const response = await this.fetchWithRetry(
        url,
        {
          method: "POST",
          credentials: "include",
          headers,
          body: formData,
        },
        { ...options, maxRetries: options?.maxRetries ?? 0 },
      );
      if (response.status === 401) {
        return this.handleUnauthorized<T>(() => doRequest());
      }
      if (!response.ok) throw await this.handleError(response);
      return response.json() as Promise<T>;
    };
    return doRequest();
  }

  async download(path: string, options?: RequestOptions): Promise<Blob> {
    const url = `${this.baseUrl}${path}`;
    const doRequest = async (): Promise<Blob> => {
      const response = await this.fetchWithRetry(
        url,
        {
          credentials: "include",
          headers: {
            ...this.buildHeaders(undefined, options),
            Accept: "*/*",
          },
        },
        options,
      );
      if (response.status === 401) {
        return this.handleUnauthorized<Blob>(() => doRequest());
      }
      if (!response.ok) throw await this.handleError(response);
      return response.blob();
    };
    return doRequest();
  }

  clearPendingRequests() {
    this.pendingRequests.clear();
  }

  clearCache() {
    this.cache.clear();
  }

  private async revalidate<T>(url: string, cacheKey: string, options?: RequestOptions): Promise<void> {
    try {
      const response = await this.fetchWithRetry(
        url,
        { credentials: "include", headers: this.buildHeaders(undefined, options) },
        { ...options, skipDeduplication: true },
      );
      if (response.ok) {
        const result = await response.json() as T;
        this.setCache(cacheKey, result, options);
      }
    } catch {
      // Revalidation failures are silent; stale data continues to be served
    }
  }

  private uploadWithProgress<T>(
    path: string,
    url: string,
    formData: FormData,
    options: RequestOptions,
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      xhr.open("POST", url, true);
      xhr.withCredentials = true;

      const headers = this.buildHeaders(undefined, options);
      for (const [key, value] of Object.entries(headers)) {
        xhr.setRequestHeader(key, value);
      }

      if (options.onProgress) {
        const onProgress = options.onProgress;
        xhr.upload.addEventListener("progress", (event) => {
          if (event.lengthComputable) {
            onProgress({
              loaded: event.loaded,
              total: event.total,
              percentage: Math.round((event.loaded / event.total) * 100),
            });
          }
        });
      }

      xhr.addEventListener("load", async () => {
        if (xhr.status === 401) {
          try {
            const result = await this.handleUnauthorized<T>(() =>
              this.upload<T>(path, formData, options),
            );
            resolve(result);
          } catch (e) {
            reject(e);
          }
          return;
        }
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            resolve(JSON.parse(xhr.responseText) as T);
          } catch {
            resolve(undefined as unknown as T);
          }
        } else {
          reject(
            transformError(
              new ApiRequestError(
                xhr.statusText || `Upload failed with status ${xhr.status}`,
                xhr.status,
              ),
            ),
          );
        }
      });

      xhr.addEventListener("error", () => {
        reject(transformError(new Error("Upload failed due to network error")));
      });

      if (options.signal) {
        options.signal.addEventListener("abort", () => xhr.abort());
      }

      if (options.timeout) {
        xhr.timeout = options.timeout;
        xhr.addEventListener("timeout", () => {
          reject(transformError(new Error("Upload timed out")));
        });
      }

      xhr.send(formData);
    });
  }

  private async handleError(response: Response): Promise<AppError> {
    try {
      const body = await response.json();
      const detail = body.errors?.[0]?.detail || body.errors?.[0]?.title || "Request failed";
      return transformError(new ApiRequestError(detail, response.status));
    } catch {
      return transformError(
        new ApiRequestError(`Request failed with status ${response.status}`, response.status),
      );
    }
  }
}

export class ApiRequestError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiRequestError";
    this.status = status;
  }
}

export const api = new ApiClient();

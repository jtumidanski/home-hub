const BASE_URL = "/api/v1";

interface RequestOptions {
  signal?: AbortSignal;
  maxRetries?: number;
  retryDelay?: number;
  skipDeduplication?: boolean;
}

const DEFAULT_MAX_RETRIES = 3;
const DEFAULT_RETRY_DELAY = 1000;
const MAX_RETRY_DELAY = 10000;

function isRetryableStatus(status: number): boolean {
  return status >= 500 || status === 429;
}

class ApiClient {
  private baseUrl: string;
  private tenantId: string | null = null;
  private pendingRequests = new Map<string, Promise<unknown>>();

  constructor(baseUrl: string = BASE_URL) {
    this.baseUrl = baseUrl;
  }

  setTenant(tenant: { id: string }) {
    this.tenantId = tenant.id;
  }

  clearTenant() {
    this.tenantId = null;
  }

  private buildHeaders(contentType?: string): Record<string, string> {
    const headers: Record<string, string> = { Accept: "application/vnd.api+json" };
    if (contentType) headers["Content-Type"] = contentType;
    if (this.tenantId) headers["X-Tenant-ID"] = this.tenantId;
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
        const response = await fetch(url, {
          ...init,
          ...(options.signal != null ? { signal: options.signal } : {}),
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

  private deduplicatedGet<T>(key: string, fetcher: () => Promise<T>): Promise<T> {
    const existing = this.pendingRequests.get(key);
    if (existing) return existing as Promise<T>;

    const promise = fetcher().finally(() => {
      this.pendingRequests.delete(key);
    });

    this.pendingRequests.set(key, promise);
    return promise;
  }

  async get<T>(path: string, options?: RequestOptions): Promise<T> {
    const url = `${this.baseUrl}${path}`;
    const dedupeKey = `GET:${url}:${this.tenantId ?? ""}`;

    const fetcher = async () => {
      const response = await this.fetchWithRetry(
        url,
        { credentials: "include", headers: this.buildHeaders() },
        options,
      );
      if (!response.ok) throw await this.handleError(response);
      return response.json() as Promise<T>;
    };

    if (options?.skipDeduplication) return fetcher();
    return this.deduplicatedGet(dedupeKey, fetcher);
  }

  async post<T>(path: string, body: unknown, options?: RequestOptions): Promise<T> {
    const response = await this.fetchWithRetry(
      `${this.baseUrl}${path}`,
      {
        method: "POST",
        credentials: "include",
        headers: this.buildHeaders("application/vnd.api+json"),
        body: JSON.stringify(body),
      },
      { ...options, maxRetries: 0 },
    );
    if (!response.ok) throw await this.handleError(response);
    return response.json() as Promise<T>;
  }

  async patch<T>(path: string, body: unknown, options?: RequestOptions): Promise<T> {
    const response = await this.fetchWithRetry(
      `${this.baseUrl}${path}`,
      {
        method: "PATCH",
        credentials: "include",
        headers: this.buildHeaders("application/vnd.api+json"),
        body: JSON.stringify(body),
      },
      { ...options, maxRetries: 0 },
    );
    if (!response.ok) throw await this.handleError(response);
    return response.json() as Promise<T>;
  }

  async delete(path: string, options?: RequestOptions): Promise<void> {
    const response = await this.fetchWithRetry(
      `${this.baseUrl}${path}`,
      {
        method: "DELETE",
        credentials: "include",
        headers: this.buildHeaders(),
      },
      { ...options, maxRetries: 0 },
    );
    if (!response.ok) throw await this.handleError(response);
  }

  async postNoContent(path: string, options?: RequestOptions): Promise<void> {
    const response = await this.fetchWithRetry(
      `${this.baseUrl}${path}`,
      {
        method: "POST",
        credentials: "include",
        headers: this.buildHeaders(),
      },
      { ...options, maxRetries: 0 },
    );
    if (!response.ok) throw await this.handleError(response);
  }

  clearPendingRequests() {
    this.pendingRequests.clear();
  }

  private async handleError(response: Response): Promise<ApiRequestError> {
    try {
      const body = await response.json();
      const detail = body.errors?.[0]?.detail || body.errors?.[0]?.title || "Request failed";
      return new ApiRequestError(detail, response.status);
    } catch {
      return new ApiRequestError(`Request failed with status ${response.status}`, response.status);
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

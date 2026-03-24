const BASE_URL = "/api/v1";

class ApiClient {
  private baseUrl: string;
  private tenantId: string | null = null;

  constructor(baseUrl: string = BASE_URL) {
    this.baseUrl = baseUrl;
  }

  setTenant(tenantId: string) {
    this.tenantId = tenantId;
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

  async get<T>(path: string): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      credentials: "include",
      headers: this.buildHeaders(),
    });
    if (!response.ok) {
      throw await this.handleError(response);
    }
    return response.json();
  }

  async post<T>(path: string, body: unknown): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      credentials: "include",
      headers: this.buildHeaders("application/vnd.api+json"),
      body: JSON.stringify(body),
    });
    if (!response.ok) {
      throw await this.handleError(response);
    }
    return response.json();
  }

  async patch<T>(path: string, body: unknown): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: "PATCH",
      credentials: "include",
      headers: this.buildHeaders("application/vnd.api+json"),
      body: JSON.stringify(body),
    });
    if (!response.ok) {
      throw await this.handleError(response);
    }
    return response.json();
  }

  async delete(path: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: "DELETE",
      credentials: "include",
      headers: this.buildHeaders(),
    });
    if (!response.ok) {
      throw await this.handleError(response);
    }
  }

  async postNoContent(path: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      credentials: "include",
      headers: this.buildHeaders(),
    });
    if (!response.ok) {
      throw await this.handleError(response);
    }
  }

  private async handleError(response: Response): Promise<Error> {
    try {
      const body = await response.json();
      const detail = body.errors?.[0]?.detail || body.errors?.[0]?.title || "Request failed";
      return new Error(detail);
    } catch {
      return new Error(`Request failed with status ${response.status}`);
    }
  }
}

export const api = new ApiClient();

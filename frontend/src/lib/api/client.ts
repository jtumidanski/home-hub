const BASE_URL = "/api/v1";

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = BASE_URL) {
    this.baseUrl = baseUrl;
  }

  async get<T>(path: string): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      credentials: "include",
      headers: { Accept: "application/vnd.api+json" },
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
      headers: {
        "Content-Type": "application/vnd.api+json",
        Accept: "application/vnd.api+json",
      },
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
      headers: {
        "Content-Type": "application/vnd.api+json",
        Accept: "application/vnd.api+json",
      },
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
    });
    if (!response.ok) {
      throw await this.handleError(response);
    }
  }

  async postNoContent(path: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      credentials: "include",
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

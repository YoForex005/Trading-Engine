/**
 * API Client - Shared utilities for HTTP requests
 * Eliminates duplicated fetch/error handling patterns
 */

export class ApiError extends Error {
  status: number;
  endpoint?: string;

  constructor(
    status: number,
    message: string,
    endpoint?: string
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.endpoint = endpoint;
  }
}

type RequestOptions = {
  headers?: Record<string, string>;
  signal?: AbortSignal;
};

class ApiClient {
  private baseURL: string;

  constructor(baseURL = '') {
    this.baseURL = baseURL;
  }

  private async handleResponse<T>(response: Response, endpoint: string): Promise<T> {
    if (!response.ok) {
      const errorText = await response.text().catch(() => response.statusText);
      throw new ApiError(
        response.status,
        `${response.status} ${response.statusText}: ${errorText}`,
        endpoint
      );
    }

    // Handle empty responses
    const text = await response.text();
    if (!text) {
      return {} as T;
    }

    try {
      return JSON.parse(text) as T;
    } catch (error) {
      throw new ApiError(
        500,
        `Failed to parse JSON response: ${error}`,
        endpoint
      );
    }
  }

  async get<T>(endpoint: string, options?: RequestOptions): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      signal: options?.signal,
    });
    return this.handleResponse<T>(response, endpoint);
  }

  async post<T>(
    endpoint: string,
    data?: unknown,
    options?: RequestOptions
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      body: data ? JSON.stringify(data) : undefined,
      signal: options?.signal,
    });
    return this.handleResponse<T>(response, endpoint);
  }

  async put<T>(
    endpoint: string,
    data?: unknown,
    options?: RequestOptions
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const response = await fetch(url, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      body: data ? JSON.stringify(data) : undefined,
      signal: options?.signal,
    });
    return this.handleResponse<T>(response, endpoint);
  }

  async delete(endpoint: string, options?: RequestOptions): Promise<void> {
    const url = `${this.baseURL}${endpoint}`;
    const response = await fetch(url, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      signal: options?.signal,
    });

    if (!response.ok) {
      const errorText = await response.text().catch(() => response.statusText);
      throw new ApiError(
        response.status,
        `${response.status} ${response.statusText}: ${errorText}`,
        endpoint
      );
    }
  }
}

// Export singleton instance
export const api = new ApiClient();

// Export for testing with custom base URL
export { ApiClient };

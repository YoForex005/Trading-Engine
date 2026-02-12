/**
 * Authenticated API Client
 * Centralized fetch wrapper with automatic JWT token injection,
 * CORS credentials, error handling, and 401 redirection
 */

import { API_CONFIG } from '@/config/api';

export interface ApiClientOptions extends RequestInit {
    skipAuth?: boolean;
    skipContentType?: boolean;
}

export class ApiError extends Error {
    constructor(
        public status: number,
        public statusText: string,
        public data?: any
    ) {
        super(`API Error ${status}: ${statusText}`);
        this.name = 'ApiError';
    }
}

/**
 * Get JWT token from localStorage
 * Checks multiple storage keys for compatibility
 */
function getAuthToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('admin_token') ||
           localStorage.getItem('rtx_token') ||
           localStorage.getItem('jwt_token') ||
           null;
}

/**
 * Handle 401 Unauthorized responses
 * Redirects to login page and clears tokens
 */
function handle401Unauthorized(): void {
    if (typeof window === 'undefined') return;

    // Clear all auth tokens
    localStorage.removeItem('admin_token');
    localStorage.removeItem('rtx_token');
    localStorage.removeItem('jwt_token');

    // Redirect to login page
    // TODO: Replace with actual login route when implemented
    console.error('[API] 401 Unauthorized - Token expired or invalid');

    // For now, just log - can implement redirect later
    // window.location.href = '/login';
}

/**
 * Authenticated fetch wrapper
 * Automatically adds Authorization header, CORS credentials, and error handling
 *
 * @param url - API endpoint URL
 * @param options - Fetch options with optional skipAuth and skipContentType flags
 * @returns Promise with parsed JSON response
 * @throws ApiError on HTTP errors
 */
export async function apiClient<T = any>(
    url: string,
    options: ApiClientOptions = {}
): Promise<T> {
    const {
        skipAuth = false,
        skipContentType = false,
        headers = {},
        ...fetchOptions
    } = options;

    // Build headers
    const requestHeaders: Record<string, string> = { ...headers as Record<string, string> };

    // Add Authorization header if not skipped
    if (!skipAuth) {
        const token = getAuthToken();
        if (token) {
            requestHeaders['Authorization'] = `Bearer ${token}`;
        }
    }

    // Add Content-Type for JSON requests
    if (!skipContentType && !requestHeaders['Content-Type']) {
        requestHeaders['Content-Type'] = 'application/json';
    }

    // Merge fetch options with security defaults
    const finalOptions: RequestInit = {
        ...fetchOptions,
        headers: requestHeaders,
        credentials: 'include', // CRITICAL: Include cookies/credentials for CORS
        mode: 'cors', // Explicit CORS mode
    };

    try {
        const response = await fetch(url, finalOptions);

        // Handle 401 Unauthorized
        if (response.status === 401) {
            handle401Unauthorized();
            throw new ApiError(401, 'Unauthorized', await response.text());
        }

        // Handle non-OK responses
        if (!response.ok) {
            let errorData;
            try {
                errorData = await response.json();
            } catch {
                errorData = await response.text();
            }
            throw new ApiError(response.status, response.statusText, errorData);
        }

        // Parse response
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            return await response.json();
        }

        // Return text for non-JSON responses
        return await response.text() as any;
    } catch (error) {
        // Re-throw ApiError instances
        if (error instanceof ApiError) {
            throw error;
        }

        // Wrap fetch errors
        if (error instanceof Error) {
            console.error('[API] Request failed:', error.message);
            throw new ApiError(0, 'Network Error', error.message);
        }

        throw error;
    }
}

/**
 * Convenience methods for common HTTP verbs
 */
export const api = {
    /**
     * GET request
     */
    get: <T = any>(url: string, options?: ApiClientOptions): Promise<T> => {
        return apiClient<T>(url, { ...options, method: 'GET' });
    },

    /**
     * POST request
     */
    post: <T = any>(url: string, data?: any, options?: ApiClientOptions): Promise<T> => {
        return apiClient<T>(url, {
            ...options,
            method: 'POST',
            body: data ? JSON.stringify(data) : undefined,
        });
    },

    /**
     * PUT request
     */
    put: <T = any>(url: string, data?: any, options?: ApiClientOptions): Promise<T> => {
        return apiClient<T>(url, {
            ...options,
            method: 'PUT',
            body: data ? JSON.stringify(data) : undefined,
        });
    },

    /**
     * PATCH request
     */
    patch: <T = any>(url: string, data?: any, options?: ApiClientOptions): Promise<T> => {
        return apiClient<T>(url, {
            ...options,
            method: 'PATCH',
            body: data ? JSON.stringify(data) : undefined,
        });
    },

    /**
     * DELETE request
     */
    delete: <T = any>(url: string, options?: ApiClientOptions): Promise<T> => {
        return apiClient<T>(url, { ...options, method: 'DELETE' });
    },
};

/**
 * Store JWT token in localStorage
 */
export function setAuthToken(token: string): void {
    if (typeof window === 'undefined') return;
    localStorage.setItem('admin_token', token);
    localStorage.setItem('rtx_token', token); // Compatibility
}

/**
 * Clear JWT token from localStorage
 */
export function clearAuthToken(): void {
    if (typeof window === 'undefined') return;
    localStorage.removeItem('admin_token');
    localStorage.removeItem('rtx_token');
    localStorage.removeItem('jwt_token');
}

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
    return getAuthToken() !== null;
}

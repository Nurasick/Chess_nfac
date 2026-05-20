/**
 * API Fetch Wrapper
 * Centralized HTTP client with auth token handling
 */

/**
 * Get the API base URL from environment
 */
function getBaseUrl(): string {
  const url = import.meta.env.VITE_API_URL;
  if (!url) {
    throw new Error('VITE_API_URL environment variable is not set');
  }
  return url;
}

/**
 * Get the stored access token from localStorage
 */
function getAccessToken(): string | null {
  return localStorage.getItem('access_token');
}

/**
 * Store the access token in localStorage
 */
export function setAccessToken(token: string): void {
  localStorage.setItem('access_token', token);
}

/**
 * Remove the access token from localStorage
 */
export function clearAccessToken(): void {
  localStorage.removeItem('access_token');
}

/**
 * API response with data envelope
 */
export interface ApiError {
  message: string;
  code?: string;
}

/**
 * Main fetch wrapper with auth header and error handling
 */
export async function apiFetch<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const url = `${getBaseUrl()}${path}`;
  const token = getAccessToken();

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options?.headers as Record<string, string>) ?? {}),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (!response.ok) {
    let errorMessage = 'An error occurred';
    let errorData: unknown = null;

    try {
      errorData = await response.json();
      if (typeof errorData === 'object' && errorData !== null) {
        const data = errorData as Record<string, unknown>;
        if (typeof data.message === 'string') {
          errorMessage = data.message;
        } else if (typeof data.error === 'string') {
          errorMessage = data.error;
        }
      }
    } catch {
      // Response is not JSON, use status text
      errorMessage = response.statusText || 'An error occurred';
    }

    const error = new Error(errorMessage) as Error & { status: number; data?: unknown };
    error.status = response.status;
    error.data = errorData;
    throw error;
  }

  const data = await response.json();
  return data as T;
}

/**
 * Typed GET request
 */
export async function get<T>(path: string, options?: RequestInit): Promise<T> {
  return apiFetch<T>(path, {
    ...options,
    method: 'GET',
  });
}

/**
 * Typed POST request
 */
export async function post<T>(
  path: string,
  body?: unknown,
  options?: RequestInit
): Promise<T> {
  return apiFetch<T>(path, {
    ...options,
    method: 'POST',
    body: body ? JSON.stringify(body) : undefined,
  });
}

/**
 * Typed PUT request
 */
export async function put<T>(
  path: string,
  body?: unknown,
  options?: RequestInit
): Promise<T> {
  return apiFetch<T>(path, {
    ...options,
    method: 'PUT',
    body: body ? JSON.stringify(body) : undefined,
  });
}

/**
 * Typed PATCH request
 */
export async function patch<T>(
  path: string,
  body?: unknown,
  options?: RequestInit
): Promise<T> {
  return apiFetch<T>(path, {
    ...options,
    method: 'PATCH',
    body: body ? JSON.stringify(body) : undefined,
  });
}

/**
 * Typed DELETE request
 */
export async function del<T>(path: string, options?: RequestInit): Promise<T> {
  return apiFetch<T>(path, {
    ...options,
    method: 'DELETE',
  });
}

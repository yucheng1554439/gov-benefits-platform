import type { ApiError } from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export class ApiClientError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message);
    this.name = 'ApiClientError';
  }
}

function getStoredToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('access_token');
}

function getStoredRefreshToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('refresh_token');
}

function getStoredAgencyId(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('agency_id');
}

export function setApiCredentials(accessToken: string, agencyId: string) {
  if (typeof window === 'undefined') return;
  localStorage.setItem('access_token', accessToken);
  localStorage.setItem('agency_id', agencyId);
}

export function clearApiCredentials() {
  if (typeof window === 'undefined') return;
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('agency_id');
  localStorage.removeItem('auth_user');
}

export interface FetchOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
  skipAuth?: boolean;
  agencyId?: string;
  _retried?: boolean;
}

function networkErrorMessage(path: string, cause: unknown): string {
  if (cause instanceof TypeError) {
    return `Unable to reach the API at ${API_BASE}${path}. Ensure the backend is running and accessible.`;
  }
  if (cause instanceof Error && cause.message) {
    return cause.message;
  }
  return 'An unexpected network error occurred.';
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getStoredRefreshToken();
  if (!refreshToken) return null;

  try {
    const response = await fetch(`${API_BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (!response.ok) return null;
    const tokens = (await response.json()) as {
      access_token: string;
      refresh_token: string;
      user?: { agency_id?: string };
    };
    localStorage.setItem('access_token', tokens.access_token);
    localStorage.setItem('refresh_token', tokens.refresh_token);
    if (tokens.user?.agency_id) {
      localStorage.setItem('agency_id', tokens.user.agency_id);
    }
    return tokens.access_token;
  } catch {
    return null;
  }
}

export async function apiFetch<T>(path: string, options: FetchOptions = {}): Promise<T> {
  const { body, skipAuth, agencyId, headers: customHeaders, _retried, ...rest } = options;

  const headers: Record<string, string> = {
    ...(customHeaders as Record<string, string>),
  };

  if (body !== undefined && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  if (!skipAuth) {
    const token = getStoredToken();
    if (!token) {
      throw new ApiClientError('Your session has expired. Please sign in again.', 401);
    }
    headers['Authorization'] = `Bearer ${token}`;
    const agency = agencyId ?? getStoredAgencyId();
    if (agency) {
      headers['X-Agency-ID'] = agency;
    }
  }

  let response: Response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      ...rest,
      headers,
      body: body instanceof FormData ? body : body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch (cause) {
    throw new ApiClientError(networkErrorMessage(path, cause), 0);
  }

  if (response.status === 401 && !skipAuth && !_retried) {
    const newToken = await refreshAccessToken();
    if (newToken) {
      return apiFetch<T>(path, { ...options, _retried: true });
    }
    clearApiCredentials();
    throw new ApiClientError('Your session has expired. Please sign in again.', 401);
  }

  if (!response.ok) {
    let message = `Request failed (${response.status})`;
    try {
      const err = (await response.json()) as ApiError;
      if (err.error) message = err.error;
    } catch {
      if (response.status === 401) message = 'Your session has expired. Please sign in again.';
      else if (response.status === 403) message = 'You do not have permission to perform this action.';
      else if (response.status >= 500) message = 'The server encountered an error. Please try again later.';
    }
    throw new ApiClientError(message, response.status);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string, options?: FetchOptions) => apiFetch<T>(path, { ...options, method: 'GET' }),
  post: <T>(path: string, body?: unknown, options?: FetchOptions) =>
    apiFetch<T>(path, { ...options, method: 'POST', body }),
  put: <T>(path: string, body?: unknown, options?: FetchOptions) =>
    apiFetch<T>(path, { ...options, method: 'PUT', body }),
  patch: <T>(path: string, body?: unknown, options?: FetchOptions) =>
    apiFetch<T>(path, { ...options, method: 'PATCH', body }),
};

export { API_BASE };

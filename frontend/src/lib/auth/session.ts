'use client';

import type { AuthUser, TokenPair } from '@/lib/api/types';
import { clearApiCredentials, setApiCredentials } from '@/lib/api/client';

const AUTH_USER_KEY = 'auth_user';
const REFRESH_TOKEN_KEY = 'refresh_token';

export function saveSession(tokens: TokenPair) {
  if (typeof window === 'undefined') return;
  setApiCredentials(tokens.access_token, tokens.user.agency_id);
  localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token);
  localStorage.setItem(AUTH_USER_KEY, JSON.stringify(tokens.user));
}

export function getSession(): AuthUser | null {
  if (typeof window === 'undefined') return null;
  const raw = localStorage.getItem(AUTH_USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AuthUser;
  } catch {
    return null;
  }
}

export function getAccessToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('access_token');
}

export function clearSession() {
  clearApiCredentials();
}

export function updateAgencyId(agencyId: string) {
  if (typeof window === 'undefined') return;
  localStorage.setItem('agency_id', agencyId);
  const user = getSession();
  if (user) {
    user.agency_id = agencyId;
    localStorage.setItem(AUTH_USER_KEY, JSON.stringify(user));
  }
}

export function getDisplayName(user: AuthUser): string {
  if (user.profile) {
    return `${user.profile.first_name} ${user.profile.last_name}`;
  }
  return user.user.email;
}

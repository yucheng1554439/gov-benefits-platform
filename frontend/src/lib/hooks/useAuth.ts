'use client';

import { useCallback, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api/client';
import type { AuthUser, TokenPair } from '@/lib/api/types';
import { clearSession, getSession, saveSession } from '@/lib/auth/session';
import { getDefaultRoute, resolvePrimaryRole } from '@/lib/auth/rbac';

export function useAuth() {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    setUser(getSession());
    setLoading(false);
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const tokens = await api.post<TokenPair>(
        '/auth/login',
        { email, password },
        { skipAuth: true },
      );
      saveSession(tokens);
      setUser(tokens.user);
      router.push(getDefaultRoute(resolvePrimaryRole(tokens.user)));
    },
    [router],
  );

  const register = useCallback(
    async (data: {
      email: string;
      password: string;
      first_name: string;
      last_name: string;
      phone?: string;
      agency_id?: string;
    }) => {
      const tokens = await api.post<TokenPair>('/auth/register', data, { skipAuth: true });
      saveSession(tokens);
      setUser(tokens.user);
      router.push('/citizen/dashboard');
    },
    [router],
  );

  const logout = useCallback(() => {
    clearSession();
    setUser(null);
    router.push('/login');
  }, [router]);

  const refreshUser = useCallback(async () => {
    try {
      const me = await api.get<AuthUser>('/auth/me');
      saveSession({
        access_token: localStorage.getItem('access_token') ?? '',
        refresh_token: localStorage.getItem('refresh_token') ?? '',
        expires_at: '',
        user: me,
      });
      setUser(me);
    } catch {
      clearSession();
      setUser(null);
    }
  }, []);

  return { user, loading, login, register, logout, refreshUser };
}

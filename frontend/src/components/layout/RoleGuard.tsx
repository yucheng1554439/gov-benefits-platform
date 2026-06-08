'use client';

import { useEffect } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import { useAuth } from '@/lib/hooks/useAuth';
import { canAccessRoute } from '@/lib/auth/rbac';
import { AppShell } from './AppShell';

interface RoleGuardProps {
  children: React.ReactNode;
}

export function RoleGuard({ children }: RoleGuardProps) {
  const { user, loading, logout } = useAuth();
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    if (loading) return;
    if (!user) {
      router.replace('/login');
      return;
    }
    if (!canAccessRoute(user, pathname)) {
      router.replace('/login');
    }
  }, [user, loading, pathname, router]);

  if (loading || !user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gov-surface">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-gov-navy border-t-transparent" />
      </div>
    );
  }

  if (!canAccessRoute(user, pathname)) return null;

  return <AppShell user={user} onLogout={logout}>{children}</AppShell>;
}

'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getSession } from '@/lib/auth/session';
import { getDefaultRoute, resolvePrimaryRole } from '@/lib/auth/rbac';

export default function HomePage() {
  const router = useRouter();

  useEffect(() => {
    const user = getSession();
    if (user) {
      router.replace(getDefaultRoute(resolvePrimaryRole(user)));
    } else {
      router.replace('/login');
    }
  }, [router]);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-gov-navy border-t-transparent" />
    </div>
  );
}

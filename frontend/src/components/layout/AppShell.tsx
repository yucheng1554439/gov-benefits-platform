'use client';

import type { AuthUser } from '@/lib/api/types';
import { resolvePrimaryRole } from '@/lib/auth/rbac';
import { Header } from './Header';
import { Sidebar } from './Sidebar';

interface AppShellProps {
  user: AuthUser;
  onLogout: () => void;
  children: React.ReactNode;
}

export function AppShell({ user, onLogout, children }: AppShellProps) {
  const role = resolvePrimaryRole(user);

  return (
    <div className="flex min-h-screen bg-gov-surface">
      <Sidebar role={role} />
      <div className="flex flex-1 flex-col">
        <Header user={user} onLogout={onLogout} />
        <main className="flex-1 overflow-auto p-6" id="main-content">
          {children}
        </main>
      </div>
    </div>
  );
}

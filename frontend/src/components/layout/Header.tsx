'use client';

import { getDisplayName } from '@/lib/auth/session';
import type { AuthUser } from '@/lib/api/types';
import { resolvePrimaryRole } from '@/lib/auth/rbac';
import { AgencySwitcher } from './AgencySwitcher';
import { Button } from '@/components/ui/Button';

interface HeaderProps {
  user: AuthUser;
  onLogout: () => void;
}

export function Header({ user, onLogout }: HeaderProps) {
  const role = resolvePrimaryRole(user);

  return (
    <header className="flex h-16 items-center justify-between border-b border-gov-border bg-white px-6">
      <div>
        <p className="text-sm text-gov-slate">Signed in as</p>
        <p className="font-semibold text-gov-navy">{getDisplayName(user)}</p>
      </div>
      <div className="flex items-center gap-4">
        <span className="rounded-full bg-gov-surface px-3 py-1 text-xs font-medium capitalize text-gov-navy">
          {role.replace('_', ' ')}
        </span>
        <AgencySwitcher currentAgencyId={user.agency_id} />
        <Button variant="outline" size="sm" onClick={onLogout}>
          Sign Out
        </Button>
      </div>
    </header>
  );
}

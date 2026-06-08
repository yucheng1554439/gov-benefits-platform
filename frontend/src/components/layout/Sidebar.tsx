'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import type { AppRole } from '@/lib/auth/rbac';
import { NAV_ITEMS } from '@/lib/auth/rbac';

interface SidebarProps {
  role: AppRole;
}

export function Sidebar({ role }: SidebarProps) {
  const pathname = usePathname();
  const items = NAV_ITEMS[role];

  return (
    <aside className="flex w-64 flex-col bg-gov-navy-dark text-white" aria-label="Main navigation">
      <div className="border-b border-white/10 px-6 py-5">
        <p className="text-xs font-medium uppercase tracking-widest text-gov-gold-light">Gov Benefits</p>
        <h1 className="mt-1 text-lg font-semibold">Benefits Portal</h1>
      </div>
      <nav className="flex-1 space-y-1 px-3 py-4">
        {items.map((item) => {
          const active = pathname === item.href || pathname.startsWith(item.href + '/');
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'block rounded-md px-3 py-2 text-sm font-medium transition-colors',
                active
                  ? 'bg-gov-navy text-white'
                  : 'text-slate-300 hover:bg-white/10 hover:text-white',
              )}
              aria-current={active ? 'page' : undefined}
            >
              {item.label}
            </Link>
          );
        })}
      </nav>
      <div className="border-t border-white/10 px-6 py-4 text-xs text-slate-400">
        WCAG AA Compliant
      </div>
    </aside>
  );
}

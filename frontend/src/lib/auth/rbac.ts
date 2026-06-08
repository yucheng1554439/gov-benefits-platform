import type { AuthUser } from '@/lib/api/types';

export type AppRole = 'citizen' | 'worker' | 'supervisor' | 'admin';

export function resolvePrimaryRole(user: AuthUser): AppRole {
  const roles = user.roles ?? [];
  if (roles.includes('admin')) return 'admin';
  if (roles.includes('supervisor')) return 'supervisor';
  if (roles.includes('case_worker') || user.agency_role === 'worker') return 'worker';
  return 'citizen';
}

export function getDefaultRoute(role: AppRole): string {
  switch (role) {
    case 'admin':
      return '/admin/users';
    case 'supervisor':
      return '/supervisor/escalations';
    case 'worker':
      return '/worker/queue';
    default:
      return '/citizen/dashboard';
  }
}

export function canAccessRoute(user: AuthUser | null, pathname: string): boolean {
  if (!user) return false;
  const role = resolvePrimaryRole(user);

  if (pathname.startsWith('/citizen')) return role === 'citizen';
  if (pathname.startsWith('/worker')) return role === 'worker' || role === 'admin';
  if (pathname.startsWith('/supervisor')) return role === 'supervisor' || role === 'admin';
  if (pathname.startsWith('/admin')) return role === 'admin';
  if (pathname.startsWith('/dashboard')) return role === 'supervisor' || role === 'admin';

  return true;
}

export const NAV_ITEMS: Record<AppRole, { label: string; href: string }[]> = {
  citizen: [
    { label: 'Dashboard', href: '/citizen/dashboard' },
    { label: 'Apply', href: '/citizen/apply' },
    { label: 'Notifications', href: '/citizen/notifications' },
  ],
  worker: [
    { label: 'Queue', href: '/worker/queue' },
    { label: 'Fraud Review', href: '/worker/fraud-review' },
    { label: 'Appeals', href: '/worker/appeals' },
  ],
  supervisor: [
    { label: 'Escalations', href: '/supervisor/escalations' },
    { label: 'Appeals', href: '/supervisor/appeals' },
    { label: 'Audit Trail', href: '/supervisor/audit' },
    { label: 'Analytics', href: '/dashboard/analytics' },
    { label: 'Geography', href: '/dashboard/geography' },
  ],
  admin: [
    { label: 'Users', href: '/admin/users' },
    { label: 'Appeals', href: '/supervisor/appeals' },
    { label: 'Audit Trail', href: '/admin/audit' },
    { label: 'Agencies', href: '/admin/agencies' },
    { label: 'Rules', href: '/admin/rules' },
    { label: 'Workflow', href: '/admin/workflow' },
    { label: 'Retention', href: '/admin/retention' },
    { label: 'Feature Flags', href: '/admin/feature-flags' },
    { label: 'SLA', href: '/admin/sla' },
    { label: 'Letters', href: '/admin/letters' },
    { label: 'Reports', href: '/admin/reports' },
    { label: 'Analytics', href: '/dashboard/analytics' },
    { label: 'Geography', href: '/dashboard/geography' },
  ],
};

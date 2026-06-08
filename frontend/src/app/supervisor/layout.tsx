import { RoleGuard } from '@/components/layout/RoleGuard';

export default function SupervisorLayout({ children }: { children: React.ReactNode }) {
  return <RoleGuard>{children}</RoleGuard>;
}

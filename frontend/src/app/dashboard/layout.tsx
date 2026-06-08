import { RoleGuard } from '@/components/layout/RoleGuard';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return <RoleGuard>{children}</RoleGuard>;
}

import { RoleGuard } from '@/components/layout/RoleGuard';

export default function CitizenLayout({ children }: { children: React.ReactNode }) {
  return <RoleGuard>{children}</RoleGuard>;
}

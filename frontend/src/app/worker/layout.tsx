import { RoleGuard } from '@/components/layout/RoleGuard';

export default function WorkerLayout({ children }: { children: React.ReactNode }) {
  return <RoleGuard>{children}</RoleGuard>;
}
